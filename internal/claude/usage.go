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

	"github.com/mkrowiarz/ccmonitor/internal/domain"
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

// usageResult holds the result of a Get call including retry metadata.
type usageResult struct {
	Limits     *domain.RateLimits
	RetryAfter time.Time
	Warnings   []string
	Err        error
}

// usageClient fetches rate limit data with in-memory caching.
type usageClient struct {
	mu         sync.Mutex
	cached     *domain.RateLimits
	fetchedAt  time.Time
	lastErr    error
	retryAfter time.Time
	ttl        time.Duration
	cachePath  string
	tokenFn    func() (string, error)
}

func newUsageClient() *usageClient {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".ccmonitor")
	os.MkdirAll(cacheDir, 0755)
	u := &usageClient{
		ttl:       usageCacheTTL,
		cachePath: filepath.Join(cacheDir, "usage-cache.json"),
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
	// Restore cached window data if present
	if dc.FiveHour != nil || dc.SevenDay != nil {
		rl := &domain.RateLimits{FetchedAt: dc.FetchedAt}
		if dc.FiveHour != nil {
			rl.FiveHour = &domain.RateWindow{Utilization: dc.FiveHour.Utilization, ResetsAt: dc.FiveHour.ResetsAt}
		}
		if dc.SevenDay != nil {
			rl.SevenDay = &domain.RateWindow{Utilization: dc.SevenDay.Utilization, ResetsAt: dc.SevenDay.ResetsAt}
		}
		u.cached = rl
		u.fetchedAt = dc.FetchedAt
	}
	// Restore retry-after if still in the future
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
	// Clear error fields on success
	data, err := json.Marshal(dc)
	if err != nil {
		return
	}
	_ = os.WriteFile(u.cachePath, data, 0644)
}

func (u *usageClient) saveRetryState(errMsg string, retryAfter time.Time) {
	dc := diskCache{RetryAfter: retryAfter, LastError: errMsg}
	// Preserve existing window data if any
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
func (u *usageClient) Get(ctx context.Context) usageResult {
	u.mu.Lock()
	defer u.mu.Unlock()

	// If in cooldown, don't call the API at all
	if !u.retryAfter.IsZero() && time.Now().Before(u.retryAfter) {
		if u.cached != nil {
			return usageResult{Limits: u.cached}
		}
		return usageResult{Err: u.lastErr, RetryAfter: u.retryAfter}
	}

	// Return cached data if still fresh
	if u.cached != nil && time.Since(u.fetchedAt) < u.ttl {
		return usageResult{Limits: u.cached}
	}

	result, err := u.fetch(ctx)
	if err != nil {
		u.lastErr = err
		u.retryAfter = time.Now().Add(u.ttl + time.Second)
		u.saveRetryState(err.Error(), u.retryAfter)
		if u.cached != nil {
			return usageResult{Limits: u.cached}
		}
		return usageResult{Err: err, RetryAfter: u.retryAfter}
	}

	// Success — clear all error state
	u.lastErr = nil
	u.retryAfter = time.Time{}

	now := time.Now()
	result.FetchedAt = now
	u.cached = result
	u.fetchedAt = now
	u.saveDiskCache(result)
	return usageResult{Limits: result}
}

func (u *usageClient) fetch(ctx context.Context) (*domain.RateLimits, error) {
	token, err := u.tokenFn()
	if err != nil {
		return nil, fmt.Errorf("oauth token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, usageEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", usageBetaTag)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("usage API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("API cooldown")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("usage API %d", resp.StatusCode)
	}

	var data usageResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("usage API decode: %w", err)
	}

	return parseUsageResponse(&data)
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
