package backend

import (
	"context"

	"github.com/mkrowiarz/ccmonitor/internal/domain"
)

// Backend is the interface that each monitoring backend must implement.
type Backend interface {
	Name() string
	Collect(ctx context.Context, opts CollectOpts) (*domain.BackendSnapshot, error)
}

// CollectOpts controls what data is collected.
type CollectOpts struct {
	IncludeRecentActivity bool
	RecentActivityLimit   int
}

// RateLimitProvider is an optional capability implemented by backends that can
// report usage/rate-limit windows without running a full Collect cycle. It is
// used by lightweight output modes such as -waybar, which poll frequently and
// must not pay for process discovery and conversation scans on every tick.
type RateLimitProvider interface {
	// RateLimits returns the current rate-limit windows. The bool reports
	// whether rate limits are supported on the current platform.
	RateLimits(ctx context.Context) (domain.RateLimits, bool)
}
