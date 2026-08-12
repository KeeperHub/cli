package execute

import (
	"net/http"
	"testing"
	"time"
)

// A successful /api/execute/contract-call broadcast can return 202 with status "completed" and
// no transactionHash; the hash only appears on the status endpoint. Treating that as final means
// --wait returns success while the caller still has no transaction to verify.
func TestWriteNeedsReconciliation(t *testing.T) {
	empty := ""
	hash := "0xabc"

	cases := []struct {
		name   string
		status string
		tx     *string
		want   bool
	}{
		{"completed without a hash must be reconciled", "completed", nil, true},
		{"completed with an empty hash must be reconciled", "completed", &empty, true},
		{"completed with a hash is final", "completed", &hash, false},
		{"failed is final regardless of the hash", "failed", nil, false},
		{"running is not terminal anyway", "running", nil, false},
		{"unconfirmed is not terminal anyway", "unconfirmed", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := writeNeedsReconciliation(tc.status, tc.tx); got != tc.want {
				t.Errorf("writeNeedsReconciliation(%q, %v) = %v, want %v", tc.status, tc.tx, got, tc.want)
			}
		})
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
