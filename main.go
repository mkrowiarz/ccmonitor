package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mkrowiarz/ccmonitor/internal/backend"
	"github.com/mkrowiarz/ccmonitor/internal/claude"
	"github.com/mkrowiarz/ccmonitor/internal/tui"
)

var version = "dev"

func printUsage() {
	title := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Bold(true)
	flagName := lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888899"))
	def := lipgloss.NewStyle().Foreground(lipgloss.Color("#666677"))

	fmt.Println()
	fmt.Println(title.Render("CLAUDE MONITOR") + " " + def.Render(version))
	fmt.Println(desc.Render("Terminal dashboard for Claude Code usage, sessions, and rate limits."))
	fmt.Println()
	fmt.Println(title.Render("USAGE"))
	fmt.Println("  " + flagName.Render("ccmonitor") + " " + desc.Render("[flags]"))
	fmt.Println()
	fmt.Println(title.Render("FLAGS"))

	flags := []struct {
		name, defVal, description string
	}{
		{"-interval N", "10", "Refresh interval in seconds"},
		{"-no-rate-limits", "", "Disable the rate limits panel"},
		{"-minimal", "", "Dashboard only, no activity/analytics tabs"},
		{"-backend NAME", "claude", "Backend to use"},
		{"-version", "", "Print version and exit"},
	}

	for _, f := range flags {
		line := "  " + flagName.Render(fmt.Sprintf("%-20s", f.name))
		line += desc.Render(f.description)
		if f.defVal != "" {
			line += " " + def.Render(fmt.Sprintf("(default: %s)", f.defVal))
		}
		fmt.Println(line)
	}
	fmt.Println()
}

func main() {
	interval := flag.Int("interval", 10, "refresh interval in seconds")
	backendName := flag.String("backend", "claude", "backend to use")
	noRateLimits := flag.Bool("no-rate-limits", false, "disable the rate limits panel")
	minimal := flag.Bool("minimal", false, "dashboard only, no activity/analytics tabs")
	showVersion := flag.Bool("version", false, "print version and exit")

	flag.Usage = printUsage
	flag.Parse()

	if *showVersion {
		fmt.Println("ccmonitor", version)
		return
	}

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
		NoRateLimits: *noRateLimits,
		Minimal:      *minimal,
	})

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
