package format

import (
	"testing"
	"time"
)

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{0, "0s"},
		{45 * time.Second, "45s"},
		{14*time.Minute + 4*time.Second, "14m"},
		{8*time.Hour + 44*time.Minute, "8h44m"},
		{34*time.Hour + 35*time.Minute + 43*time.Second, "34h35m"},
	}

	for _, tt := range tests {
		got := FormatUptime(tt.input)
		if got != tt.expected {
			t.Errorf("FormatUptime(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
