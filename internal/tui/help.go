package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHelpOverlay renders a centered help modal with keyboard shortcuts
// and optionally a rate limits guide.
func renderHelpOverlay(s Styles, width, height int, minimal, showRateLimits bool, bg string) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorModel)).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorLabel))

	// Keyboard shortcuts
	keysTitle := s.Title.Render("KEYBOARD SHORTCUTS")

	keys := []struct{ key, desc string }{
		{"q", "Quit"},
		{"r", "Force refresh"},
		{"?", "Toggle this help"},
	}
	if !minimal {
		keys = append(keys,
			struct{ key, desc string }{"tab", "Cycle tabs"},
			struct{ key, desc string }{"1", "Dashboard"},
			struct{ key, desc string }{"2", "Activity"},
			struct{ key, desc string }{"3", "Analytics"},
		)
	}

	var keyRows []string
	for _, k := range keys {
		keyRows = append(keyRows, keyStyle.Render(padRight(k.key, 8))+descStyle.Render(k.desc))
	}
	keysPanel := lipgloss.JoinVertical(lipgloss.Left, keysTitle, "", strings.Join(keyRows, "\n"))

	var modalContent string

	if showRateLimits {
		// Modal overhead: border(2) + padding(2) + blank(1) + dismiss(1) = 6
		const modalOverhead = 6
		availHeight := height - 2 // header + footer

		rlTall := renderRateLimitsHelpTall(s, keyStyle, descStyle)
		tallHeight := lipgloss.Height(keysPanel) + 1 + lipgloss.Height(rlTall) + modalOverhead

		if tallHeight <= availHeight {
			// Tall pane: single column (keys, then RL below)
			modalContent = lipgloss.JoinVertical(lipgloss.Left,
				keysPanel, "", rlTall,
			)
		} else {
			// Short pane: keys left, RL split into two sub-columns right
			rlWide := renderRateLimitsHelpWide(s, keyStyle, descStyle)

			maxH := lipgloss.Height(keysPanel)
			if h := lipgloss.Height(rlWide); h > maxH {
				maxH = h
			}
			separator := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorBorder)).
				Render(strings.Repeat("│\n", maxH-1) + "│")

			modalContent = lipgloss.JoinHorizontal(lipgloss.Top,
				keysPanel, "   ", separator, "   ", rlWide,
			)
		}
	} else {
		modalContent = keysPanel
	}

	dismiss := s.Dim.Render("Press any key to close")
	content := lipgloss.JoinVertical(lipgloss.Left, modalContent, "", dismiss)

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorTitle)).
		Padding(1, 3).
		Render(content)

	// Extract header (first line) and footer (last line) from bg
	bgLines := strings.Split(bg, "\n")
	header := ""
	footer := ""
	if len(bgLines) > 0 {
		header = bgLines[0]
	}
	if len(bgLines) > 1 {
		footer = bgLines[len(bgLines)-1]
	}

	middleHeight := height - 2
	if middleHeight < 1 {
		middleHeight = 1
	}
	centered := lipgloss.Place(width, middleHeight, lipgloss.Center, lipgloss.Center, modal)

	return lipgloss.JoinVertical(lipgloss.Left, header, centered, footer)
}

// renderRateLimitsHelpTall renders rate limits info as a single column (for tall panes).
func renderRateLimitsHelpTall(s Styles, keyStyle, descStyle lipgloss.Style) string {
	dot := func(style lipgloss.Style) string { return style.Render("●") }

	rlTitle := s.Title.Render("RATE LIMITS")
	rlRows := []string{
		descStyle.Render("Usage is tracked across two rolling windows:"),
		"",
		keyStyle.Render("5h  ") + descStyle.Render("Short-term burst usage"),
		keyStyle.Render("7d  ") + descStyle.Render("Longer-term sustained usage"),
		"",
		descStyle.Render("Burn-rate indicator:"),
		dot(s.StatusOk) + descStyle.Render("  On track (usage ≤ elapsed + 5%)"),
		dot(s.StatusWarn) + descStyle.Render("  Elevated (usage > elapsed + 5%)"),
		dot(s.StatusErr) + descStyle.Render("  Hot (usage > elapsed + 15%)"),
		"",
		descStyle.Render("Data refreshes from the API every 10 minutes."),
	}

	return lipgloss.JoinVertical(lipgloss.Left, rlTitle, "", strings.Join(rlRows, "\n"))
}

// renderRateLimitsHelpWide renders rate limits info as two sub-columns (for short panes).
func renderRateLimitsHelpWide(s Styles, keyStyle, descStyle lipgloss.Style) string {
	dot := func(style lipgloss.Style) string { return style.Render("●") }

	rlTitle := s.Title.Render("RATE LIMITS")

	col1 := lipgloss.JoinVertical(lipgloss.Left,
		descStyle.Render("Rolling windows:"),
		"",
		keyStyle.Render("5h  ")+descStyle.Render("Short-term burst"),
		keyStyle.Render("7d  ")+descStyle.Render("Sustained usage"),
		"",
		descStyle.Render("Refreshes every 10 min."),
	)

	col2 := lipgloss.JoinVertical(lipgloss.Left,
		descStyle.Render("Burn-rate indicator:"),
		"",
		dot(s.StatusOk)+descStyle.Render("  On track (≤ elapsed + 5%)"),
		dot(s.StatusWarn)+descStyle.Render("  Elevated (> elapsed + 5%)"),
		dot(s.StatusErr)+descStyle.Render("  Hot (> elapsed + 15%)"),
	)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, col1, "    ", col2)

	return lipgloss.JoinVertical(lipgloss.Left, rlTitle, "", columns)
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
