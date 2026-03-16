//go:build linux

package claude

// readOAuthToken is not supported on Linux.
func readOAuthToken() (string, error) {
	return "", nil
}

// rateLimitsSupported returns false on Linux where Keychain is not available.
func rateLimitsSupported() bool { return false }
