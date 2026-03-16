package claude

import (
	"encoding/json"
	"os"
	"testing"
	"time"
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
