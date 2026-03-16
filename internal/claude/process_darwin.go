//go:build darwin

package claude

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func discoverProcesses(ctx context.Context) ([]processInfo, error) {
	cmd := exec.CommandContext(ctx, "ps", "-eo", "pid,pcpu,pmem,etime,comm")
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
		p, ok := parsePSLine(line)
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

func parsePSLine(line string) (processInfo, bool) {
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

// parseElapsedDarwin parses macOS ps etime format:
// "MM:SS", "HH:MM:SS", or "D-HH:MM:SS"
func parseElapsedDarwin(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	var days int
	rest := s

	if idx := strings.Index(s, "-"); idx >= 0 {
		d, err := strconv.Atoi(s[:idx])
		if err != nil {
			return 0, fmt.Errorf("invalid days in %q: %w", s, err)
		}
		days = d
		rest = s[idx+1:]
	}

	parts := strings.Split(rest, ":")
	switch len(parts) {
	case 2:
		// MM:SS
		m, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes in %q: %w", s, err)
		}
		sec, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds in %q: %w", s, err)
		}
		return time.Duration(days)*24*time.Hour +
			time.Duration(m)*time.Minute +
			time.Duration(sec)*time.Second, nil
	case 3:
		// HH:MM:SS
		h, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid hours in %q: %w", s, err)
		}
		m, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes in %q: %w", s, err)
		}
		sec, err := strconv.Atoi(parts[2])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds in %q: %w", s, err)
		}
		return time.Duration(days)*24*time.Hour +
			time.Duration(h)*time.Hour +
			time.Duration(m)*time.Minute +
			time.Duration(sec)*time.Second, nil
	default:
		return 0, fmt.Errorf("unexpected elapsed format: %q", s)
	}
}

func resolveProjectName(pid int) string {
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid), "-Fn")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "n/") {
			path := line[1:] // strip the 'n' prefix
			return filepath.Base(path)
		}
	}
	return "unknown"
}
