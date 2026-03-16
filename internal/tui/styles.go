package tui

import "github.com/charmbracelet/lipgloss"

// Colors
const (
	colorBg       = "#1a1b2e"
	colorPanel    = "#232436"
	colorBorder   = "#444466"
	colorTitle    = "#A6E3A1" // green
	colorModel    = "#CBA6F7" // purple/magenta
	colorValue    = "#FFFFFF"
	colorLabel    = "#888899"
	colorDim      = "#666677"
	colorOk       = "#A6E3A1"
	colorDegraded = "#F9E2AF" // yellow
	colorError    = "#F38BA8" // red
)

// Styles holds all lipgloss styles used by the TUI.
type Styles struct {
	App         lipgloss.Style
	Panel       lipgloss.Style
	Title       lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	ModelName   lipgloss.Style
	StatusOk    lipgloss.Style
	StatusWarn  lipgloss.Style
	StatusErr   lipgloss.Style
	Header      lipgloss.Style
	Footer      lipgloss.Style
	Dim         lipgloss.Style
	TableHeader lipgloss.Style
}

// NewStyles creates the default style set.
func NewStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Background(lipgloss.Color(colorBg)),

		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTitle)).
			Bold(true),

		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorLabel)),

		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorValue)).
			Bold(true),

		ModelName: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorModel)),

		StatusOk: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorOk)),

		StatusWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDegraded)),

		StatusErr: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorError)),

		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTitle)).
			Bold(true),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim)),

		Dim: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim)),

		TableHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorLabel)).
			Bold(true),
	}
}
