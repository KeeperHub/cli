package execute

import (
	"net/http"
	"testing"
	"time"
)

// A successful /api/execute/contract-call broadcast can return 202 with status "completed" and
// no transactionHash; the hash only appears on the status endpoint, so a direct-write response in
// this shape is worth one reconciling fetch. On a status response the same shape is legitimate,
// because an action that submits nothing onchain completes this way, so it is only reported.
func TestCompletedWithoutTransaction(t *testing.T) {
	empty := ""
	hash := "0xabc"

	cases := []struct {
		name   string
		status string
		tx     *string
		want   bool
	}{
		{"completed without a hash", "completed", nil, true},
		{"completed with an empty hash", "completed", &empty, true},
		{"completed with a hash", "completed", &hash, false},
		{"failed carries no such expectation", "failed", nil, false},
		{"running is not completed", "running", nil, false},
		{"unconfirmed already implies a broadcast", "unconfirmed", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := completedWithoutTransaction(tc.status, tc.tx); got != tc.want {
				t.Errorf("completedWithoutTransaction(%q, %v) = %v, want %v", tc.status, tc.tx, got, tc.want)
			}
		})
	}
}

// unconfirmed is terminal for a client: nothing moves it until the reconciler runs, so a poll loop
// that treats it as pending just burns requests until the caller's timeout.
func TestExecTerminalStatuses(t *testing.T) {
	for _, status := range []string{"completed", "failed", "unconfirmed"} {
		if !execTerminalStatuses[status] {
			t.Errorf("expected %q to be terminal", status)
		}
	}
	for _, status := range []string{"pending", "running"} {
		if execTerminalStatuses[status] {
			t.Errorf("expected %q not to be terminal", status)
		}
	}
}

func TestNextPollDelay(t *testing.T) {
	cases := []struct {
		name         string
		header       string
		setHeader    bool
		wantDelay    time.Duration
		wantTerminal bool
	}{
		{"no header falls back to the default", "", false, defaultPollInterval, false},
		{"a hint is honoured", "5", true, 5 * time.Second, false},
		{"a hint of zero means terminal", "0", true, 0, true},
		{"surrounding whitespace is tolerated", "  3 ", true, 3 * time.Second, false},
		{"an unparseable hint falls back rather than failing", "soon", true, defaultPollInterval, false},
		{"a negative hint falls back", "-1", true, defaultPollInterval, false},
		{"a hint at the ceiling is honoured", "30", true, maxPollIntervalSecs * time.Second, false},
		{"an oversized hint is clamped to the ceiling", "3600", true, maxPollIntervalSecs * time.Second, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{Header: http.Header{}}
			if tc.setHeader {
				resp.Header.Set("X-Poll-Interval-Hint", tc.header)
			}
			delay, terminal := nextPollDelay(resp)
			if delay != tc.wantDelay || terminal != tc.wantTerminal {
				t.Errorf("nextPollDelay(%q) = (%v, %v), want (%v, %v)",
					tc.header, delay, terminal, tc.wantDelay, tc.wantTerminal)
			}
		})
	}
}
