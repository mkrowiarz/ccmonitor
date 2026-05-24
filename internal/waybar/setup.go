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
        // "on-click": "$TERMINAL -e ccmonitor"   // optional: open the dashboard (see step 4)
    },
`, execPath+" -waybar")

	b.WriteString("\n2. Reference \"custom/claude\" in one of your modules-* arrays, e.g.:\n\n")
	b.WriteString("    \"modules-right\": [\"custom/claude\", \"clock\"]\n")

	b.WriteString("\n3. Add styling to ~/.config/waybar/style.css (optional).\n")
	b.WriteString("   The first block gives the module the same \"pill\" look most\n")
	b.WriteString("   themes use; tweak it to match your bar, or fold #custom-claude\n")
	b.WriteString("   into your existing pill selector. The color classes layer on top.\n\n")
	b.WriteString("    #custom-claude {\n")
	b.WriteString("        padding: 4px 10px;\n")
	b.WriteString("        margin: 0px 2px;\n")
	b.WriteString("        background-color: rgba(48, 52, 70, 0.35);\n")
	b.WriteString("        border: 2px solid rgba(98, 104, 128, 0);\n")
	b.WriteString("        border-radius: 5px;\n")
	b.WriteString("    }\n")
	b.WriteString("    #custom-claude:hover    { border: 2px solid rgba(98, 104, 128, 1); }\n\n")
	b.WriteString("    #custom-claude.ok       { color: #a6e3a1; }            /* < 50%  */\n")
	b.WriteString("    #custom-claude.warning  { color: #f9e2af; }            /* 50-80% */\n")
	b.WriteString("    #custom-claude.critical { color: #f38ba8; font-weight: bold; } /* >= 80% */\n")
	b.WriteString("    #custom-claude.error    { color: #6c7086; }            /* token missing / API down */\n")

	b.WriteString("\n4. Click to open the full dashboard in a popup (optional, Hyprland).\n")
	b.WriteString("   Simplest: set the on-click above to \"$TERMINAL -e ccmonitor\".\n")
	b.WriteString("   For a floating, toggleable dropdown, point on-click at a script:\n\n")
	b.WriteString("        \"on-click\": \"~/.config/waybar/scripts/ccmonitor-toggle.sh\"\n\n")
	b.WriteString("   ~/.config/waybar/scripts/ccmonitor-toggle.sh:\n\n")
	fmt.Fprintf(&b, `        #!/usr/bin/env bash
        set -euo pipefail
        class=com.ccmonitor.popup   # must be a valid GTK app-id (reverse-DNS, has a dot)
        ws=ccmonitor
        if hyprctl clients -j | jq -e ".[] | select(.class==\"$class\")" >/dev/null; then
            hyprctl dispatch togglespecialworkspace "$ws"
        else
            hyprctl dispatch exec "[workspace special:$ws silent] %s --class=$class -e ccmonitor"
            hyprctl dispatch togglespecialworkspace "$ws"
        fi
`, "$TERMINAL")
	b.WriteString("\n   and a Hyprland window rule (float + center the popup):\n\n")
	b.WriteString("        windowrulev2 = float, class:^(com\\.ccmonitor\\.popup)$\n")
	b.WriteString("        windowrulev2 = size 900 520, class:^(com\\.ccmonitor\\.popup)$\n")
	b.WriteString("        windowrulev2 = center, class:^(com\\.ccmonitor\\.popup)$\n\n")
	b.WriteString("   Note: some terminals (e.g. ghostty) ignore --class unless it is a\n")
	b.WriteString("   valid GTK app-id containing a dot; a bare name like 'ccmonitor-popup'\n")
	b.WriteString("   is silently dropped, which breaks the toggle's already-open check.\n")

	b.WriteString("\n5. Reload Waybar:\n\n")
	b.WriteString("    killall -SIGUSR2 waybar\n")

	b.WriteString("\nThe module text starts with a Nerd Font glyph; use a Nerd Font in your bar.\n")
	return b.String()
}
