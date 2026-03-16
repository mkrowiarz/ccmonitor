package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/michal/ccmonitor/internal/domain"
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

// usageClient fetches rate limit data with in-memory caching.
type usageClient struct {
	mu        sync.Mutex
	cached    *domain.RateLimits
	fetchedAt time.Time
	ttl       time.Duration
	tokenFn   func() (string, error)
}

func newUsageClient() *usageClient {
	return &usageClient{
		ttl:     usageCacheTTL,
		tokenFn: readOAuthToken,
	}
}

// Get returns cached rate limits if fresh, otherwise fetches from API.
func (u *usageClient) Get(ctx context.Context) (*domain.RateLimits, []string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.cached != nil && time.Since(u.fetchedAt) < u.ttl {
		return u.cached, nil, nil
	}

	result, err := u.fetch(ctx)
	if err != nil {
		// Graceful degradation: return stale cache if available
		if u.cached != nil {
			return u.cached, []string{"rate limits stale: " + err.Error()}, nil
		}
		return nil, nil, err
	}

	u.cached = result
	u.fetchedAt = time.Now()
	return result, nil, nil
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
		return nil, fmt.Errorf("usage API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("usage API %d: %s", resp.StatusCode, string(body))
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
