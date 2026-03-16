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
