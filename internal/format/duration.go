package format

import (
	"fmt"
	"time"
)

// FormatUptime formats a duration into compact form like "8h44m", "14m", "45s".
func FormatUptime(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	switch {
	case h >= 1:
		return fmt.Sprintf("%dh%dm", h, m)
	case m >= 1:
		return fmt.Sprintf("%dm", m)
	default:
		return fmt.Sprintf("%ds", s)
	}
}

// FormatReset formats a time-until-reset duration. Long windows (>24h) render
// as "3d 11h 33m" so multi-day resets read clearly; shorter ones fall back to
// the compact FormatUptime form like "4h28m".
func FormatReset(d time.Duration) string {
	days := int(d.Hours()) / 24
	h := int(d.Hours()) % 24
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	switch {
	case days >= 1:
		return fmt.Sprintf("%dd %dh %dm", days, h, m)
	case h >= 1:
		return fmt.Sprintf("%dh %dm", h, m)
	case m >= 1:
		return fmt.Sprintf("%dm", m)
	default:
		return fmt.Sprintf("%ds", s)
	}
}
