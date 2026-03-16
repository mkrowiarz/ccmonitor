package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/michal/ccmonitor/internal/domain"
	"github.com/michal/ccmonitor/internal/format"
)

// panelOverhead is the horizontal space consumed by border (2) + padding (2).
const panelOverhead = 4

// renderTodayPanel renders the "Today" usage panel.
func renderTodayPanel(s Styles, usage *domain.UsageSummary, width, height int) string {
	inner := width - panelOverhead
	var lines []string

	lines = append(lines, s.Title.Render("TODAY"))

	if usage != nil {
		lines = append(lines, formatKV(s, "Messages", formatOptionalCount(usage.TodayMessages), inner))
		lines = append(lines, formatKV(s, "Sessions", formatOptionalCount(usage.TodaySessions), inner))
	} else {
		lines = append(lines, formatKV(s, "Messages", "-", inner))
		lines = append(lines, formatKV(s, "Sessions", "-", inner))
	}

	lines = append(lines, "")
	lines = append(lines, s.Title.Render("TOKENS TODAY"))

	if usage != nil && len(usage.TodayTokens) > 0 {
		for _, mt := range usage.TodayTokens {
			name := stripClaudePrefix(mt.ModelName)
			lines = append(lines, formatModelKV(s, name, format.FormatCount(mt.TokenCount), inner))
		}
	} else {
		lines = append(lines, s.Dim.Render("No token data"))
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderLifetimePanel renders the "Lifetime" usage panel.
func renderLifetimePanel(s Styles, usage *domain.UsageSummary, width, height int) string {
	inner := width - panelOverhead
	var lines []string

	lines = append(lines, s.Title.Render("LIFETIME"))

	if usage != nil {
		lines = append(lines, formatKV(s, "Messages", formatOptionalCount(usage.LifetimeMessages), inner))
		lines = append(lines, formatKV(s, "Sessions", formatOptionalCount(usage.LifetimeSessions), inner))
	} else {
		lines = append(lines, formatKV(s, "Messages", "-", inner))
		lines = append(lines, formatKV(s, "Sessions", "-", inner))
	}

	lines = append(lines, "")
	lines = append(lines, s.Title.Render("TOKENS LIFETIME"))

	if usage != nil && len(usage.LifetimeTokens) > 0 {
		for _, mt := range usage.LifetimeTokens {
			name := stripClaudePrefix(mt.ModelName)
			lines = append(lines, formatModelKV(s, name, format.FormatCount(mt.TokenCount), inner))
		}
	} else {
		lines = append(lines, s.Dim.Render("No token data"))
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderSessionsPanel renders the "Active Sessions" panel.
func renderSessionsPanel(s Styles, sessions []domain.ActiveSession, width, height int) string {
	var lines []string

	lines = append(lines, s.Title.Render("ACTIVE SESSIONS"))

	count := len(sessions)
	if count == 0 {
		lines = append(lines, s.Dim.Render("No active sessions"))
	} else {
		countLabel := "instance"
		if count > 1 {
			countLabel = "instances"
		}
		lines = append(lines, s.StatusOk.Render(fmt.Sprintf("● %d active %s", count, countLabel)))

		// Table header
		innerWidth := width - panelOverhead
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

// renderActivityPanel renders the "Recent Activity" panel.
func renderActivityPanel(s Styles, events []domain.RecentEvent, width, height int) string {
	var lines []string

	lines = append(lines, s.Title.Render("RECENT ACTIVITY"))

	if len(events) == 0 {
		lines = append(lines, s.Dim.Render("No recent activity"))
	} else {
		maxEvents := height - 3 // border + title
		if maxEvents < 1 {
			maxEvents = 1
		}
		for i, ev := range events {
			if i >= maxEvents {
				break
			}
			timeStr := s.Dim.Render(ev.Timestamp.Format("15:04"))
			proj := s.Label.Render(truncate(ev.ProjectName, 14))
			display := truncate(ev.Display, width-28)
			lines = append(lines, fmt.Sprintf("%s %s %s", timeStr, proj, display))
		}
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// --- helpers ---

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

func formatSessionRow(style lipgloss.Style, project, pid, cpu, mem, uptime string, innerWidth int) string {
	colProject := 14
	colPID := 8
	colCPU := 8
	colMem := 10

	row := fmt.Sprintf("%-*s %-*s %-*s %-*s %s",
		colProject, project,
		colPID, pid,
		colCPU, cpu,
		colMem, mem,
		uptime,
	)
	return style.Render(row)
}

func formatOptionalCount(v *int64) string {
	if v == nil {
		return "-"
	}
	return format.FormatCount(*v)
}

func stripClaudePrefix(name string) string {
	name = strings.TrimPrefix(name, "claude-")
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

func truncate(s string, maxLen int) string {
	if maxLen < 1 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}
