package claude

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mkrowiarz/ccmonitor/internal/domain"
)

type historyEntry struct {
	Display   string `json:"display"`
	Timestamp int64  `json:"timestamp"`
	Project   string `json:"project"`
	SessionID string `json:"sessionId"`
}

func parseHistory(path string, limit int) ([]domain.RecentEvent, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, []string{"history.jsonl not found"}, nil
		}
		return nil, nil, fmt.Errorf("opening history file: %w", err)
	}
	defer f.Close()

	var events []domain.RecentEvent
	var warnings []string
	malformed := 0

	scanner := bufio.NewScanner(f)
	// Increase buffer for potentially long lines
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry historyEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			malformed++
			continue
		}

		display := strings.TrimSpace(entry.Display)
		if len(display) > 50 {
			display = display[:50]
		}

		events = append(events, domain.RecentEvent{
			Timestamp:   time.UnixMilli(entry.Timestamp),
			ProjectName: filepath.Base(entry.Project),
			Display:     display,
			SessionID:   entry.SessionID,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("reading history file: %w", err)
	}

	if malformed > 0 {
		warnings = append(warnings, fmt.Sprintf("skipped %d malformed lines", malformed))
	}

	// Sort by timestamp descending
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, warnings, nil
}
