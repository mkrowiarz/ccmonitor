package claude

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/mkrowiarz/ccmonitor/internal/backend"
	"github.com/mkrowiarz/ccmonitor/internal/domain"
)

// ClaudeBackend implements backend.Backend for local Claude Code monitoring.
type ClaudeBackend struct {
	claudeDir   string
	usageClient *usageClient
	convScanner *conversationScanner
}

// Ensure ClaudeBackend implements the backend interfaces.
var (
	_ backend.Backend           = (*ClaudeBackend)(nil)
	_ backend.RateLimitProvider = (*ClaudeBackend)(nil)
)

// New creates a ClaudeBackend using ~/.claude as the data directory.
func New() *ClaudeBackend {
	home, _ := os.UserHomeDir()
	return &ClaudeBackend{
		claudeDir:   filepath.Join(home, ".claude"),
		usageClient: newUsageClient(),
		convScanner: newConversationScanner(),
	}
}

// NewWithDir creates a ClaudeBackend with a custom directory (for testing).
func NewWithDir(dir string) *ClaudeBackend {
	return &ClaudeBackend{
		claudeDir:   dir,
		usageClient: newUsageClient(),
		convScanner: newConversationScanner(),
	}
}

// Name returns the backend name.
func (b *ClaudeBackend) Name() string {
	return "claude"
}

// RateLimits fetches only the usage/rate-limit windows, skipping the heavier
// process, conversation, and history scans done by Collect. The bool reports
// whether rate limits are supported on the current platform. Results are served
// from the usage client's on-disk cache between API refreshes, so this is cheap
// to call on a short polling interval. Implements backend.RateLimitProvider.
func (b *ClaudeBackend) RateLimits(ctx context.Context) (domain.RateLimits, bool) {
	if !rateLimitsSupported() {
		return domain.RateLimits{}, false
	}
	var rl domain.RateLimits
	ur := b.usageClient.Get(ctx)
	if ur.Err != nil {
		rl.Error = ur.Err.Error()
		rl.RetryAfter = ur.RetryAfter
	} else if ur.Limits != nil {
		rl = *ur.Limits
	}
	return rl, true
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

	// 2. Compute stats from conversation JSONL files
	usage, statsWarnings, statsErr := b.convScanner.computeStats(b.claudeDir)
	snap.Warnings = append(snap.Warnings, statsWarnings...)
	if statsErr != nil {
		snap.Warnings = append(snap.Warnings, "conversation scan error: "+statsErr.Error())
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

	// 4. Fetch rate limits from usage API (best-effort)
	if rl, enabled := b.RateLimits(ctx); enabled {
		snap.RateLimitsEnabled = true
		snap.RateLimits = rl
	}

	// Determine overall status
	if len(snap.Warnings) > 0 && snap.Status == domain.StatusOk {
		snap.Status = domain.StatusDegraded
	}

	return snap, nil
}
