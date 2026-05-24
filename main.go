package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mkrowiarz/ccmonitor/internal/backend"
	"github.com/mkrowiarz/ccmonitor/internal/claude"
	"github.com/mkrowiarz/ccmonitor/internal/selfupdate"
	"github.com/mkrowiarz/ccmonitor/internal/tui"
	"github.com/mkrowiarz/ccmonitor/internal/waybar"
)

var version = ""

func init() {
	if version != "" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	} else {
		version = "dev"
	}
}

func printUsage() {
	title := lipgloss.NewStyle().Foreground(lipgloss.Color(tui.ColorTitle)).Bold(true)
	flagName := lipgloss.NewStyle().Foreground(lipgloss.Color(tui.ColorModel))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color(tui.ColorLabel))
	def := lipgloss.NewStyle().Foreground(lipgloss.Color(tui.ColorDim))

	fmt.Println()
	fmt.Println(title.Render("CLAUDE MONITOR") + " " + def.Render(version))
	fmt.Println(desc.Render("Terminal dashboard for Claude Code usage, sessions, and rate limits."))
	fmt.Println()
	fmt.Println(title.Render("USAGE"))
	fmt.Println("  " + flagName.Render("ccmonitor") + " " + desc.Render("[flags]"))
	fmt.Println("  " + flagName.Render("ccmonitor") + " " + desc.Render("<command>"))
	fmt.Println()
	fmt.Println(title.Render("COMMANDS"))
	fmt.Println("  " + flagName.Render(fmt.Sprintf("%-20s", "waybar-setup")) + desc.Render("Print Waybar module setup instructions and exit"))
	fmt.Println("  " + flagName.Render(fmt.Sprintf("%-20s", "update")) + desc.Render("Update ccmonitor to the latest release"))
	fmt.Println()
	fmt.Println(title.Render("FLAGS"))

	flags := []struct {
		name, defVal, description string
	}{
		{"-interval N", "10", "Refresh interval in seconds"},
		{"-no-rate-limits", "", "Disable the rate limits panel"},
		{"-minimal", "", "Dashboard only, no activity/analytics tabs"},
		{"-waybar", "", "Print one Waybar JSON line for rate limits and exit"},
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
	// Subcommands are matched before flag parsing.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "waybar-setup":
			exe, err := os.Executable()
			if err != nil {
				exe = "ccmonitor"
			}
			fmt.Print(waybar.SetupText(exe))
			return
		case "update":
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			if err := selfupdate.Update(ctx, version, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	interval := flag.Int("interval", 10, "refresh interval in seconds")
	backendName := flag.String("backend", "claude", "backend to use")
	noRateLimits := flag.Bool("no-rate-limits", false, "disable the rate limits panel")
	minimal := flag.Bool("minimal", false, "dashboard only, no activity/analytics tabs")
	waybarOut := flag.Bool("waybar", false, "print one Waybar JSON line for rate limits and exit")
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

	// Waybar mode: print one JSON line for rate limits and exit (no TUI).
	if *waybarOut {
		rp, ok := b.(backend.RateLimitProvider)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: backend %q does not provide rate limits\n", b.Name())
			os.Exit(1)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		rl, enabled := rp.RateLimits(ctx)
		line, err := waybar.Render(rl, enabled)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(line)
		return
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
