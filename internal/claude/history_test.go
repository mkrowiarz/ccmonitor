package claude

import (
	"testing"
)

func TestParseHistoryValid(t *testing.T) {
	events, warnings, err := parseHistory(testdataPath("history-valid.jsonl"), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("events count = %d, want 3", len(events))
	}

	// Should be sorted by timestamp desc
	if events[0].SessionID != "abc-123" || events[0].Display != "/exit" {
		t.Errorf("first event = %+v, want /exit abc-123", events[0])
	}
	if events[1].SessionID != "def-456" {
		t.Errorf("second event sessionID = %q, want def-456", events[1].SessionID)
	}
	if events[2].Display != "Yes" {
		t.Errorf("third event display = %q, want Yes", events[2].Display)
	}

	// Check project names are last path component
	if events[0].ProjectName != "foo" {
		t.Errorf("first event project = %q, want foo", events[0].ProjectName)
	}
	if events[1].ProjectName != "bar" {
		t.Errorf("second event project = %q, want bar", events[1].ProjectName)
	}

	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestParseHistoryMalformed(t *testing.T) {
	events, warnings, err := parseHistory(testdataPath("history-malformed.jsonl"), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events count = %d, want 2", len(events))
	}

	// Should have warning about 2 skipped lines
	if len(warnings) != 1 {
		t.Fatalf("warnings count = %d, want 1", len(warnings))
	}
	if warnings[0] != "skipped 2 malformed lines" {
		t.Errorf("warning = %q, want %q", warnings[0], "skipped 2 malformed lines")
	}
}

func TestParseHistoryMissing(t *testing.T) {
	events, warnings, err := parseHistory(testdataPath("nonexistent.jsonl"), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if events != nil {
		t.Error("expected nil events for missing file")
	}
	if len(warnings) == 0 {
		t.Error("expected at least one warning for missing file")
	}
}
