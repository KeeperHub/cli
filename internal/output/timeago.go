package output

import (
	"fmt"
	"time"
)

// TimeAgo converts an ISO 8601 timestamp string to a human-friendly relative
// time like "2h ago", "3d ago", or "just now". On parse error it returns the
// original string unchanged.
func TimeAgo(isoString string) string {
	t, err := time.Parse(time.RFC3339, isoString)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, isoString)
		if err != nil {
			return isoString
		}
	}
	return timeAgoSince(time.Since(t))
}

func timeAgoSince(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return "just now"
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}
	days := hours / 24
	if days < 30 {
		return fmt.Sprintf("%dd ago", days)
	}
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%dmo ago", months)
	}
	years := months / 12
	return fmt.Sprintf("%dy ago", years)
}
