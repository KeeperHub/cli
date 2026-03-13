package output

import (
	"os"

	"golang.org/x/term"
)

// TerminalWidth returns the width of the terminal connected to stdout.
// Returns 120 as a sensible default if the width cannot be determined.
func TerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 120
	}
	return width
}

// TruncateString truncates s to max characters. If s exceeds max, it is
// shortened and "..." is appended. max must be at least 4 for truncation
// to apply; otherwise the original string is returned.
func TruncateString(s string, max int) string {
	if max < 4 || len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
