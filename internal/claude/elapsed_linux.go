//go:build linux

package claude

import "time"

func parseElapsed(s string) (time.Duration, error) {
	return parseElapsedLinux(s)
}
