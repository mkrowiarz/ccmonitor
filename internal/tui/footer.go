package tui

import "fmt"

// renderFooter renders the bottom help bar.
func renderFooter(s Styles, width int, minimal bool) string {
	dot := s.Dim.Render("\u00b7")
	if minimal {
		text := fmt.Sprintf(" q quit  %s  r refresh  %s  ? help", dot, dot)
		return s.Footer.Width(width).Render(text)
	}
	text := fmt.Sprintf(" q quit  %s  r refresh  %s  tab cycle tabs  %s  [1] dashboard  [2] activity  [3] analytics  %s  ? help", dot, dot, dot, dot)
	return s.Footer.Width(width).Render(text)
}
