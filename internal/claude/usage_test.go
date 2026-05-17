package claude

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mkrowiarz/ccmonitor/internal/domain"
)

func TestParseUsageResponse(t *testing.T) {
	data, err := os.ReadFile("../../testdata/usage-response.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var resp usageResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	rl, err := parseUsageResponse(&resp)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if rl.FiveHour == nil {
		t.Fatal("expected FiveHour to be set")
	}
	if rl.FiveHour.Utilization != 11.0 {
		t.Errorf("FiveHour utilization = %v, want 11.0", rl.FiveHour.Utilization)
	}
	wantReset := time.Date(2026, 3, 16, 23, 0, 0, 536281000, time.UTC)
	if !rl.FiveHour.ResetsAt.Equal(wantReset) {
		t.Errorf("FiveHour resets_at = %v, want %v", rl.FiveHour.ResetsAt, wantReset)
	}

	if rl.SevenDay == nil {
		t.Fatal("expected SevenDay to be set")
	}
	if rl.SevenDay.Utilization != 20.0 {
		t.Errorf("SevenDay utilization = %v, want 20.0", rl.SevenDay.Utilization)
	}
}

// TestGet_ExpiredWindowForcesRefetch verifies that a cached window whose
// ResetsAt has passed is dropped and a refetch is forced, even while the
// client is in API cooldown. Prevents stale utilization being shown after
// the window has reset server-side.
func TestGet_ExpiredWindowForcesRefetch(t *testing.T) {
	now := time.Now()
	dir := t.TempDir()

	expiredFiveHour := &domain.RateWindow{Utilization: 14.0, ResetsAt: now.Add(-5 * time.Minute)}
	freshSevenDay := &domain.RateWindow{Utilization: 47.0, ResetsAt: now.Add(48 * time.Hour)}

	fetchCalls := 0
	u := &usageClient{
		ttl:       usageCacheTTL,
		cachePath: filepath.Join(dir, "usage-cache.json"),
		cached: &domain.RateLimits{
			FiveHour:  expiredFiveHour,
			SevenDay:  freshSevenDay,
			FetchedAt: now.Add(-2 * time.Minute), // fresh by TTL
		},
		fetchedAt:  now.Add(-2 * time.Minute),
		retryAfter: now.Add(5 * time.Minute), // in cooldown
		lastErr:    errors.New("simulated API failure"),
		tokenFn: func() (string, error) {
			fetchCalls++
			return "", errors.New("no token in test")
		},
	}

	res := u.Get(context.Background())

	// Expired 5h window must be gone from result.
	if res.Limits != nil && res.Limits.FiveHour != nil {
		t.Errorf("expired FiveHour should be dropped, got %+v", res.Limits.FiveHour)
	}
	// Fetch must have been attempted (cooldown bypassed).
	if fetchCalls != 1 {
		t.Errorf("expected refetch attempt, got fetchCalls=%d", fetchCalls)
	}
	// Cached state in client should also have FiveHour nil now.
	if u.cached != nil && u.cached.FiveHour != nil {
		t.Errorf("client cache should have dropped expired FiveHour, got %+v", u.cached.FiveHour)
	}
}

// TestGet_MaxStaleGuardRejectsAncientCache verifies that cache older than
// usageMaxStale is not returned even when fetch fails and the client would
// otherwise fall back to cached data.
func TestGet_MaxStaleGuardRejectsAncientCache(t *testing.T) {
	now := time.Now()
	dir := t.TempDir()

	ancient := now.Add(-15 * time.Hour)
	u := &usageClient{
		ttl:       usageCacheTTL,
		cachePath: filepath.Join(dir, "usage-cache.json"),
		cached: &domain.RateLimits{
			FiveHour:  &domain.RateWindow{Utilization: 14.0, ResetsAt: now.Add(1 * time.Hour)},
			SevenDay:  &domain.RateWindow{Utilization: 47.0, ResetsAt: now.Add(48 * time.Hour)},
			FetchedAt: ancient,
		},
		fetchedAt: ancient,
		tokenFn: func() (string, error) {
			return "", errors.New("simulated token failure")
		},
	}

	res := u.Get(context.Background())

	if res.Err == nil {
		t.Fatal("expected error result when cache is past usageMaxStale and fetch fails")
	}
	if res.Limits != nil {
		t.Errorf("expected nil Limits past usageMaxStale, got %+v", res.Limits)
	}
}

// TestLoadDiskCache_DropsExpiredWindows verifies that on-disk windows whose
// ResetsAt has passed are not restored into the in-memory cache.
func TestLoadDiskCache_DropsExpiredWindows(t *testing.T) {
	now := time.Now()
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "usage-cache.json")

	dc := diskCache{
		FetchedAt: now.Add(-30 * time.Minute),
		FiveHour:  &diskWindow{Utilization: 14.0, ResetsAt: now.Add(-1 * time.Minute)},
		SevenDay:  &diskWindow{Utilization: 47.0, ResetsAt: now.Add(48 * time.Hour)},
	}
	data, err := json.Marshal(dc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	u := &usageClient{ttl: usageCacheTTL, cachePath: cachePath}
	u.loadDiskCache()

	if u.cached == nil {
		t.Fatal("expected cached to be set with surviving SevenDay")
	}
	if u.cached.FiveHour != nil {
		t.Errorf("expected expired FiveHour to be dropped, got %+v", u.cached.FiveHour)
	}
	if u.cached.SevenDay == nil || u.cached.SevenDay.Utilization != 47.0 {
		t.Errorf("expected fresh SevenDay preserved, got %+v", u.cached.SevenDay)
	}
}
