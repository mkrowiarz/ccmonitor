package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/michal/ccmonitor/internal/backend"
	"github.com/michal/ccmonitor/internal/domain"
)

// Model is the root Bubble Tea model for ccmonitor.
type Model struct {
	backend      backend.Backend
	snapshot     *domain.BackendSnapshot
	interval     time.Duration
	showActivity bool
	width        int
	height       int
	styles       Styles
	err          error
}

// Options configures the TUI model.
type Options struct {
	Backend      backend.Backend
	Interval     time.Duration
	ShowActivity bool
}

// NewModel creates a new TUI model with the given options.
func NewModel(opts Options) Model {
	return Model{
		backend:      opts.Backend,
		interval:     opts.Interval,
		showActivity: opts.ShowActivity,
		styles:       NewStyles(),
	}
}

// Init returns the initial commands for the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.collectCmd(),
		m.tickCmd(),
	)
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		return m, m.collectCmd()

	case SnapshotMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.snapshot = msg.Snapshot
			m.err = nil
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

// View renders the full TUI layout.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Determine interval in seconds for header display
	intervalSec := int(m.interval.Seconds())

	// Render header and footer
	header := renderHeader(m.styles, m.snapshot, intervalSec, m.width)
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
	if m.snapshot != nil {
		usage = &m.snapshot.Usage
		sessions = m.snapshot.ActiveSessions
		events = m.snapshot.RecentEvents
	}

	// Horizontal single-row layout: today | lifetime | sessions [| activity]
	// Left two panels get fixed ~25% each, sessions/activity gets the rest
	leftW := m.width / 4
	if leftW < 24 {
		leftW = 24
	}
	remainW := m.width - leftW*2
	sessW := remainW
	actW := 0
	if m.showActivity {
		sessW = remainW / 2
		actW = remainW - sessW
	}

	todayPanel := renderTodayPanel(m.styles, usage, leftW, panelHeight)
	lifetimePanel := renderLifetimePanel(m.styles, usage, leftW, panelHeight)
	sessionsPanel := renderSessionsPanel(m.styles, sessions, sessW, panelHeight)

	panels := []string{todayPanel, lifetimePanel, sessionsPanel}
	if m.showActivity {
		activityPanel := renderActivityPanel(m.styles, events, actW, panelHeight)
		panels = append(panels, activityPanel)
	}

	grid := lipgloss.JoinHorizontal(lipgloss.Top, panels...)

	return lipgloss.JoinVertical(lipgloss.Left, header, grid, footer)
}

// collectCmd creates a command that collects a snapshot from the backend.
func (m Model) collectCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		opts := backend.CollectOpts{
			IncludeRecentActivity: m.showActivity,
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
