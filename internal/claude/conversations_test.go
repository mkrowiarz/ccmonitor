package claude

import (
	"testing"
	"time"
)

func TestConversationScannerComputeStats(t *testing.T) {
	// Use testdata directory as a fake claude dir (has projects/ subdir).
	claudeDir := testdataPath("")

	scanner := newConversationScanner()
	summary, warnings, err := scanner.computeStats(claudeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if len(warnings) > 0 {
		t.Logf("warnings: %v", warnings)
	}

	today := time.Now().Format("2006-01-02")

	// Source date should be today.
	if summary.SourceDate != today {
		t.Errorf("SourceDate = %q, want %q", summary.SourceDate, today)
	}

	// Lifetime: 4 user messages across 2 sessions.
	if summary.LifetimeMessages == nil || *summary.LifetimeMessages != 4 {
		var v int64
		if summary.LifetimeMessages != nil {
			v = *summary.LifetimeMessages
		}
		t.Errorf("LifetimeMessages = %d, want 4", v)
	}
	if summary.LifetimeSessions == nil || *summary.LifetimeSessions != 2 {
		var v int64
		if summary.LifetimeSessions != nil {
			v = *summary.LifetimeSessions
		}
		t.Errorf("LifetimeSessions = %d, want 2", v)
	}

	// Lifetime tokens: opus=100+50+200+300+100+500+200+100=1550, sonnet=50+25=75
	lifetimeByModel := make(map[string]int64)
	for _, lt := range summary.LifetimeTokens {
		lifetimeByModel[lt.ModelName] = lt.TokenCount
	}
	if lifetimeByModel["claude-opus-4-6"] != 1550 {
		t.Errorf("lifetime opus tokens = %d, want 1550", lifetimeByModel["claude-opus-4-6"])
	}
	if lifetimeByModel["claude-sonnet-4-6"] != 75 {
		t.Errorf("lifetime sonnet tokens = %d, want 75", lifetimeByModel["claude-sonnet-4-6"])
	}

	// Daily activity should include both dates.
	if len(summary.DailyActivity) < 2 {
		t.Fatalf("DailyActivity count = %d, want >= 2", len(summary.DailyActivity))
	}

	// Caching: second call should use cache (same results).
	summary2, _, err := scanner.computeStats(claudeDir)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if *summary2.LifetimeMessages != *summary.LifetimeMessages {
		t.Errorf("cached result differs: %d vs %d", *summary2.LifetimeMessages, *summary.LifetimeMessages)
	}
}

func TestConversationScannerEmptyDir(t *testing.T) {
	scanner := newConversationScanner()
	summary, _, err := scanner.computeStats(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.LifetimeMessages == nil || *summary.LifetimeMessages != 0 {
		t.Errorf("expected 0 lifetime messages for empty dir")
	}
}
