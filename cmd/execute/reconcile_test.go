package execute_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/keeperhub/cli/cmd/execute"
	"github.com/keeperhub/cli/pkg/iostreams"
)

// The regression this addresses. A write reports completed with no transaction hash, which the
// live API does for /api/execute/contract-call, so --wait must reconcile against the status
// endpoint instead of returning a success the caller cannot verify.
func TestTransferCmd_WaitReconcilesCompletedWithoutHash(t *testing.T) {
	statusCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/status") {
			statusCalls++
			w.Header().Set("X-Poll-Interval-Hint", "0")
			_, _ = w.Write([]byte(`{"executionId":"exec-1","status":"completed","transactionHash":"0xdeadbeef"}`))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-1","status":"completed"}`))
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	cmd := execute.NewTransferCmd(newTransferFactory(ios, srv))
	cmd.SetArgs([]string{"--chain", "84532", "--to", "0xabc", "--amount", "0.1", "--wait"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCalls == 0 {
		t.Fatal("expected the status endpoint to be polled so the caller ends up with a transaction hash")
	}
	if out := buf.String(); !strings.Contains(out, "0xdeadbeef") {
		t.Errorf("expected the reconciled transaction hash in the output, got: %q", out)
	}
}

func TestContractCallCmd_WaitReconcilesCompletedWithoutHash(t *testing.T) {
	statusCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/status") {
			statusCalls++
			w.Header().Set("X-Poll-Interval-Hint", "0")
			_, _ = w.Write([]byte(`{"executionId":"exec-cc","status":"completed","transactionHash":"0xc0ffee"}`))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-cc","status":"completed"}`))
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	cmd := execute.NewContractCallCmd(newContractCallFactory(ios, srv))
	cmd.SetArgs([]string{
		"--chain", "84532",
		"--contract", "0x2A6FC8182Bf9928Ef7517dA980dC79e8107c555A",
		"--method", "ping",
		"--wait",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCalls == 0 {
		t.Fatal("expected the status endpoint to be polled for a completed write with no hash")
	}
	if out := buf.String(); !strings.Contains(out, "0xc0ffee") {
		t.Errorf("expected the reconciled transaction hash in the output, got: %q", out)
	}
}

// The server's pacing is honoured rather than a fixed client-side timer. A large hint on the
// first poll would previously have been ignored, and the client would have polled on its own
// two second cadence regardless of what the server asked for.
func TestPollHonoursServerInterval(t *testing.T) {
	var firstPollAt, secondPollAt time.Time
	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.HasSuffix(r.URL.Path, "/status") {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-slow","status":"running"}`))
			return
		}
		polls++
		switch polls {
		case 1:
			firstPollAt = time.Now()
			w.Header().Set("X-Poll-Interval-Hint", "1")
			_, _ = w.Write([]byte(`{"executionId":"exec-slow","status":"running"}`))
		default:
			secondPollAt = time.Now()
			w.Header().Set("X-Poll-Interval-Hint", "0")
			_, _ = w.Write([]byte(`{"executionId":"exec-slow","status":"completed","transactionHash":"0xok"}`))
		}
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	cmd := execute.NewTransferCmd(newTransferFactory(ios, srv))
	cmd.SetArgs([]string{"--chain", "84532", "--to", "0xabc", "--amount", "0.1", "--wait"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if polls < 2 {
		t.Fatalf("expected at least two polls, got %d", polls)
	}
	gap := secondPollAt.Sub(firstPollAt)
	if gap < 900*time.Millisecond {
		t.Errorf("second poll came after %v, expected roughly the 1s the server asked for", gap)
	}
	if gap > 1900*time.Millisecond {
		t.Errorf("second poll came after %v, which suggests the old fixed 2s timer rather than the hint", gap)
	}
}
