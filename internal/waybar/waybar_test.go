package waybar

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mkrowiarz/ccmonitor/internal/domain"
)

func decode(t *testing.T, s string) output {
	t.Helper()
	var o output
	if err := json.Unmarshal([]byte(s), &o); err != nil {
		t.Fatalf("invalid JSON %q: %v", s, err)
	}
	return o
}

func TestRenderDisabled(t *testing.T) {
	s, err := Render(domain.RateLimits{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if o := decode(t, s); o.Class != "error" {
		t.Errorf("class = %q, want error", o.Class)
	}
}

func TestRenderNoData(t *testing.T) {
	s, err := Render(domain.RateLimits{Error: "API cooldown"}, true)
	if err != nil {
		t.Fatal(err)
	}
	o := decode(t, s)
	if o.Class != "error" {
		t.Errorf("class = %q, want error", o.Class)
	}
	if !strings.Contains(o.Tooltip, "API cooldown") {
		t.Errorf("tooltip = %q, want it to mention the error", o.Tooltip)
	}
}

func TestRenderClassByMaxUtilization(t *testing.T) {
	future := time.Now().Add(time.Hour)
	cases := []struct {
		util float64
		want string
	}{
		{0, "ok"},
		{49, "ok"},
		{50, "warning"},
		{79, "warning"},
		{80, "critical"},
		{100, "critical"},
	}
	for _, c := range cases {
		rl := domain.RateLimits{FiveHour: &domain.RateWindow{Utilization: c.util, ResetsAt: future}}
		s, err := Render(rl, true)
		if err != nil {
			t.Fatal(err)
		}
		if o := decode(t, s); o.Class != c.want {
			t.Errorf("util %.0f: class = %q, want %q", c.util, o.Class, c.want)
		}
	}
}

func TestRenderBothWindows(t *testing.T) {
	future := time.Now().Add(time.Hour)
	rl := domain.RateLimits{
		FiveHour: &domain.RateWindow{Utilization: 41.4, ResetsAt: future},
		SevenDay: &domain.RateWindow{Utilization: 7.2, ResetsAt: future},
	}
	s, err := Render(rl, true)
	if err != nil {
		t.Fatal(err)
	}
	o := decode(t, s)
	if !strings.Contains(o.Text, "5h 41%") || !strings.Contains(o.Text, "7d 7%") {
		t.Errorf("text = %q, want both windows rounded", o.Text)
	}
	// Class follows the higher window (41 -> ok).
	if o.Class != "ok" {
		t.Errorf("class = %q, want ok", o.Class)
	}
}
