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
