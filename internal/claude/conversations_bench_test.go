package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBenchmarkStartup(t *testing.T) {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude")

	// Skip in CI or when no conversation data exists.
	projectsDir := filepath.Join(claudeDir, "projects")
	if _, err := os.Stat(projectsDir); err != nil {
		t.Skip("no ~/.claude/projects directory, skipping benchmark")
	}

	cachePath := filepath.Join(home, ".ccmonitor", "conv-cache.json")

	// Cold start: remove disk cache
	os.Remove(cachePath)

	scanner1 := newConversationScanner()
	start := time.Now()
	summary, _, err := scanner1.computeStats(claudeDir)
	cold := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Cold start: %v (%d files)", cold, len(scanner1.fileCache))
	if summary.LifetimeMessages != nil {
		t.Logf("  Messages: %d", *summary.LifetimeMessages)
	}

	if len(scanner1.fileCache) == 0 {
		t.Skip("no conversation files found, skipping cache verification")
	}

	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("disk cache not written: %v", err)
	}
	t.Logf("  Disk cache: %s", formatBytes(info.Size()))

	// Warm start: new scanner loads from disk
	scanner2 := newConversationScanner()
	start = time.Now()
	_, _, err = scanner2.computeStats(claudeDir)
	warm := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Warm start: %v", warm)

	// In-memory: same scanner, second call
	start = time.Now()
	_, _, err = scanner2.computeStats(claudeDir)
	mem := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("In-memory:  %v", mem)
}

func formatBytes(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%dB", b)
	}
	return fmt.Sprintf("%.1fKB", float64(b)/1024)
}
