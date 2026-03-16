package claude

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/michal/ccmonitor/internal/backend"
	"github.com/michal/ccmonitor/internal/domain"
)

// ClaudeBackend implements backend.Backend for local Claude Code monitoring.
type ClaudeBackend struct {
	claudeDir   string
	usageClient *usageClient
}

// Ensure ClaudeBackend implements backend.Backend.
var _ backend.Backend = (*ClaudeBackend)(nil)

// New creates a ClaudeBackend using ~/.claude as the data directory.
func New() *ClaudeBackend {
	home, _ := os.UserHomeDir()
	return &ClaudeBackend{
		claudeDir:   filepath.Join(home, ".claude"),
		usageClient: newUsageClient(),
	}
}

// NewWithDir creates a ClaudeBackend with a custom directory (for testing).
func NewWithDir(dir string) *ClaudeBackend {
	return &ClaudeBackend{
		claudeDir:   dir,
		usageClient: newUsageClient(),
	}
}

// Name returns the backend name.
func (b *ClaudeBackend) Name() string {
	return "claude"
}

// Collect gathers a snapshot of Claude Code activity.
func (b *ClaudeBackend) Collect(ctx context.Context, opts backend.CollectOpts) (*domain.BackendSnapshot, error) {
	snap := &domain.BackendSnapshot{
		BackendName: b.Name(),
		CollectedAt: time.Now(),
		Status:      domain.StatusOk,
	}

	// 1. Discover processes
	procs, err := discoverProcesses(ctx)
	if err != nil {
		snap.Warnings = append(snap.Warnings, "process discovery failed: "+err.Error())
		snap.Status = domain.StatusDegraded
	} else {
		for _, p := range procs {
			uptime, parseErr := parseElapsed(p.Elapsed)
			if parseErr != nil {
				snap.Warnings = append(snap.Warnings, "elapsed parse error: "+parseErr.Error())
			}
			projName := resolveProjectName(p.PID)
			snap.ActiveSessions = append(snap.ActiveSessions, domain.ActiveSession{
				PID:         p.PID,
				ProjectName: projName,
				CPUPercent:  p.CPUPercent,
				MemPercent:  p.MemPercent,
				Uptime:      uptime,
			})
		}
	}

	// 2. Parse stats cache
	statsPath := filepath.Join(b.claudeDir, "stats-cache.json")
	usage, statsWarnings, statsErr := parseStatsCache(statsPath)
	snap.Warnings = append(snap.Warnings, statsWarnings...)
	if statsErr != nil {
		snap.Warnings = append(snap.Warnings, "stats-cache error: "+statsErr.Error())
		snap.Status = domain.StatusDegraded
	} else if usage != nil {
		snap.Usage = *usage
	}

	// 3. Parse history if requested
	if opts.IncludeRecentActivity {
		limit := opts.RecentActivityLimit
		if limit <= 0 {
			limit = 10
		}
		histPath := filepath.Join(b.claudeDir, "history.jsonl")
		events, histWarnings, histErr := parseHistory(histPath, limit)
		snap.Warnings = append(snap.Warnings, histWarnings...)
		if histErr != nil {
			snap.Warnings = append(snap.Warnings, "history error: "+histErr.Error())
			snap.Status = domain.StatusDegraded
		} else {
			snap.RecentEvents = events
		}
	}

	// 4. Fetch rate limits from usage API
	rateLimits, usageWarnings, usageErr := b.usageClient.Get(ctx)
	snap.Warnings = append(snap.Warnings, usageWarnings...)
	if usageErr != nil {
		snap.Warnings = append(snap.Warnings, "rate limits: "+usageErr.Error())
	} else if rateLimits != nil {
		snap.RateLimits = *rateLimits
	}

	// Determine overall status
	if len(snap.Warnings) > 0 && snap.Status == domain.StatusOk {
		snap.Status = domain.StatusDegraded
	}

	return snap, nil
}
