package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/michal/ccmonitor/internal/domain"
	"github.com/michal/ccmonitor/internal/format"
)

// renderTodayPanel renders the "Today" usage panel (top-left).
func renderTodayPanel(s Styles, usage *domain.UsageSummary, width, height int) string {
	var lines []string

	lines = append(lines, s.Title.Render("TODAY"))

	if usage != nil {
		lines = append(lines, formatKV(s, "Messages", formatOptionalCount(usage.TodayMessages), width-6))
		lines = append(lines, formatKV(s, "Sessions", formatOptionalCount(usage.TodaySessions), width-6))
	} else {
		lines = append(lines, formatKV(s, "Messages", "-", width-6))
		lines = append(lines, formatKV(s, "Sessions", "-", width-6))
	}

	lines = append(lines, "")
	lines = append(lines, s.Title.Render("TOKENS TODAY"))

	if usage != nil && len(usage.TodayTokens) > 0 {
		for _, mt := range usage.TodayTokens {
			name := stripClaudePrefix(mt.ModelName)
			lines = append(lines, formatModelKV(s, name, format.FormatCount(mt.TokenCount), width-6))
		}
	} else {
		lines = append(lines, s.Dim.Render("No token data"))
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderLifetimePanel renders the "Lifetime" usage panel (bottom-left).
func renderLifetimePanel(s Styles, usage *domain.UsageSummary, width, height int) string {
	var lines []string

	lines = append(lines, s.Title.Render("LIFETIME"))

	if usage != nil {
		lines = append(lines, formatKV(s, "Messages", formatOptionalCount(usage.LifetimeMessages), width-6))
		lines = append(lines, formatKV(s, "Sessions", formatOptionalCount(usage.LifetimeSessions), width-6))
	} else {
		lines = append(lines, formatKV(s, "Messages", "-", width-6))
		lines = append(lines, formatKV(s, "Sessions", "-", width-6))
	}

	lines = append(lines, "")
	lines = append(lines, s.Title.Render("TOKENS LIFETIME"))

	if usage != nil && len(usage.LifetimeTokens) > 0 {
		for _, mt := range usage.LifetimeTokens {
			name := stripClaudePrefix(mt.ModelName)
			lines = append(lines, formatModelKV(s, name, format.FormatCount(mt.TokenCount), width-6))
		}
	} else {
		lines = append(lines, s.Dim.Render("No token data"))
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderSessionsPanel renders the "Active Sessions" panel (top-right).
func renderSessionsPanel(s Styles, sessions []domain.ActiveSession, width, height int) string {
	var lines []string

	lines = append(lines, s.Title.Render("ACTIVE SESSIONS"))

	count := len(sessions)
	if count == 0 {
		lines = append(lines, "")
		lines = append(lines, s.Dim.Render("No active sessions"))
	} else {
		countLabel := "instance"
		if count > 1 {
			countLabel = "instances"
		}
		lines = append(lines, s.StatusOk.Render(fmt.Sprintf("\u25cf %d active %s", count, countLabel)))
		lines = append(lines, "")

		// Table header
		innerWidth := width - 6
		header := formatSessionRow(s.TableHeader, "project", "pid", "cpu", "mem", "uptime", innerWidth)
		lines = append(lines, header)

		for _, sess := range sessions {
			proj := truncate(sess.ProjectName, 14)
			pid := fmt.Sprintf("%d", sess.PID)
			cpu := fmt.Sprintf("%.1f%%", sess.CPUPercent)
			mem := fmt.Sprintf("%.1f%%", sess.MemPercent)
			uptime := format.FormatUptime(sess.Uptime)
			lines = append(lines, formatSessionRow(lipgloss.NewStyle().Foreground(lipgloss.Color(colorValue)), proj, pid, cpu, mem, uptime, innerWidth))
		}
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderActivityPanel renders the "Recent Activity" panel (bottom-right).
func renderActivityPanel(s Styles, events []domain.RecentEvent, width, height int) string {
	var lines []string

	lines = append(lines, s.Title.Render("RECENT ACTIVITY"))

	if len(events) == 0 {
		lines = append(lines, "")
		lines = append(lines, s.Dim.Render("No recent activity"))
	} else {
		// Calculate how many events fit in the panel
		maxEvents := height - 4 // borders + padding + title
		if maxEvents < 1 {
			maxEvents = 1
		}
		for i, ev := range events {
			if i >= maxEvents {
				break
			}
			timeStr := s.Dim.Render(ev.Timestamp.Format("15:04"))
			proj := s.Label.Render(truncate(ev.ProjectName, 14))
			display := truncate(ev.Display, width-30)
			lines = append(lines, fmt.Sprintf("%s %s %s", timeStr, proj, display))
		}
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// --- helpers ---

// formatKV formats a label-value line with the value right-aligned.
func formatKV(s Styles, label, value string, innerWidth int) string {
	labelRendered := s.Label.Render(label)
	valueRendered := s.Value.Render(value)
	labelW := lipgloss.Width(labelRendered)
	valueW := lipgloss.Width(valueRendered)
	gap := innerWidth - labelW - valueW
	if gap < 1 {
		gap = 1
	}
	return labelRendered + strings.Repeat(" ", gap) + valueRendered
}

// formatModelKV formats a model-name + value line with purple model name.
func formatModelKV(s Styles, model, value string, innerWidth int) string {
	modelRendered := s.ModelName.Render(model)
	valueRendered := s.Value.Render(value)
	modelW := lipgloss.Width(modelRendered)
	valueW := lipgloss.Width(valueRendered)
	gap := innerWidth - modelW - valueW
	if gap < 1 {
		gap = 1
	}
	return modelRendered + strings.Repeat(" ", gap) + valueRendered
}

// formatSessionRow formats a row for the sessions table with fixed columns.
func formatSessionRow(style lipgloss.Style, project, pid, cpu, mem, uptime string, innerWidth int) string {
	// Fixed column widths
	colProject := 14
	colPID := 8
	colCPU := 8
	colMem := 10
	// uptime gets the rest

	row := fmt.Sprintf("%-*s %-*s %-*s %-*s %s",
		colProject, project,
		colPID, pid,
		colCPU, cpu,
		colMem, mem,
		uptime,
	)
	return style.Render(row)
}

// formatOptionalCount formats an optional int64 pointer using FormatCount.
func formatOptionalCount(v *int64) string {
	if v == nil {
		return "-"
	}
	return format.FormatCount(*v)
}

// stripClaudePrefix removes the "claude-" prefix and date suffixes from model names for brevity.
// e.g. "claude-sonnet-4-5-20250929" → "sonnet-4-5"
func stripClaudePrefix(name string) string {
	name = strings.TrimPrefix(name, "claude-")
	// Strip date suffixes like "-20250929"
	if len(name) > 9 && name[len(name)-9] == '-' {
		suffix := name[len(name)-8:]
		allDigits := true
		for _, c := range suffix {
			if c < '0' || c > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			name = name[:len(name)-9]
		}
	}
	return name
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if maxLen < 1 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "\u2026"
	}
	return string(runes[:maxLen-1]) + "\u2026"
}
