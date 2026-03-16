package claude

import (
	"os"
	"testing"
)

func TestReadOAuthToken_ErrorHandling(t *testing.T) {
	// Skip in CI where keychain isn't available
	if os.Getenv("CI") != "" {
		t.Skip("skipping keychain test in CI")
	}

	// This tests that the function handles errors gracefully.
	// In environments without the keychain entry, it should return
	// a clear error rather than panicking.
	_, err := readOAuthToken()
	if err != nil {
		t.Logf("readOAuthToken returned expected error: %v", err)
	}
	// If it succeeds, the token was found — that's fine too.
}
