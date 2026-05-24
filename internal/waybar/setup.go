package waybar

import (
	"fmt"
	"strings"
)

// SetupText returns copy-paste instructions for adding the ccmonitor module to
// Waybar. execPath is embedded into the module's exec line so the module works
// regardless of the bar's PATH. The output is plain text (no ANSI) so the
// snippets paste cleanly into config files.
func SetupText(execPath string) string {
	if execPath == "" {
		execPath = "ccmonitor"
	}

	var b strings.Builder
	b.WriteString("Waybar setup for ccmonitor\n")
	b.WriteString("==========================\n\n")

	b.WriteString("1. Add this module to ~/.config/waybar/config.jsonc:\n\n")
	fmt.Fprintf(&b, `    "custom/claude": {
        "exec": %q,
        "return-type": "json",
        "interval": 60,
        "tooltip": true
        // "on-click": "$TERMINAL -e ccmonitor"   // optional: open the dashboard
    },
`, execPath+" -waybar")

	b.WriteString("\n2. Reference \"custom/claude\" in one of your modules-* arrays, e.g.:\n\n")
	b.WriteString("    \"modules-right\": [\"custom/claude\", \"clock\"]\n")

	b.WriteString("\n3. Add styling to ~/.config/waybar/style.css (optional):\n\n")
	b.WriteString("    #custom-claude.ok       { color: #a6e3a1; }            /* < 50%  */\n")
	b.WriteString("    #custom-claude.warning  { color: #f9e2af; }            /* 50-80% */\n")
	b.WriteString("    #custom-claude.critical { color: #f38ba8; font-weight: bold; } /* >= 80% */\n")
	b.WriteString("    #custom-claude.error    { color: #6c7086; }            /* token missing / API down */\n")

	b.WriteString("\n4. Reload Waybar:\n\n")
	b.WriteString("    killall -SIGUSR2 waybar\n")

	b.WriteString("\nThe module text starts with a Nerd Font glyph; use a Nerd Font in your bar.\n")
	return b.String()
}
