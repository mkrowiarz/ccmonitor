package waybar

import (
	"strings"
	"testing"
)

func TestSetupTextEmbedsExecPath(t *testing.T) {
	got := SetupText("/home/u/.local/bin/ccmonitor")
	if !strings.Contains(got, `"/home/u/.local/bin/ccmonitor -waybar"`) {
		t.Errorf("setup text missing exec path with -waybar:\n%s", got)
	}
	for _, want := range []string{"custom/claude", "modules-right", "#custom-claude.critical", "SIGUSR2"} {
		if !strings.Contains(got, want) {
			t.Errorf("setup text missing %q", want)
		}
	}
}

func TestSetupTextDefaultsExecPath(t *testing.T) {
	if got := SetupText(""); !strings.Contains(got, `"ccmonitor -waybar"`) {
		t.Errorf("empty exec path should default to ccmonitor:\n%s", got)
	}
}
