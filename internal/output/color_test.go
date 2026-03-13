package output

import "testing"

func TestColorStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		isTTY   bool
		noColor bool
		want    string
	}{
		{"green active tty", "active", true, false, ansiGreen + "active" + ansiReset},
		{"green success tty", "success", true, false, ansiGreen + "success" + ansiReset},
		{"red error tty", "error", true, false, ansiRed + "error" + ansiReset},
		{"red failed tty", "failed", true, false, ansiRed + "failed" + ansiReset},
		{"yellow paused tty", "paused", true, false, ansiYellow + "paused" + ansiReset},
		{"yellow running tty", "running", true, false, ansiYellow + "running" + ansiReset},
		{"yellow pending tty", "pending", true, false, ansiYellow + "pending" + ansiReset},
		{"unknown status unchanged", "draft", true, false, "draft"},
		{"not tty returns plain", "active", false, false, "active"},
		{"noColor returns plain", "error", true, true, "error"},
		{"not tty and noColor", "success", false, true, "success"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColorStatus(tt.status, tt.isTTY, tt.noColor)
			if got != tt.want {
				t.Errorf("ColorStatus(%q, %v, %v) = %q, want %q", tt.status, tt.isTTY, tt.noColor, got, tt.want)
			}
		})
	}
}
