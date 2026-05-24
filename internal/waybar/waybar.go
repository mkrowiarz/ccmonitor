// Package waybar renders Claude rate-limit data as the JSON line consumed by a
// Waybar custom module (https://github.com/Alexays/Waybar). A module configured
// with "exec": "ccmonitor -waybar" and "return-type": "json" gets text, a
// tooltip, and a CSS class it can style by utilization.
package waybar

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/mkrowiarz/ccmonitor/internal/domain"
	"github.com/mkrowiarz/ccmonitor/internal/format"
)

// icon is the nf-md-robot glyph; rendering it requires a Nerd Font in the bar.
const icon = "\U000f06a9"

// Utilization thresholds (percent) at which the CSS class escalates.
const (
	warnThreshold = 50
	critThreshold = 80
)

// output is the JSON shape Waybar expects from a custom module's exec command.
type output struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
}

// Render returns a single line of Waybar JSON describing the rate-limit
// windows. enabled reports whether the platform supports rate limits at all.
// It always produces valid, renderable JSON; the only error is from marshaling.
func Render(rl domain.RateLimits, enabled bool) (string, error) {
	if !enabled {
		return marshal(output{
			Text:    icon + " n/a",
			Tooltip: "Claude rate limits are not available on this platform",
			Class:   "error",
		})
	}

	if rl.FiveHour == nil && rl.SevenDay == nil {
		msg := rl.Error
		if msg == "" {
			msg = "no usage data (open Claude Code to refresh the token)"
		}
		return marshal(output{
			Text:    icon + " !",
			Tooltip: "Claude usage unavailable: " + msg,
			Class:   "error",
		})
	}

	maxUtil := 0.0
	parts := make([]string, 0, 2)
	var tip strings.Builder
	tip.WriteString("Claude Code usage")

	if w := rl.FiveHour; w != nil {
		u := int(math.Round(w.Utilization))
		parts = append(parts, fmt.Sprintf("5h %d%%", u))
		fmt.Fprintf(&tip, "\n5-hour:  %3d%%  (resets in %s)", u, format.FormatUptime(time.Until(w.ResetsAt)))
		maxUtil = math.Max(maxUtil, w.Utilization)
	}
	if w := rl.SevenDay; w != nil {
		u := int(math.Round(w.Utilization))
		parts = append(parts, fmt.Sprintf("7d %d%%", u))
		fmt.Fprintf(&tip, "\n7-day:   %3d%%  (resets in %s)", u, format.FormatUptime(time.Until(w.ResetsAt)))
		maxUtil = math.Max(maxUtil, w.Utilization)
	}

	class := "ok"
	switch {
	case maxUtil >= critThreshold:
		class = "critical"
	case maxUtil >= warnThreshold:
		class = "warning"
	}

	return marshal(output{
		Text:    icon + " " + strings.Join(parts, " · "),
		Tooltip: tip.String(),
		Class:   class,
	})
}

func marshal(o output) (string, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
