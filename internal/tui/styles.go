package tui

import "github.com/charmbracelet/lipgloss"

// Colors (exported for use in CLI help output).
const (
	ColorBg       = "#1a1b2e"
	ColorPanel    = "#232436"
	ColorBorder   = "#444466"
	ColorTitle    = "#A6E3A1" // green
	ColorModel    = "#CBA6F7" // purple/magenta
	ColorValue    = "#FFFFFF"
	ColorLabel    = "#888899"
	ColorDim      = "#666677"
	ColorOk       = "#A6E3A1"
	ColorDegraded = "#F9E2AF" // yellow
	ColorError    = "#F38BA8" // red
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
			Background(lipgloss.Color(ColorBg)),

		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitle)).
			Bold(true),

		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLabel)),

		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorValue)).
			Bold(true),

		ModelName: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorModel)),

		StatusOk: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorOk)),

		StatusWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDegraded)),

		StatusErr: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)),

		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitle)).
			Bold(true),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDim)),

		Dim: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDim)),

		TableHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLabel)).
			Bold(true),
	}
}
