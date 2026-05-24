//go:build linux

package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// readOAuthToken reads the Claude Code OAuth access token from
// ~/.claude/.credentials.json, where Claude Code stores it on Linux (there is
// no Keychain). The token is refreshed by Claude Code during normal use.
func readOAuthToken() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	path := filepath.Join(home, ".claude", ".credentials.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read credentials: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return "", fmt.Errorf("empty credentials file")
	}
	return parseAccessToken(data)
}

// rateLimitsSupported returns true on Linux, where Claude Code stores
// credentials in ~/.claude/.credentials.json.
func rateLimitsSupported() bool { return true }
