//go:build darwin

package claude

import (
	"fmt"
	"os/exec"
	"strings"
)

// readOAuthToken reads the Claude Code OAuth access token from the macOS Keychain.
func readOAuthToken() (string, error) {
	out, err := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w").Output()
	if err != nil {
		return "", fmt.Errorf("keychain lookup failed: %w", err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return "", fmt.Errorf("empty keychain entry")
	}
	return parseAccessToken([]byte(raw))
}

// rateLimitsSupported returns true on macOS where Keychain credentials are available.
func rateLimitsSupported() bool { return true }
