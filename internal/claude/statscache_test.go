package claude

import (
	"path/filepath"
	"runtime"
	"testing"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", name)
}

func TestParseStatsCacheValid(t *testing.T) {
	summary, warnings, err := parseStatsCache(testdataPath("stats-cache-valid.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}

	// Today messages and sessions
	if summary.TodayMessages == nil || *summary.TodayMessages != 4000 {
		t.Errorf("TodayMessages = %v, want 4000", summary.TodayMessages)
	}
	if summary.TodaySessions == nil || *summary.TodaySessions != 56 {
		t.Errorf("TodaySessions = %v, want 56", summary.TodaySessions)
	}

	// Today tokens: 3 models
	if len(summary.TodayTokens) != 3 {
		t.Errorf("TodayTokens count = %d, want 3", len(summary.TodayTokens))
	}

	// Lifetime messages
	if summary.LifetimeMessages == nil || *summary.LifetimeMessages != 107570 {
		t.Errorf("LifetimeMessages = %v, want 107570", summary.LifetimeMessages)
	}

	// Lifetime sessions
	if summary.LifetimeSessions == nil || *summary.LifetimeSessions != 1147 {
		t.Errorf("LifetimeSessions = %v, want 1147", summary.LifetimeSessions)
	}

	// Lifetime tokens: check sums
	// claude-opus-4-6: 1000+2000+246000000+18000000 = 264003000
	// claude-sonnet-4-5: 500+1500+384000000+30000000 = 414002000
	// claude-sonnet-4-6: 200+4000+1867000000+63000000 = 1930004200
	expectedLifetime := map[string]int64{
		"claude-opus-4-6":   264003000,
		"claude-sonnet-4-5": 414002000,
		"claude-sonnet-4-6": 1930004200,
	}
	if len(summary.LifetimeTokens) != 3 {
		t.Fatalf("LifetimeTokens count = %d, want 3", len(summary.LifetimeTokens))
	}
	for _, lt := range summary.LifetimeTokens {
		want, ok := expectedLifetime[lt.ModelName]
		if !ok {
			t.Errorf("unexpected model %q in lifetime tokens", lt.ModelName)
			continue
		}
		if lt.TokenCount != want {
			t.Errorf("lifetime tokens for %s = %d, want %d", lt.ModelName, lt.TokenCount, want)
		}
	}

	// Daily activity
	if len(summary.DailyActivity) != 2 {
		t.Fatalf("DailyActivity count = %d, want 2", len(summary.DailyActivity))
	}
	if summary.DailyActivity[0].Date != "2026-03-08" {
		t.Errorf("DailyActivity[0].Date = %q, want %q", summary.DailyActivity[0].Date, "2026-03-08")
	}
	if summary.DailyActivity[0].MessageCount != 100 {
		t.Errorf("DailyActivity[0].MessageCount = %d, want 100", summary.DailyActivity[0].MessageCount)
	}
	if summary.DailyActivity[1].Date != "2026-03-09" {
		t.Errorf("DailyActivity[1].Date = %q, want %q", summary.DailyActivity[1].Date, "2026-03-09")
	}
	if summary.DailyActivity[1].MessageCount != 4000 {
		t.Errorf("DailyActivity[1].MessageCount = %d, want 4000", summary.DailyActivity[1].MessageCount)
	}

	// Source date
	if summary.SourceDate != "2026-03-09" {
		t.Errorf("SourceDate = %q, want %q", summary.SourceDate, "2026-03-09")
	}

	_ = warnings // warnings are acceptable
}

func TestParseStatsCacheMalformed(t *testing.T) {
	_, _, err := parseStatsCache(testdataPath("stats-cache-malformed.json"))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestParseStatsCacheMissing(t *testing.T) {
	summary, warnings, err := parseStatsCache(testdataPath("nonexistent-file.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != nil {
		t.Error("expected nil summary for missing file")
	}
	if len(warnings) == 0 {
		t.Error("expected at least one warning for missing file")
	}
}
