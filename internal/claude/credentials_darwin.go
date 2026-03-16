//go:build darwin

package claude

import (
	"encoding/json"
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

	var creds struct {
		ClaudeAiOauth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return "", fmt.Errorf("credential JSON parse error: %w", err)
	}

	token := creds.ClaudeAiOauth.AccessToken
	if token == "" {
		return "", fmt.Errorf("no access token in credentials")
	}
	return token, nil
}

// rateLimitsSupported returns true on macOS where Keychain credentials are available.
func rateLimitsSupported() bool { return true }
