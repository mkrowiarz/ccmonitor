package tui

import (
	"github.com/michal/ccmonitor/internal/domain"
)

// TickMsg triggers a new data collection cycle.
type TickMsg struct{}

// SnapshotMsg carries the result of a collection cycle.
type SnapshotMsg struct {
	Snapshot *domain.BackendSnapshot
	Err      error
}
