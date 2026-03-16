package format

import "fmt"

// FormatCount formats a number into compact human-readable form.
// Examples: 56 → "56", 4000 → "4.0K", 107570 → "107K", 265003000 → "265M", 1934004200 → "1B"
func FormatCount(n int64) string {
	switch {
	case n >= 1_000_000_000:
		return formatCompact(n, 1_000_000_000, "B")
	case n >= 1_000_000:
		return formatCompact(n, 1_000_000, "M")
	case n >= 1000:
		return formatCompact(n, 1000, "K")
	default:
		return fmt.Sprintf("%d", n)
	}
}

func formatCompact(n int64, divisor int64, suffix string) string {
	whole := n / divisor
	if whole >= 10 {
		return fmt.Sprintf("%d%s", whole, suffix)
	}
	// For values < 10: show one decimal only if it divides evenly, otherwise truncate to integer.
	if n%divisor == 0 {
		return fmt.Sprintf("%d.0%s", whole, suffix)
	}
	return fmt.Sprintf("%d%s", whole, suffix)
}
