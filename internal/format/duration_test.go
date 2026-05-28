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

func TestFormatReset(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{45 * time.Second, "45s"},
		{4*time.Hour + 28*time.Minute, "4h 28m"},
		{23*time.Hour + 59*time.Minute, "23h 59m"},
		{24 * time.Hour, "1d 0h 0m"},
		{51*time.Hour + 48*time.Minute, "2d 3h 48m"},
		{3*24*time.Hour + 11*time.Hour + 33*time.Minute, "3d 11h 33m"},
	}

	for _, tt := range tests {
		got := FormatReset(tt.input)
		if got != tt.expected {
			t.Errorf("FormatReset(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
