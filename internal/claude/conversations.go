package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mkrowiarz/ccmonitor/internal/domain"
)

// conversationScanner computes usage stats by scanning JSONL conversation files.
// It caches per-file results keyed by modification time, both in memory and on disk.
type conversationScanner struct {
	mu        sync.Mutex
	fileCache map[string]*fileCacheEntry
	cachePath string // path to the on-disk cache file
	dirty     bool   // true if in-memory cache has changed since last save
}

type fileCacheEntry struct {
	modTime   time.Time
	sessionID string
	dates     map[string]*dateStats
}

type dateStats struct {
	messages      int64
	tokensByModel map[string]int64
}

// Serializable versions for the disk cache.
type convDiskCache struct {
	Version int                       `json:"version"`
	Files   map[string]*diskFileEntry `json:"files"`
}

type diskFileEntry struct {
	ModTime   time.Time                    `json:"modTime"`
	SessionID string                       `json:"sessionId"`
	Dates     map[string]*diskDateStats    `json:"dates"`
}

type diskDateStats struct {
	Messages      int64            `json:"messages"`
	TokensByModel map[string]int64 `json:"tokensByModel"`
}

// jsonlEntry is the minimal structure we need from each JSONL line.
type jsonlEntry struct {
	Type      string        `json:"type"`
	Timestamp string        `json:"timestamp"`
	SessionID string        `json:"sessionId"`
	Message   *jsonlMessage `json:"message,omitempty"`
}

type jsonlMessage struct {
	Model string      `json:"model"`
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
	// Set cache path on first call.
	if cs.cachePath == "" {
		home, _ := os.UserHomeDir()
		cacheDir := filepath.Join(home, ".ccmonitor")
		os.MkdirAll(cacheDir, 0755)
		cs.cachePath = filepath.Join(cacheDir, "conv-cache.json")
		cs.loadDiskCache()
	}

	projectsDir := filepath.Join(claudeDir, "projects")

	// Find all non-subagent JSONL files.
	var files []string
	err := filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
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

	// Identify files that need parsing (not in cache or modtime changed).
	type parseJob struct {
		path    string
		modTime time.Time
	}
	var toParse []parseJob
	activeFiles := make(map[string]bool, len(files))

	for _, path := range files {
		activeFiles[path] = true

		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		cached, ok := cs.fileCache[path]
		if ok && cached.modTime.Equal(info.ModTime()) {
			continue
		}

		toParse = append(toParse, parseJob{path: path, modTime: info.ModTime()})
	}

	// Parse files concurrently.
	if len(toParse) > 0 {
		workers := runtime.NumCPU()
		if workers > 8 {
			workers = 8
		}
		if workers > len(toParse) {
			workers = len(toParse)
		}

		type parseResult struct {
			path  string
			entry *fileCacheEntry
		}

		jobs := make(chan parseJob, len(toParse))
		results := make(chan parseResult, len(toParse))

		var wg sync.WaitGroup
		for range workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					entry := parseConversationFile(job.path)
					results <- parseResult{path: job.path, entry: entry}
				}
			}()
		}

		for _, job := range toParse {
			jobs <- job
		}
		close(jobs)

		go func() {
			wg.Wait()
			close(results)
		}()

		for res := range results {
			if res.entry != nil {
				cs.fileCache[res.path] = res.entry
				cs.dirty = true
			}
		}
	}

	// Remove cache entries for deleted files.
	for path := range cs.fileCache {
		if !activeFiles[path] {
			delete(cs.fileCache, path)
			cs.dirty = true
		}
	}

	// Persist cache to disk if changed.
	if cs.dirty {
		cs.saveDiskCache()
		cs.dirty = false
	}

	return cs.aggregate()
}

func (cs *conversationScanner) loadDiskCache() {
	data, err := os.ReadFile(cs.cachePath)
	if err != nil {
		return
	}

	var dc convDiskCache
	if err := json.Unmarshal(data, &dc); err != nil || dc.Version != 1 {
		return
	}

	for path, de := range dc.Files {
		entry := &fileCacheEntry{
			modTime:   de.ModTime,
			sessionID: de.SessionID,
			dates:     make(map[string]*dateStats, len(de.Dates)),
		}
		for date, dd := range de.Dates {
			entry.dates[date] = &dateStats{
				messages:      dd.Messages,
				tokensByModel: dd.TokensByModel,
			}
		}
		cs.fileCache[path] = entry
	}
}

func (cs *conversationScanner) saveDiskCache() {
	dc := convDiskCache{
		Version: 1,
		Files:   make(map[string]*diskFileEntry, len(cs.fileCache)),
	}

	for path, entry := range cs.fileCache {
		de := &diskFileEntry{
			ModTime:   entry.modTime,
			SessionID: entry.sessionID,
			Dates:     make(map[string]*diskDateStats, len(entry.dates)),
		}
		for date, ds := range entry.dates {
			de.Dates[date] = &diskDateStats{
				Messages:      ds.messages,
				TokensByModel: ds.tokensByModel,
			}
		}
		dc.Files[path] = de
	}

	data, err := json.Marshal(dc)
	if err != nil {
		return
	}

	// Write atomically via temp file.
	tmp := cs.cachePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return
	}
	os.Rename(tmp, cs.cachePath)
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
	scanner.Buffer(make([]byte, 256*1024), 1024*1024)

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

	type dayAgg struct {
		messages      int64
		sessions      map[string]bool
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
