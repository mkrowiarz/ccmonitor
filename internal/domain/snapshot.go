package domain

import "time"

// BackendStatus represents the health state of a backend.
type BackendStatus int

const (
	StatusOk BackendStatus = iota
	StatusDegraded
	StatusUnavailable
)

func (s BackendStatus) String() string {
	switch s {
	case StatusOk:
		return "ok"
	case StatusDegraded:
		return "degraded"
	case StatusUnavailable:
		return "unavailable"
	default:
		return "unknown"
	}
}

// BackendSnapshot is the normalized result of a single backend collection cycle.
type BackendSnapshot struct {
	BackendName    string
	Status         BackendStatus
	CollectedAt    time.Time
	ActiveSessions []ActiveSession
	Usage          UsageSummary
	RecentEvents   []RecentEvent
	Warnings       []string
}

// ActiveSession represents a single running Claude process.
type ActiveSession struct {
	PID         int
	ProjectName string
	CPUPercent  float64
	MemPercent  float64
	Uptime      time.Duration
}

// ModelTokens holds token count for a single model.
type ModelTokens struct {
	ModelName  string
	TokenCount int64
}

// UsageSummary holds today and lifetime usage metrics.
type UsageSummary struct {
	TodayMessages    *int64
	TodaySessions    *int64
	TodayTokens      []ModelTokens
	LifetimeMessages *int64
	LifetimeSessions *int64
	LifetimeTokens   []ModelTokens
	SourceDate       string // lastComputedDate from cache
}

// RecentEvent represents a single activity entry from history.
type RecentEvent struct {
	Timestamp   time.Time
	ProjectName string
	Display     string
	SessionID   string
}
