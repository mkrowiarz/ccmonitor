package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/charmbracelet/lipgloss"
	"github.com/mkrowiarz/ccmonitor/internal/domain"
	"github.com/mkrowiarz/ccmonitor/internal/format"
)

// panelOverhead is the horizontal space consumed by border (2) + padding (2).
const panelOverhead = 4

// maxModels is the maximum number of model rows shown before truncating.
const maxModels = 3

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

	var todayTokens []domain.ModelTokens
	if usage != nil {
		todayTokens = usage.TodayTokens
	}
	lines = append(lines, renderModelTokens(s, todayTokens, inner, maxModels)...)

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

	var lifetimeTokens []domain.ModelTokens
	if usage != nil {
		lifetimeTokens = usage.LifetimeTokens
	}
	lines = append(lines, renderModelTokens(s, lifetimeTokens, inner, maxModels)...)

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// Model colors for bar chart segments.
var modelColors = []string{colorOk, colorModel, colorDegraded, colorError}

// renderActivityTab renders the Activity tab: recent prompts + processes.
func renderActivityTab(s Styles, sessions []domain.ActiveSession, events []domain.RecentEvent, width, height int) string {
	halfW := width / 2
	rightW := width - halfW

	recentPanel := renderRecentPanel(s, events, halfW, height)
	processPanel := renderProcessesView(s, sessions, rightW, height)

	return lipgloss.JoinHorizontal(lipgloss.Top, recentPanel, processPanel)
}

