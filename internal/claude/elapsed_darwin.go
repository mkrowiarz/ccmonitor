//go:build darwin

package claude

import "time"

func parseElapsed(s string) (time.Duration, error) {
	return parseElapsedDarwin(s)
}
