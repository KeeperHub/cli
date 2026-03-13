package output

import (
	"testing"
	"time"
)

func TestTimeAgo(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "just now",
			input: now.Add(-10 * time.Second).Format(time.RFC3339),
			want:  "just now",
		},
		{
			name:  "minutes ago",
			input: now.Add(-5 * time.Minute).Format(time.RFC3339),
			want:  "5m ago",
		},
		{
			name:  "hours ago",
			input: now.Add(-2 * time.Hour).Format(time.RFC3339),
			want:  "2h ago",
		},
		{
			name:  "days ago",
			input: now.Add(-3 * 24 * time.Hour).Format(time.RFC3339),
			want:  "3d ago",
		},
		{
			name:  "months ago",
			input: now.Add(-65 * 24 * time.Hour).Format(time.RFC3339),
			want:  "2mo ago",
		},
		{
			name:  "years ago",
			input: now.Add(-400 * 24 * time.Hour).Format(time.RFC3339),
			want:  "1y ago",
		},
		{
			name:  "invalid string returned as-is",
			input: "not-a-date",
			want:  "not-a-date",
		},
		{
			name:  "RFC3339Nano format",
			input: now.Add(-30 * time.Minute).Format(time.RFC3339Nano),
			want:  "30m ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimeAgo(tt.input)
			if got != tt.want {
				t.Errorf("TimeAgo(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTimeAgoSince(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "just now"},
		{"30 seconds", 30 * time.Second, "just now"},
		{"59 seconds", 59 * time.Second, "just now"},
		{"1 minute", 60 * time.Second, "1m ago"},
		{"45 minutes", 45 * time.Minute, "45m ago"},
		{"1 hour", 60 * time.Minute, "1h ago"},
		{"23 hours", 23 * time.Hour, "23h ago"},
		{"1 day", 24 * time.Hour, "1d ago"},
		{"29 days", 29 * 24 * time.Hour, "29d ago"},
		{"30 days", 30 * 24 * time.Hour, "1mo ago"},
		{"11 months", 330 * 24 * time.Hour, "11mo ago"},
		{"12 months", 365 * 24 * time.Hour, "1y ago"},
		{"2 years", 730 * 24 * time.Hour, "2y ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timeAgoSince(tt.duration)
			if got != tt.want {
				t.Errorf("timeAgoSince(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}
