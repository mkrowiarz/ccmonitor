package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/michal/ccmonitor/internal/domain"
	"github.com/michal/ccmonitor/internal/format"
)

const (
	usageEndpoint = "https://api.anthropic.com/api/oauth/usage"
	usageBetaTag  = "oauth-2025-04-20"
	usageCacheTTL = 10 * time.Minute
)

// usageResponse is the JSON shape returned by the usage API.
type usageResponse struct {
	FiveHour struct {
		Utilization float64 `json:"utilization"`
		ResetsAt    string  `json:"resets_at"`
	} `json:"five_hour"`
	SevenDay struct {
		Utilization float64 `json:"utilization"`
		ResetsAt    string  `json:"resets_at"`
	} `json:"seven_day"`
}

// diskCache is the JSON shape written to disk.
type diskCache struct {
	FiveHour   *diskWindow `json:"five_hour,omitempty"`
	SevenDay   *diskWindow `json:"seven_day,omitempty"`
	FetchedAt  time.Time   `json:"fetched_at"`
	RetryAfter time.Time   `json:"retry_after,omitempty"`
	LastError  string      `json:"last_error,omitempty"`
}

type diskWindow struct {
	Utilization float64   `json:"utilization"`
	ResetsAt    time.Time `json:"resets_at"`
}

// usageClient fetches rate limit data with in-memory caching.
type usageClient struct {
	mu          sync.Mutex
	cached      *domain.RateLimits
	fetchedAt   time.Time
	lastErr     error
	retryAfter  time.Time
	ttl         time.Duration
	cachePath   string
	tokenFn     func() (string, error)
}

func newUsageClient() *usageClient {
	home, _ := os.UserHomeDir()
	u := &usageClient{
		ttl:       usageCacheTTL,
		cachePath: filepath.Join(home, ".claude", "ccmonitor-usage-cache.json"),
		tokenFn:   readOAuthToken,
	}
	u.loadDiskCache()
	return u
}

func (u *usageClient) loadDiskCache() {
	data, err := os.ReadFile(u.cachePath)
	if err != nil {
		return
	}
	var dc diskCache
	if err := json.Unmarshal(data, &dc); err != nil {
		return
	}
	rl := &domain.RateLimits{FetchedAt: dc.FetchedAt}
	if dc.FiveHour != nil {
		rl.FiveHour = &domain.RateWindow{Utilization: dc.FiveHour.Utilization, ResetsAt: dc.FiveHour.ResetsAt}
	}
	if dc.SevenDay != nil {
		rl.SevenDay = &domain.RateWindow{Utilization: dc.SevenDay.Utilization, ResetsAt: dc.SevenDay.ResetsAt}
	}
	u.cached = rl
	u.fetchedAt = dc.FetchedAt
	if !dc.RetryAfter.IsZero() && time.Now().Before(dc.RetryAfter) {
		u.lastErr = fmt.Errorf("%s", dc.LastError)
		u.retryAfter = dc.RetryAfter
	}
}

func (u *usageClient) saveDiskCache(rl *domain.RateLimits) {
	dc := diskCache{FetchedAt: rl.FetchedAt}
	if rl.FiveHour != nil {
		dc.FiveHour = &diskWindow{Utilization: rl.FiveHour.Utilization, ResetsAt: rl.FiveHour.ResetsAt}
	}
	if rl.SevenDay != nil {
		dc.SevenDay = &diskWindow{Utilization: rl.SevenDay.Utilization, ResetsAt: rl.SevenDay.ResetsAt}
	}
	data, err := json.Marshal(dc)
	if err != nil {
		return
	}
	_ = os.WriteFile(u.cachePath, data, 0644)
}

func (u *usageClient) saveRetryState(errMsg string, retryAfter time.Time) {
	// Load existing cache to preserve data, just update retry fields
	dc := diskCache{RetryAfter: retryAfter, LastError: errMsg}
	if existing, err := os.ReadFile(u.cachePath); err == nil {
		_ = json.Unmarshal(existing, &dc)
		dc.RetryAfter = retryAfter
		dc.LastError = errMsg
	}
	data, err := json.Marshal(dc)
	if err != nil {
		return
	}
	_ = os.WriteFile(u.cachePath, data, 0644)
}

// Get returns cached rate limits if fresh, otherwise fetches from API.
func (u *usageClient) Get(ctx context.Context) (*domain.RateLimits, []string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.cached != nil && time.Since(u.fetchedAt) < u.ttl {
		return u.cached, nil, nil
	}

	// Don't retry if we're within a retry-after window
	if u.lastErr != nil && time.Now().Before(u.retryAfter) {
		if u.cached != nil {
			return u.cached, nil, nil
		}
		return nil, nil, u.lastErr
	}

	result, retryAfter, err := u.fetch(ctx)
	if err != nil {
		u.lastErr = err
		if retryAfter.IsZero() {
			u.retryAfter = time.Now().Add(u.ttl)
		} else {
			u.retryAfter = retryAfter
		}
		u.saveRetryState(err.Error(), u.retryAfter)
		// Graceful degradation: return stale cache if available
		if u.cached != nil {
			return u.cached, []string{"rate limits stale: " + err.Error()}, nil
		}
		return nil, nil, err
	}
	u.lastErr = nil

	now := time.Now()
	result.FetchedAt = now
	u.cached = result
	u.fetchedAt = now
	u.saveDiskCache(result)
	return result, nil, nil
}

func (u *usageClient) fetch(ctx context.Context) (*domain.RateLimits, time.Time, error) {
	var noRetry time.Time

	token, err := u.tokenFn()
	if err != nil {
		return nil, noRetry, fmt.Errorf("oauth token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, usageEndpoint, nil)
	if err != nil {
		return nil, noRetry, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", usageBetaTag)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, noRetry, fmt.Errorf("usage API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		retryAt := parseRetryAfter(resp.Header.Get("retry-after"))
		if resp.StatusCode == http.StatusTooManyRequests {
			wait := time.Until(retryAt)
			if wait <= 0 {
				wait = u.ttl
			}
			return nil, retryAt, fmt.Errorf("API cooldown, %s left", format.FormatUptime(wait))
		}
		return nil, retryAt, fmt.Errorf("usage API %d", resp.StatusCode)
	}

	var data usageResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, noRetry, fmt.Errorf("usage API decode: %w", err)
	}

	rl, err := parseUsageResponse(&data)
	return rl, noRetry, err
}

// parseRetryAfter parses the retry-after header (seconds) into an absolute time.
func parseRetryAfter(header string) time.Time {
	if header == "" {
		return time.Time{}
	}
	if secs, err := time.ParseDuration(header + "s"); err == nil {
		return time.Now().Add(secs)
	}
	// Try as HTTP-date
	if t, err := time.Parse(time.RFC1123, header); err == nil {
		return t
	}
	return time.Time{}
}

func parseUsageResponse(data *usageResponse) (*domain.RateLimits, error) {
	rl := &domain.RateLimits{}

	if t, err := time.Parse(time.RFC3339Nano, data.FiveHour.ResetsAt); err == nil {
		rl.FiveHour = &domain.RateWindow{
			Utilization: data.FiveHour.Utilization,
			ResetsAt:    t,
		}
	}

	if t, err := time.Parse(time.RFC3339Nano, data.SevenDay.ResetsAt); err == nil {
		rl.SevenDay = &domain.RateWindow{
			Utilization: data.SevenDay.Utilization,
			ResetsAt:    t,
		}
	}

	return rl, nil
}
