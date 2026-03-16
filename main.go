package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/michal/ccmonitor/internal/backend"
	"github.com/michal/ccmonitor/internal/claude"
	"github.com/michal/ccmonitor/internal/tui"
)

func main() {
	interval := flag.Int("interval", 5, "refresh interval in seconds")
	recentActivity := flag.Bool("recent-activity", false, "show recent activity panel")
	backendName := flag.String("backend", "claude", "backend to use")
	flag.Parse()

	// Register backends
	backend.Register(claude.New())

	// Get selected backend
	b, err := backend.Get(*backendName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nAvailable backends: %v\n", err, backend.List())
		os.Exit(1)
	}

	// Create and run Bubble Tea program
	model := tui.NewModel(tui.Options{
		Backend:      b,
		Interval:     time.Duration(*interval) * time.Second,
		ShowActivity: *recentActivity,
	})

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