// renderRecentPanel renders the recent activity list.
func renderRecentPanel(s Styles, events []domain.RecentEvent, width, height int) string {
	inner := width - panelOverhead - 2
	var lines []string
	lines = append(lines, s.Title.Render("RECENT"))

	if len(events) == 0 {
		lines = append(lines, s.Dim.Render("No recent activity"))
	} else {
		availLines := height - panelOverhead - 1 // subtract title line
		maxEvents := availLines - 1              // reserve 1 for potential "+N more"
		if maxEvents < 1 {
			maxEvents = 1
		}

		visible := events
		truncatedCount := 0
		if len(visible) > maxEvents {
			truncatedCount = len(visible) - maxEvents
			visible = visible[:maxEvents]
		}

		for _, ev := range visible {
			timeStr := s.Dim.Render(ev.Timestamp.Format("15:04"))
			proj := s.Label.Render(truncate(ev.ProjectName, 14))
			used := lipgloss.Width(timeStr) + 1 + lipgloss.Width(proj) + 1
			remaining := inner - used
			if remaining < 10 {
				remaining = 10
			}
			display := truncate(ev.Display, remaining)
			lines = append(lines, fmt.Sprintf("%s %s %s", timeStr, proj, display))
		}

		if truncatedCount > 0 {
			lines = append(lines, s.Dim.Render(fmt.Sprintf("  +%d more", truncatedCount)))
		}
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderAnalyticsTab renders the Analytics tab: sparkline + token bar chart.
func renderAnalyticsTab(s Styles, usage *domain.UsageSummary, width, height int) string {
	halfW := width / 2
	rightW := width - halfW

	sparkPanel := renderMessagesBarPanel(s, usage, halfW, height)
	barPanel := renderTokenBarPanel(s, usage, rightW, height)

	return lipgloss.JoinHorizontal(lipgloss.Top, sparkPanel, barPanel)
}

// renderMessagesBarPanel renders a bar chart of messages per day.
func renderMessagesBarPanel(s Styles, usage *domain.UsageSummary, width, height int) string {
	inner := width - panelOverhead
	var lines []string
	lines = append(lines, s.Title.Render("MESSAGES PER DAY"))

	if usage == nil || len(usage.DailyActivity) < 2 {
		lines = append(lines, s.Dim.Render("Not enough data"))
		content := strings.Join(lines, "\n")
		return s.Panel.Width(width - 2).Height(height - 2).Render(content)
	}

	entries := usage.DailyActivity
	maxDays := inner / 6 // fit 5-char labels + 1 gap
	if maxDays > 14 {
		maxDays = 14
	}
	if maxDays < 2 {
		maxDays = 2
	}
	if len(entries) > maxDays {
		entries = entries[len(entries)-maxDays:]
	}

	first := entries[0].Date[5:]
	last := entries[len(entries)-1].Date[5:]
	lines = append(lines, s.Dim.Render(fmt.Sprintf("%s → %s (%dd)", first, last, len(entries))))

	// Peak
	var maxMsgs int64
	for _, e := range entries {
		if e.MessageCount > maxMsgs {
			maxMsgs = e.MessageCount
		}
	}
	lines = append(lines, s.Dim.Render("peak "+format.FormatCount(maxMsgs)))

	barH := height - 2 - len(lines)
	if barH < 3 {
		barH = 3
	}
	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))
	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorOk)).Background(lipgloss.Color(colorOk))
	bc := barchart.New(inner, barH, barchart.WithStyles(axisStyle, labelStyle), barchart.WithBarGap(1))

	for _, entry := range entries {
		label := entry.Date[8:] + "/" + entry.Date[5:7]
		bc.Push(barchart.BarData{
			Label:  label,
			Values: []barchart.BarValue{{Value: float64(entry.MessageCount), Style: barStyle}},
		})
	}
	bc.Draw()
	lines = append(lines, bc.View())

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderTokenBarPanel renders a bar chart of daily token usage by model.
func renderTokenBarPanel(s Styles, usage *domain.UsageSummary, width, height int) string {
	inner := width - panelOverhead
	var lines []string
	lines = append(lines, s.Title.Render("TOKENS BY MODEL"))

	if usage == nil || len(usage.DailyModelTokens) < 2 {
		lines = append(lines, s.Dim.Render("Not enough data"))
		content := strings.Join(lines, "\n")
		return s.Panel.Width(width - 2).Height(height - 2).Render(content)
	}

	entries := usage.DailyModelTokens
	maxDays := inner / 6
	if maxDays > 14 {
		maxDays = 14
	}
	if maxDays < 2 {
		maxDays = 2
	}
	if len(entries) > maxDays {
		entries = entries[len(entries)-maxDays:]
	}

	// Collect all model names
	modelSet := make(map[string]bool)
	for _, e := range entries {
		for model := range e.TokensByModel {
			modelSet[model] = true
		}
	}
	var models []string
	for m := range modelSet {
		models = append(models, m)
	}
	sort.Strings(models)

	// Date range subtitle
	first := entries[0].Date[5:]
	last := entries[len(entries)-1].Date[5:]
	lines = append(lines, s.Dim.Render(fmt.Sprintf("%s → %s (%dd)", first, last, len(entries))))

	// Legend
	var legendParts []string
	for i, model := range models {
		color := modelColors[i%len(modelColors)]
		dot := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("■")
		legendParts = append(legendParts, dot+" "+s.Dim.Render(stripClaudePrefix(model)))
	}
	lines = append(lines, strings.Join(legendParts, "  "))

	// Find max total tokens per day for peak label
	var maxTotal int64
	for _, entry := range entries {
		var total int64
		for _, tokens := range entry.TokensByModel {
			total += tokens
		}
		if total > maxTotal {
			maxTotal = total
		}
	}
	lines = append(lines, s.Dim.Render("peak "+format.FormatCount(maxTotal)))

	barH := height - 2 - len(lines)
	if barH < 3 {
		barH = 3
	}
	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))
	bc := barchart.New(inner, barH, barchart.WithStyles(axisStyle, labelStyle), barchart.WithBarGap(1))

	for _, entry := range entries {
		var values []barchart.BarValue
		for i, model := range models {
			tokens := entry.TokensByModel[model]
			color := modelColors[i%len(modelColors)]
			values = append(values, barchart.BarValue{
				Name:  stripClaudePrefix(model),
				Value: float64(tokens) / 1000,
				Style: lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Background(lipgloss.Color(color)),
			})
		}
		label := entry.Date[8:] + "/" + entry.Date[5:7]
		bc.Push(barchart.BarData{Label: label, Values: values})
	}
	bc.Draw()
	lines = append(lines, bc.View())

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

		// Fixed-width right columns; project fills the rest.
		innerWidth := width - panelOverhead
		header := formatSessionRow(s.TableHeader, "project", "pid", "cpu", "mem", "uptime", innerWidth)
		lines = append(lines, header)

		// Each session takes 1 line
		availLines := height - panelOverhead - 3 // title + count + header
		maxSessions := availLines - 1            // reserve 1 for potential "+N more"
		if maxSessions < 1 {
			maxSessions = 1
		}

		visible := sessions
		truncated := 0
		if len(visible) > maxSessions {
			truncated = len(visible) - maxSessions
			visible = visible[:maxSessions]
		}

		colPID, colCPU, colMem, colUptime := 6, 6, 6, 8
		fixedW := colPID + colCPU + colMem + colUptime
		colProject := innerWidth - fixedW
		if colProject < 8 {
			colProject = 8
		}

		for _, sess := range visible {
			proj := truncate(sess.ProjectName, colProject-1) // 1 char gap
			pid := fmt.Sprintf("%d", sess.PID)
			cpu := fmt.Sprintf("%.1f%%", sess.CPUPercent)
			mem := fmt.Sprintf("%.1f%%", sess.MemPercent)
			uptime := format.FormatUptime(sess.Uptime)

			projPart := s.ModelName.Render(fmt.Sprintf("%-*s", colProject, proj))
			rest := fmt.Sprintf("%*s%*s%*s%*s", colPID, pid, colCPU, cpu, colMem, mem, colUptime, uptime)
			lines = append(lines, projPart+s.Value.Render(rest))
		}

		if truncated > 0 {
			lines = append(lines, s.Dim.Render(fmt.Sprintf("  +%d more", truncated)))
		}
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderRateLimitsPanel renders the rate limit windows panel.
func renderRateLimitsPanel(s Styles, limits domain.RateLimits, width, height int) string {
	inner := width - panelOverhead
	var lines []string

	title := s.Title.Render("RATE LIMITS")
	if !limits.FetchedAt.IsZero() {
		ago := time.Since(limits.FetchedAt)
		var agoStr string
		if ago < time.Minute {
			agoStr = fmt.Sprintf("%ds ago", int(ago.Seconds()))
		} else {
			agoStr = fmt.Sprintf("%dm ago", int(ago.Minutes()))
		}
		agoRendered := s.Dim.Render(agoStr)
		gap := inner - lipgloss.Width(title) - lipgloss.Width(agoRendered)
		if gap < 1 {
			gap = 1
		}
		title = title + strings.Repeat(" ", gap) + agoRendered
	}
	lines = append(lines, title)

	if limits.FiveHour == nil && limits.SevenDay == nil {
		if !limits.RetryAfter.IsZero() && time.Now().Before(limits.RetryAfter) {
			wait := time.Until(limits.RetryAfter)
			lines = append(lines, s.StatusWarn.Render("API cooldown"))
			lines = append(lines, "")
			lines = append(lines, formatKV(s, "Available in", format.FormatUptime(wait), inner))
		} else if limits.Error != "" {
			lines = append(lines, s.StatusWarn.Render(truncate(limits.Error, inner)))
		} else {
			lines = append(lines, s.Dim.Render("Waiting for data..."))
		}
		content := strings.Join(lines, "\n")
		return s.Panel.Width(width - 2).Height(height - 2).Render(content)
	}

	if w := limits.FiveHour; w != nil {
		lines = append(lines, "")
		lines = append(lines, renderWindowCompact(s, "5h", w, 5*time.Hour, inner))
	}
	if w := limits.SevenDay; w != nil {
		lines = append(lines, "")
		lines = append(lines, renderWindowCompact(s, "7d", w, 7*24*time.Hour, inner))
	}

	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderWindowCompact renders a rate window in 2 lines:
//
//	5h  12.0% ████░░░░░░░░░░ ●
//	    resets 3h1m
func renderWindowCompact(s Styles, label string, w *domain.RateWindow, windowDur time.Duration, inner int) string {
	util := w.Utilization
	utilStyle := s.StatusOk
	if util >= 80 {
		utilStyle = s.StatusErr
	} else if util >= 50 {
		utilStyle = s.StatusWarn
	}

	// Burn rate dot
	remaining := time.Until(w.ResetsAt)
	if remaining < 0 {
		remaining = 0
	}
	elapsed := windowDur - remaining
	if elapsed < 0 {
		elapsed = 0
	}
	elapsedPct := 0.0
	if windowDur > 0 {
		elapsedPct = float64(elapsed) / float64(windowDur) * 100
	}
	diff := util - elapsedPct
	burnDot := s.StatusOk.Render("●")
	if diff > 15 {
		burnDot = s.StatusErr.Render("●")
	} else if diff > 5 {
		burnDot = s.StatusWarn.Render("●")
	}

	// Line 1: "5h  12.0% ████░░░░░░ ●"
	prefix := fmt.Sprintf("%-3s %5.1f%% ", label, util)
	prefixW := len(prefix)
	barWidth := inner - prefixW - 2 // 2 for " ●"
	if barWidth < 3 {
		barWidth = 3
	}
	filled := int(math.Round(util / 100 * float64(barWidth)))
	if filled > barWidth {
		filled = barWidth
	}
	bar := utilStyle.Render(strings.Repeat("█", filled)) + s.Dim.Render(strings.Repeat("░", barWidth-filled))
	line1 := s.Label.Render(label) + utilStyle.Render(fmt.Sprintf(" %5.1f%% ", util)) + bar + " " + burnDot

	// Line 2: right-aligned "resets 3h1m"
	resetStr := format.FormatUptime(remaining)
	resetText := s.Dim.Render("resets " + resetStr)
	pad := inner - lipgloss.Width(resetText)
	if pad < 0 {
		pad = 0
	}
	line2 := strings.Repeat(" ", pad) + resetText

	return line1 + "\n" + line2
}

// renderProcessesView renders the full-width Processes tab.
func renderProcessesView(s Styles, sessions []domain.ActiveSession, width, height int) string {
	var lines []string

	lines = append(lines, s.Title.Render("PROCESSES"))

	count := len(sessions)
	if count == 0 {
		lines = append(lines, "")
		lines = append(lines, s.Dim.Render("No active processes"))
		content := strings.Join(lines, "\n")
		return s.Panel.Width(width - 2).Height(height - 2).Render(content)
	}

	countLabel := "process"
	if count > 1 {
		countLabel = "processes"
	}
	lines = append(lines, s.StatusOk.Render(fmt.Sprintf("● %d active %s", count, countLabel)))
	lines = append(lines, "")

	// Column widths adapt to panel width
	innerWidth := width - panelOverhead
	colProject := 14
	colPID := 8
	colCPU := 8
	colMem := 8
	colUptime := 10
	if innerWidth > 70 {
		colProject = 24
		colPID = 10
		colCPU = 10
		colMem = 10
		colUptime = 12
	}

	headerRow := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		colProject, "PROJECT",
		colPID, "PID",
		colCPU, "CPU",
		colMem, "MEM",
		colUptime, "UPTIME",
	)
	lines = append(lines, s.TableHeader.Render(headerRow))

	for _, sess := range sessions {
		proj := truncate(sess.ProjectName, colProject)
		pid := fmt.Sprintf("%d", sess.PID)
		cpu := fmt.Sprintf("%.1f%%", sess.CPUPercent)
		mem := fmt.Sprintf("%.1f%%", sess.MemPercent)
		uptime := format.FormatUptime(sess.Uptime)

		projStr := s.ModelName.Render(fmt.Sprintf("%-*s", colProject, proj))
		rest := fmt.Sprintf(" %-*s %-*s %-*s %-*s",
			colPID, pid,
			colCPU, cpu,
			colMem, mem,
			colUptime, uptime,
		)
		lines = append(lines, projStr+s.Value.Render(rest))
	}
	content := strings.Join(lines, "\n")
	return s.Panel.Width(width - 2).Height(height - 2).Render(content)
}

// renderModelTokens renders a list of model token rows, capped at max with "+N more".
func renderModelTokens(s Styles, tokens []domain.ModelTokens, inner, max int) []string {
	if len(tokens) == 0 {
		return []string{s.Dim.Render("No token data")}
	}
	var lines []string
	visible := tokens
	truncated := 0
	if len(visible) > max {
		truncated = len(visible) - max
		visible = visible[:max]
	}
	for _, mt := range visible {
		name := stripClaudePrefix(mt.ModelName)
		lines = append(lines, formatModelKV(s, name, format.FormatCount(mt.TokenCount), inner))
	}
	if truncated > 0 {
		lines = append(lines, s.Dim.Render(fmt.Sprintf("  +%d more", truncated)))
	}
	return lines
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
	colPID, colCPU, colMem, colUptime := 6, 6, 6, 8
	fixedW := colPID + colCPU + colMem + colUptime
	colProject := innerWidth - fixedW
	if colProject < 8 {
		colProject = 8
	}

	row := fmt.Sprintf("%-*s%*s%*s%*s%*s",
		colProject, project,
		colPID, pid,
		colCPU, cpu,
		colMem, mem,
		colUptime, uptime,
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
