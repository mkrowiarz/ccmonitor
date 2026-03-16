package tui

import tea "github.com/charmbracelet/bubbletea"

func handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "q", "ctrl+c":
		return tea.Quit
	case "r":
		return func() tea.Msg { return TickMsg{} }
	}
	return nil
}
