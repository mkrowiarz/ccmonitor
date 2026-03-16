package claude

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/mkrowiarz/ccmonitor/internal/domain"
)

type statsCache struct {
	Version          int                `json:"version"`
	LastComputedDate string             `json:"lastComputedDate"`
	DailyActivity    []dailyActivity    `json:"dailyActivity"`
	DailyModelTokens []dailyModelTokens `json:"dailyModelTokens"`
	ModelUsage       map[string]modelUsageEntry `json:"modelUsage"`
	TotalSessions    int64              `json:"totalSessions"`
	TotalMessages    int64              `json:"totalMessages"`
}

type dailyActivity struct {
	Date           string `json:"date"`
	MessageCount   int64  `json:"messageCount"`
	SessionCount   int64  `json:"sessionCount"`
	ToolCallCount  int64  `json:"toolCallCount"`
}

type dailyModelTokens struct {
	Date          string           `json:"date"`
	TokensByModel map[string]int64 `json:"tokensByModel"`
}

type modelUsageEntry struct {
	InputTokens             int64 `json:"inputTokens"`
	OutputTokens            int64 `json:"outputTokens"`
	CacheReadInputTokens    int64 `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int64 `json:"cacheCreationInputTokens"`
}

func parseStatsCache(path string) (*domain.UsageSummary, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, []string{"stats-cache.json not found"}, nil
		}
		return nil, nil, fmt.Errorf("reading stats-cache: %w", err)
	}

	var sc statsCache
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, nil, fmt.Errorf("parsing stats-cache JSON: %w", err)
	}

	var warnings []string
	summary := &domain.UsageSummary{
		SourceDate: sc.LastComputedDate,
	}

	// Today's activity
	todayDate := sc.LastComputedDate
	foundActivity := false
	for _, da := range sc.DailyActivity {
		if da.Date == todayDate {
			msgs := da.MessageCount
			sess := da.SessionCount
			summary.TodayMessages = &msgs
			summary.TodaySessions = &sess
			foundActivity = true
			break
		}
	}
	if !foundActivity {
		warnings = append(warnings, "no daily activity entry for "+todayDate)
	}

	// Today's model tokens
	foundTokens := false
	for _, dt := range sc.DailyModelTokens {
		if dt.Date == todayDate {
			for model, count := range dt.TokensByModel {
				summary.TodayTokens = append(summary.TodayTokens, domain.ModelTokens{
					ModelName:  model,
					TokenCount: count,
				})
			}
			sort.Slice(summary.TodayTokens, func(i, j int) bool {
				return summary.TodayTokens[i].ModelName < summary.TodayTokens[j].ModelName
			})
			foundTokens = true
			break
		}
	}
	if !foundTokens {
		warnings = append(warnings, "no daily model tokens entry for "+todayDate)
	}

	// Daily activity history (sorted by date ascending)
	for _, da := range sc.DailyActivity {
		summary.DailyActivity = append(summary.DailyActivity, domain.DailyActivityEntry{
			Date:         da.Date,
			MessageCount: da.MessageCount,
			SessionCount: da.SessionCount,
		})
	}
	sort.Slice(summary.DailyActivity, func(i, j int) bool {
		return summary.DailyActivity[i].Date < summary.DailyActivity[j].Date
	})

	// Daily model tokens history (sorted by date ascending)
	for _, dt := range sc.DailyModelTokens {
		summary.DailyModelTokens = append(summary.DailyModelTokens, domain.DailyModelTokensEntry{
			Date:          dt.Date,
			TokensByModel: dt.TokensByModel,
		})
	}
	sort.Slice(summary.DailyModelTokens, func(i, j int) bool {
		return summary.DailyModelTokens[i].Date < summary.DailyModelTokens[j].Date
	})

	// Lifetime counts
	ltMsgs := sc.TotalMessages
	ltSess := sc.TotalSessions
	summary.LifetimeMessages = &ltMsgs
	summary.LifetimeSessions = &ltSess

	// Lifetime tokens per model
	for model, usage := range sc.ModelUsage {
		total := usage.InputTokens + usage.OutputTokens + usage.CacheReadInputTokens + usage.CacheCreationInputTokens
		summary.LifetimeTokens = append(summary.LifetimeTokens, domain.ModelTokens{
			ModelName:  model,
			TokenCount: total,
		})
	}
	sort.Slice(summary.LifetimeTokens, func(i, j int) bool {
		return summary.LifetimeTokens[i].ModelName < summary.LifetimeTokens[j].ModelName
	})

	return summary, warnings, nil
}
