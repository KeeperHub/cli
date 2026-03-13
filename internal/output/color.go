package output

// ANSI escape codes for terminal colors.
const (
	ansiReset  = "\033[0m"
	ansiGreen  = "\033[32m"
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
)

// ColorStatus wraps a status string with ANSI color codes for TTY display.
// Green: active, success. Red: error, failed. Yellow: paused, running, pending.
// Returns the unmodified string when not a TTY or when noColor is true.
func ColorStatus(status string, isTTY bool, noColor bool) string {
	if !isTTY || noColor {
		return status
	}

	switch status {
	case "active", "success":
		return ansiGreen + status + ansiReset
	case "error", "failed":
		return ansiRed + status + ansiReset
	case "paused", "running", "pending":
		return ansiYellow + status + ansiReset
	default:
		return status
	}
}
