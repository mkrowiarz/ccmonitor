package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/sparkline"
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

// renderActivityView renders the Activity tab: sparkline + recent messages.
func renderActivityView(s Styles, usage *domain.UsageSummary, events []domain.RecentEvent, width, height int) string {
	// --- Left side: sparkline chart ---
	sparkW := width / 3
	if sparkW < 24 {
		sparkW = 24
	}
	sparkInner := sparkW - panelOverhead
	var sparkLines []string
	sparkLines = append(sparkLines, s.Title.Render("MESSAGES PER DAY"))

	if usage != nil && len(usage.DailyActivity) >= 2 {
		first := usage.DailyActivity[0].Date
		last := usage.DailyActivity[len(usage.DailyActivity)-1].Date
		sparkLines = append(sparkLines, s.Dim.Render(first+" → "+last))

		sparkH := height - 2 - len(sparkLines) - 1
		if sparkH < 1 {
			sparkH = 1
		}

		data := make([]float64, len(usage.DailyActivity))
		for i, da := range usage.DailyActivity {
			data[i] = float64(da.MessageCount)
		}

		sl := sparkline.New(sparkInner, sparkH,
			sparkline.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(colorOk))),
		)
		sl.PushAll(data)
		sl.Draw()
		sparkLines = append(sparkLines, sl.View())
	} else {
		sparkLines = append(sparkLines, s.Dim.Render("Not enough data"))
	}

	sparkContent := strings.Join(sparkLines, "\n")
	sparkPanel := s.Panel.Width(sparkW - 2).Height(height - 2).Render(sparkContent)

	// --- Right side: recent messages ---
	recentW := width - sparkW
	var recentLines []string
	recentLines = append(recentLines, s.Title.Render("RECENT"))

	if len(events) == 0 {
		recentLines = append(recentLines, s.Dim.Render("No recent activity"))
	} else {
		maxEvents := height - 4 // border + title + padding
		if maxEvents < 1 {
			maxEvents = 1
		}
		recentInner := recentW - panelOverhead
		for i, ev := range events {
			if i >= maxEvents {
				break
			}
			timeStr := s.Dim.Render(ev.Timestamp.Format("15:04"))
			proj := s.Label.Render(truncate(ev.ProjectName, 14))
			remaining := recentInner - 6 - 15 // time width + proj width approx
			if remaining < 10 {
				remaining = 10
			}
			display := truncate(ev.Display, remaining)
			recentLines = append(recentLines, fmt.Sprintf("%s %s %s", timeStr, proj, display))
		}
	}

	recentContent := strings.Join(recentLines, "\n")
	recentPanel := s.Panel.Width(recentW - 2).Height(height - 2).Render(recentContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, sparkPanel, recentPanel)
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

// renderRateLimitsPanel renders the rate limit windows panel.
func renderRateLimitsPanel(s Styles, limits domain.RateLimits, width, height int) string {
	inner := width - panelOverhead
	var lines []string

	lines = append(lines, s.Title.Render("RATE LIMITS"))

	if limits.FiveHour == nil && limits.SevenDay == nil {
		lines = append(lines, s.Dim.Render("No rate limit data"))
		content := strings.Join(lines, "\n")
		return s.Panel.Width(width - 2).Height(height - 2).Render(content)
	}

	now := time.Now()

	if w := limits.FiveHour; w != nil {
		lines = append(lines, "")
		lines = append(lines, s.Label.Render("5-HOUR WINDOW"))
		lines = append(lines, renderWindowDetail(s, w, now, 5*time.Hour, inner))
	}

	if w := limits.SevenDay; w != nil {
		lines = append(lines, "")
		lines = append(lines, s.Label.Render("7-DAY WINDOW"))
		lines = append(lines, renderWindowDetail(s, w, now, 7*24*time.Hour, inner))
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderWindowDetail renders utilization, time remaining, and burn-rate for a window.
func renderWindowDetail(s Styles, w *domain.RateWindow, now time.Time, windowDur time.Duration, inner int) string {
	var lines []string

	// Utilization with color
	util := w.Utilization
	utilStyle := s.StatusOk
	if util >= 80 {
		utilStyle = s.StatusErr
	} else if util >= 50 {
		utilStyle = s.StatusWarn
	}

	// Progress bar
	barWidth := inner - 8 // space for "100.0% "
	if barWidth < 5 {
		barWidth = 5
	}
	filled := int(math.Round(util / 100 * float64(barWidth)))
	if filled > barWidth {
		filled = barWidth
	}
	bar := utilStyle.Render(strings.Repeat("█", filled)) + s.Dim.Render(strings.Repeat("░", barWidth-filled))
	lines = append(lines, fmt.Sprintf("%s %s", utilStyle.Render(fmt.Sprintf("%5.1f%%", util)), bar))

	// Time remaining
	remaining := time.Until(w.ResetsAt)
	if remaining < 0 {
		remaining = 0
	}
	lines = append(lines, formatKV(s, "Resets in", format.FormatUptime(remaining), inner))

	// Burn rate indicator
	elapsed := windowDur - remaining
	if elapsed < 0 {
		elapsed = 0
	}
	elapsedPct := 0.0
	if windowDur > 0 {
		elapsedPct = float64(elapsed) / float64(windowDur) * 100
	}
	diff := util - elapsedPct
	burnLabel := "●"
	burnStyle := s.StatusOk
	if diff > 15 {
		burnStyle = s.StatusErr
		burnLabel = "● HIGH"
	} else if diff > 5 {
		burnStyle = s.StatusWarn
		burnLabel = "● ELEVATED"
	}
	lines = append(lines, formatKV(s, "Burn rate", burnStyle.Render(burnLabel), inner))

	return strings.Join(lines, "\n")
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
