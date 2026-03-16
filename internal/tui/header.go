package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/michal/ccmonitor/internal/domain"
)

// renderHeader renders the top bar showing app title, tabs, status, refresh interval, and time.
func renderHeader(s Styles, snapshot *domain.BackendSnapshot, interval int, activeTab int, width int) string {
	tabNames := []string{"Dashboard", "Activity"}
	var tabs []string
	for i, name := range tabNames {
		label := fmt.Sprintf(" %d:%s ", i+1, name)
		if i == activeTab {
			tabs = append(tabs, s.Header.Render(label))
		} else {
			tabs = append(tabs, s.Dim.Render(label))
		}
	}
	left := s.Header.Render("CLAUDE MONITOR") + " " + strings.Join(tabs, "")

	// Determine status
	status := domain.StatusOk
	statusTime := time.Now()
	if snapshot != nil {
		status = snapshot.Status
		statusTime = snapshot.CollectedAt
	}

	var dotStyle lipgloss.Style
	var statusLabel string
	switch status {
	case domain.StatusOk:
		dotStyle = s.StatusOk
		statusLabel = "ok"
	case domain.StatusDegraded:
		dotStyle = s.StatusWarn
		statusLabel = "degraded"
	case domain.StatusUnavailable:
		dotStyle = s.StatusErr
		statusLabel = "unavailable"
	default:
		dotStyle = s.Dim
		statusLabel = "unknown"
	}

	dot := dotStyle.Render("\u25cf")
	statusText := dotStyle.Render(statusLabel)
	refreshText := s.Dim.Render(fmt.Sprintf("refresh %ds", interval))
	timeText := s.Dim.Render(statusTime.Format("2006-01-02 15:04"))

	right := fmt.Sprintf("%s %s  %s  %s", dot, statusText, refreshText, timeText)

	// Calculate spacing (account for 1-char left padding + 1-char right padding)
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := width - leftWidth - rightWidth - 2
	if gap < 1 {
		gap = 1
	}

	return fmt.Sprintf(" %s%s%s ", left, strings.Repeat(" ", gap), right)
}
