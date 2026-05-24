package claude

import (
	"encoding/json"
	"fmt"
)

// parseAccessToken extracts the Claude Code OAuth access token from the raw
// credentials JSON. The same shape is used by the macOS Keychain entry and the
// Linux ~/.claude/.credentials.json file, so both platforms share this parser.
func parseAccessToken(raw []byte) (string, error) {
	var creds struct {
		ClaudeAiOauth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(raw, &creds); err != nil {
		return "", fmt.Errorf("credential JSON parse error: %w", err)
	}
	if creds.ClaudeAiOauth.AccessToken == "" {
		return "", fmt.Errorf("no access token in credentials")
	}
	return creds.ClaudeAiOauth.AccessToken, nil
}
