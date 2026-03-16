package claude

import (
	"testing"
	"time"
)

func TestParseElapsedDarwin(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"14:04", 14*time.Minute + 4*time.Second},
		{"10:35:43", 10*time.Hour + 35*time.Minute + 43*time.Second},
		{"1-10:35:43", 34*time.Hour + 35*time.Minute + 43*time.Second},
		{"0:05", 5 * time.Second},
		{"2-00:00:00", 48 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseElapsedDarwin(tt.input)
			if err != nil {
				t.Fatalf("parseElapsedDarwin(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseElapsedDarwin(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
