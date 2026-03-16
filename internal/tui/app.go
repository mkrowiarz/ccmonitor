package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/michal/ccmonitor/internal/backend"
	"github.com/michal/ccmonitor/internal/domain"
)

// Tab indices for the TUI views.
const (
	tabDashboard = iota
	tabActivity
	tabProcesses
	tabCount // sentinel for wrapping
)

// Model is the root Bubble Tea model for ccmonitor.
type Model struct {
	backend   backend.Backend
	snapshot  *domain.BackendSnapshot
	interval  time.Duration
	activeTab int
	width        int
	height       int
	styles       Styles
	err          error
}

// Options configures the TUI model.
type Options struct {
	Backend  backend.Backend
	Interval time.Duration
}

// NewModel creates a new TUI model with the given options.
func NewModel(opts Options) Model {
	return Model{
		backend:  opts.Backend,
		interval: opts.Interval,
		styles:   NewStyles(),
	}
}

// Init returns the initial commands for the model.
func (m Model) Init() tea.Cmd {
	return m.collectCmd()
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, tea.ClearScreen

	case TickMsg:
		return m, m.collectCmd()

	case SnapshotMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.snapshot = msg.Snapshot
			m.err = nil
		}
		return m, m.tickCmd()

	case TabMsg:
		if msg.Tab == -1 {
			m.activeTab = (m.activeTab + 1) % tabCount
		} else {
			m.activeTab = msg.Tab
		}
		return m, nil

	case tea.KeyMsg:
		cmd := handleKeyPress(msg)
		if cmd != nil {
			return m, cmd
		}
	}
	return m, nil
}

const (
	minWidth  = 100
	minHeight = 14
)

// View renders the full TUI layout.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.width < minWidth || m.height < minHeight {
		return renderTooSmall(m.styles, m.width, m.height)
	}

	// Determine interval in seconds for header display
	intervalSec := int(m.interval.Seconds())

	// Render header and footer
	header := renderHeader(m.styles, m.snapshot, intervalSec, m.activeTab, m.width)
	footer := renderFooter(m.styles, m.width)

	// Available height for panels (subtract header + footer lines)
	panelHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if panelHeight < 4 {
		panelHeight = 4
	}

	// Extract data from snapshot
	var usage *domain.UsageSummary
	var sessions []domain.ActiveSession
	var events []domain.RecentEvent
	var rateLimits domain.RateLimits
	if m.snapshot != nil {
		usage = &m.snapshot.Usage
		sessions = m.snapshot.ActiveSessions
		events = m.snapshot.RecentEvents
		rateLimits = m.snapshot.RateLimits
	}

	var body string
	switch m.activeTab {
	case tabActivity:
		body = renderActivityView(m.styles, usage, events, m.width, panelHeight)
	case tabProcesses:
		body = renderProcessesView(m.styles, sessions, m.width, panelHeight)
	default:
		body = m.renderDashboard(usage, sessions, rateLimits, panelHeight)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// renderDashboard renders the horizontal panel layout.
func (m Model) renderDashboard(usage *domain.UsageSummary, sessions []domain.ActiveSession, rateLimits domain.RateLimits, panelHeight int) string {
	colW := m.width * 22 / 100
	if colW < 22 {
		colW = 22
	}
	sessW := m.width - colW*3

	todayPanel := renderTodayPanel(m.styles, usage, colW, panelHeight)
	lifetimePanel := renderLifetimePanel(m.styles, usage, colW, panelHeight)
	rateLimitsPanel := renderRateLimitsPanel(m.styles, rateLimits, colW, panelHeight)
	sessionsPanel := renderSessionsPanel(m.styles, sessions, sessW, panelHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, todayPanel, lifetimePanel, rateLimitsPanel, sessionsPanel)
}

// collectCmd creates a command that collects a snapshot from the backend.
func (m Model) collectCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		opts := backend.CollectOpts{
			IncludeRecentActivity: true,
		}

		snapshot, err := m.backend.Collect(ctx, opts)
		return SnapshotMsg{
			Snapshot: snapshot,
			Err:      err,
		}
	}
}

// tickCmd creates a command that waits for the configured interval then sends a TickMsg.
func (m Model) tickCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(m.interval)
		return TickMsg{}
	}
}

// renderTooSmall renders a btop-style warning when the terminal is too small.
func renderTooSmall(s Styles, w, h int) string {
	title := s.Title.Render("Terminal size too small:")
	current := fmt.Sprintf("Width = %s  Height = %s",
		s.StatusErr.Render(fmt.Sprintf("%d", w)),
		s.StatusErr.Render(fmt.Sprintf("%d", h)),
	)
	needed := fmt.Sprintf("Width = %s  Height = %s",
		s.StatusOk.Render(fmt.Sprintf("%d", minWidth)),
		s.StatusOk.Render(fmt.Sprintf("%d", minHeight)),
	)
	msg := lipgloss.JoinVertical(lipgloss.Center,
		"",
		title,
		current,
		"",
		s.Label.Render("Needed:"),
		needed,
		"",
	)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, msg)
}
