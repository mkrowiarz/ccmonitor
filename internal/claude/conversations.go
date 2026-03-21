package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mkrowiarz/ccmonitor/internal/domain"
)

// conversationScanner computes usage stats by scanning JSONL conversation files.
// It caches per-file results and only re-parses files whose modification time changed.
type conversationScanner struct {
	mu        sync.Mutex
	fileCache map[string]*fileCacheEntry
}

type fileCacheEntry struct {
	modTime   time.Time
	sessionID string
	// Per-date aggregated stats for this file.
	dates map[string]*dateStats
}

type dateStats struct {
	messages     int64
	tokensByModel map[string]int64
}

// jsonlEntry is the minimal structure we need from each JSONL line.
type jsonlEntry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"sessionId"`
	Message   *jsonlMessage   `json:"message,omitempty"`
}

type jsonlMessage struct {
	Model string     `json:"model"`
	Usage *jsonlUsage `json:"usage,omitempty"`
}

type jsonlUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
}

func newConversationScanner() *conversationScanner {
	return &conversationScanner{
		fileCache: make(map[string]*fileCacheEntry),
	}
}

// computeStats scans JSONL files under claudeDir/projects and returns a UsageSummary.
func (cs *conversationScanner) computeStats(claudeDir string) (*domain.UsageSummary, []string, error) {
	projectsDir := filepath.Join(claudeDir, "projects")

	// Find all non-subagent JSONL files.
	var files []string
	err := filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible dirs
		}
		if d.IsDir() {
			if d.Name() == "subagents" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".jsonl" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Track which files still exist for cache cleanup.
	activeFiles := make(map[string]bool, len(files))

	for _, path := range files {
		activeFiles[path] = true

		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		cached, ok := cs.fileCache[path]
		if ok && cached.modTime.Equal(info.ModTime()) {
			continue // file hasn't changed
		}

		// Parse or re-parse this file.
		entry := parseConversationFile(path)
		if entry != nil {
			cs.fileCache[path] = entry
		}
	}

	// Remove cache entries for deleted files.
	for path := range cs.fileCache {
		if !activeFiles[path] {
			delete(cs.fileCache, path)
		}
	}

	// Aggregate all cached data into a summary.
	return cs.aggregate()
}

func parseConversationFile(path string) *fileCacheEntry {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	sessionID := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	entry := &fileCacheEntry{
		modTime:   info.ModTime(),
		sessionID: sessionID,
		dates:     make(map[string]*dateStats),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 1024*1024) // up to 1MB lines

	for scanner.Scan() {
		var e jsonlEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}

		if len(e.Timestamp) < 10 {
			continue
		}
		date := e.Timestamp[:10]

		switch e.Type {
		case "user":
			ds := entry.getOrCreateDate(date)
			ds.messages++

		case "assistant":
			if e.Message != nil && e.Message.Usage != nil {
				u := e.Message.Usage
				tokens := u.InputTokens + u.OutputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens
				if tokens > 0 && e.Message.Model != "" {
					ds := entry.getOrCreateDate(date)
					if ds.tokensByModel == nil {
						ds.tokensByModel = make(map[string]int64)
					}
					ds.tokensByModel[e.Message.Model] += tokens
				}
			}
		}

		// Use sessionId from entry if available.
		if e.SessionID != "" {
			entry.sessionID = e.SessionID
		}
	}

	return entry
}

func (e *fileCacheEntry) getOrCreateDate(date string) *dateStats {
	ds, ok := e.dates[date]
	if !ok {
		ds = &dateStats{
			tokensByModel: make(map[string]int64),
		}
		e.dates[date] = ds
	}
	return ds
}

func (cs *conversationScanner) aggregate() (*domain.UsageSummary, []string, error) {
	today := time.Now().Format("2006-01-02")

	// Per-date aggregations.
	type dayAgg struct {
		messages     int64
		sessions     map[string]bool
		tokensByModel map[string]int64
	}
	days := make(map[string]*dayAgg)

	var totalMessages int64
	allSessions := make(map[string]bool)
	totalTokensByModel := make(map[string]int64)

	for _, entry := range cs.fileCache {
		for date, ds := range entry.dates {
			d, ok := days[date]
			if !ok {
				d = &dayAgg{
					sessions:      make(map[string]bool),
					tokensByModel: make(map[string]int64),
				}
				days[date] = d
			}
			d.messages += ds.messages
			d.sessions[entry.sessionID] = true
			for model, tokens := range ds.tokensByModel {
				d.tokensByModel[model] += tokens
			}

			totalMessages += ds.messages
			allSessions[entry.sessionID] = true
			for model, tokens := range ds.tokensByModel {
				totalTokensByModel[model] += tokens
			}
		}
	}

	summary := &domain.UsageSummary{
		SourceDate: today,
	}

	// Today's stats.
	if d, ok := days[today]; ok {
		msgs := d.messages
		sess := int64(len(d.sessions))
		summary.TodayMessages = &msgs
		summary.TodaySessions = &sess
		for model, tokens := range d.tokensByModel {
			summary.TodayTokens = append(summary.TodayTokens, domain.ModelTokens{
				ModelName:  model,
				TokenCount: tokens,
			})
		}
		sort.Slice(summary.TodayTokens, func(i, j int) bool {
			return summary.TodayTokens[i].ModelName < summary.TodayTokens[j].ModelName
		})
	} else {
		zero := int64(0)
		summary.TodayMessages = &zero
		summary.TodaySessions = &zero
	}

	// Lifetime stats.
	ltSess := int64(len(allSessions))
	summary.LifetimeMessages = &totalMessages
	summary.LifetimeSessions = &ltSess
	for model, tokens := range totalTokensByModel {
		summary.LifetimeTokens = append(summary.LifetimeTokens, domain.ModelTokens{
			ModelName:  model,
			TokenCount: tokens,
		})
	}
	sort.Slice(summary.LifetimeTokens, func(i, j int) bool {
		return summary.LifetimeTokens[i].ModelName < summary.LifetimeTokens[j].ModelName
	})

	// Daily activity history.
	for date, d := range days {
		summary.DailyActivity = append(summary.DailyActivity, domain.DailyActivityEntry{
			Date:         date,
			MessageCount: d.messages,
			SessionCount: int64(len(d.sessions)),
		})
		summary.DailyModelTokens = append(summary.DailyModelTokens, domain.DailyModelTokensEntry{
			Date:          date,
			TokensByModel: d.tokensByModel,
		})
	}
	sort.Slice(summary.DailyActivity, func(i, j int) bool {
		return summary.DailyActivity[i].Date < summary.DailyActivity[j].Date
	})
	sort.Slice(summary.DailyModelTokens, func(i, j int) bool {
		return summary.DailyModelTokens[i].Date < summary.DailyModelTokens[j].Date
	})

	return summary, nil, nil
}
