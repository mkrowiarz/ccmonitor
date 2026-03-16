package tui

import "fmt"

// renderFooter renders the bottom help bar.
func renderFooter(s Styles, width int) string {
	text := fmt.Sprintf(" q quit  %s  r refresh  %s  tab/1-3 view", s.Dim.Render("\u00b7"), s.Dim.Render("\u00b7"))
	return s.Footer.Width(width).Render(text)
}
