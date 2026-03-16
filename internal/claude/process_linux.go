//go:build linux

package claude

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func discoverProcesses(ctx context.Context) ([]processInfo, error) {
	cmd := exec.CommandContext(ctx, "ps", "-eo", "pid,pcpu,pmem,etimes,comm")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}

	var procs []processInfo
	scanner := bufio.NewScanner(bytes.NewReader(out))
	// skip header
	if scanner.Scan() {
		// discard header line
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		p, ok := parsePSLineLinux(line)
		if !ok {
			continue
		}
		base := filepath.Base(p.Comm)
		if base != "claude" {
			continue
		}
		p.Comm = base
		procs = append(procs, p)
	}
	return procs, scanner.Err()
}

func parsePSLineLinux(line string) (processInfo, bool) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return processInfo{}, false
	}
	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return processInfo{}, false
	}
	cpu, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return processInfo{}, false
	}
	mem, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return processInfo{}, false
	}
	elapsed := fields[3]
	comm := strings.Join(fields[4:], " ")

	return processInfo{
		PID:        pid,
		CPUPercent: cpu,
		MemPercent: mem,
		Elapsed:    elapsed,
		Comm:       comm,
	}, true
}

// parseElapsedLinux parses Linux etimes (elapsed time in seconds).
func parseElapsedLinux(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	secs, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid etimes value %q: %w", s, err)
	}
	return time.Duration(secs) * time.Second, nil
}

func resolveProjectName(pid int) string {
	link := fmt.Sprintf("/proc/%d/cwd", pid)
	target, err := os.Readlink(link)
	if err != nil {
		return "unknown"
	}
	return filepath.Base(target)
}
