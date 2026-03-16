package tui

import tea "github.com/charmbracelet/bubbletea"

// TabMsg signals a tab switch.
type TabMsg struct{ Tab int }

func handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "q", "ctrl+c":
		return tea.Quit
	case "r":
		return func() tea.Msg { return TickMsg{} }
	case "tab":
		return func() tea.Msg { return TabMsg{Tab: -1} } // -1 = cycle
	case "1":
		return func() tea.Msg { return TabMsg{Tab: tabDashboard} }
	case "2":
		return func() tea.Msg { return TabMsg{Tab: tabActivity} }
	}
	return nil
}
