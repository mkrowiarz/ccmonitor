package format

import "testing"

func TestFormatCount(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{56, "56"},
		{999, "999"},
		{1000, "1.0K"},
		{4000, "4.0K"},
		{10000, "10K"},
		{48000, "48K"},
		{107570, "107K"},
		{155000, "155K"},
		{265003000, "265M"},
		{417000000, "417M"},
		{1000000000, "1.0B"},
		{1934004200, "1B"},
	}

	for _, tt := range tests {
		got := FormatCount(tt.input)
		if got != tt.expected {
			t.Errorf("FormatCount(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
