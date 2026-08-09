package execute_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/keeperhub/cli/cmd/execute"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func serveStatus(t *testing.T, resp execute.ExecStatusResponse) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}))
}

func verifiedReceipt(hash string) execute.ExecReceipt {
	return execute.ExecReceipt{
		Hash:          hash,
		ChainID:       84532,
		Verified:      true,
		ReceiptStatus: "success",
	}
}

func TestExecStatusCmd_RequireVerified_PassesWithVerifiedReceipts(t *testing.T) {
	txHash := "0xabc"
	srv := serveStatus(t, execute.ExecStatusResponse{
		ExecutionID:     "exec-verified",
		Status:          "completed",
		TransactionHash: &txHash,
		Receipts:        []execute.ExecReceipt{verifiedReceipt("0xabc")},
	})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newStatusFactory(ios, srv)

	cmd := execute.NewStatusCmd(f)
	cmd.SetArgs([]string{"exec-verified", "--require-verified"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "0xabc") {
		t.Errorf("expected receipt hash in output, got: %q", out)
	}
	if !strings.Contains(out, "verified") {
		t.Errorf("expected verified marker in output, got: %q", out)
	}
}

func TestExecStatusCmd_RequireVerified_FailsWhenNoReceipts(t *testing.T) {
	srv := serveStatus(t, execute.ExecStatusResponse{
		ExecutionID: "exec-noproof",
		Status:      "completed",
	})
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newStatusFactory(ios, srv)

	cmd := execute.NewStatusCmd(f)
	cmd.SetArgs([]string{"exec-noproof", "--require-verified"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when completed execution has no receipts, got nil")
	}
	if !strings.Contains(err.Error(), "no receipts") {
		t.Errorf("expected 'no receipts' in error, got: %q", err.Error())
	}
}

func TestExecStatusCmd_RequireVerified_FailsWhenReceiptUnverified(t *testing.T) {
	srv := serveStatus(t, execute.ExecStatusResponse{
		ExecutionID: "exec-unverified",
		Status:      "completed",
		Receipts: []execute.ExecReceipt{{
			Hash:          "0xdead",
			ChainID:       84532,
			Verified:      false,
			ReceiptStatus: "success",
		}},
	})
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newStatusFactory(ios, srv)

	cmd := execute.NewStatusCmd(f)
	cmd.SetArgs([]string{"exec-unverified", "--require-verified"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unverified receipt, got nil")
	}
	if !strings.Contains(err.Error(), "0xdead") {
		t.Errorf("expected offending hash in error, got: %q", err.Error())
	}
}

func TestExecStatusCmd_RequireVerified_FailsWhenReceiptNotSuccess(t *testing.T) {
	for _, rs := range []string{"reverted", "not_found", "timeout", "safe_inner_failure"} {
		t.Run(rs, func(t *testing.T) {
			srv := serveStatus(t, execute.ExecStatusResponse{
				ExecutionID: "exec-" + rs,
				Status:      "completed",
				Receipts: []execute.ExecReceipt{{
					Hash:          "0xbeef",
					ChainID:       84532,
					Verified:      true,
					ReceiptStatus: rs,
				}},
			})
			defer srv.Close()

			ios, _, _, _ := iostreams.Test()
			f := newStatusFactory(ios, srv)

			cmd := execute.NewStatusCmd(f)
			cmd.SetArgs([]string{"exec-" + rs, "--require-verified"})

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error for receiptStatus=%s, got nil", rs)
			}
			if !strings.Contains(err.Error(), rs) {
				t.Errorf("expected receiptStatus %q in error, got: %q", rs, err.Error())
			}
		})
	}
}

func TestExecStatusCmd_WithoutRequireVerified_CompletedWithoutReceiptsStillPasses(t *testing.T) {
	srv := serveStatus(t, execute.ExecStatusResponse{
		ExecutionID: "exec-backcompat",
		Status:      "completed",
	})
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newStatusFactory(ios, srv)

	cmd := execute.NewStatusCmd(f)
	cmd.SetArgs([]string{"exec-backcompat"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected back-compat success without --require-verified, got: %v", err)
	}
}

func TestExecStatusCmd_Watch_RequireVerified_PassesOnVerifiedTerminal(t *testing.T) {
	callCount := 0
	txHash := "0xwatched"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var resp execute.ExecStatusResponse
		if callCount >= 2 {
			resp = execute.ExecStatusResponse{
				ExecutionID:     "exec-watch-verified",
				Status:          "completed",
				TransactionHash: &txHash,
				Receipts:        []execute.ExecReceipt{verifiedReceipt("0xwatched")},
			}
		} else {
			resp = execute.ExecStatusResponse{
				ExecutionID: "exec-watch-verified",
				Status:      "pending",
			}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newStatusFactory(ios, srv)

	cmd := execute.NewStatusCmd(f)
	cmd.SetArgs([]string{"exec-watch-verified", "--watch", "--require-verified"})

	done := make(chan error, 1)
	go func() {
		done <- cmd.Execute()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("command timed out")
	}

	if !strings.Contains(buf.String(), "0xwatched") {
		t.Errorf("expected receipt hash in output, got: %q", buf.String())
	}
}

func TestExecStatusCmd_Watch_TimeoutExpires(t *testing.T) {
	srv := serveStatus(t, execute.ExecStatusResponse{
		ExecutionID: "exec-stuck",
		Status:      "pending",
	})
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newStatusFactory(ios, srv)

	cmd := execute.NewStatusCmd(f)
	cmd.SetArgs([]string{"exec-stuck", "--watch", "--timeout", "100ms"})

	done := make(chan error, 1)
	go func() {
		done <- cmd.Execute()
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected timeout error, got nil")
		}
		if !strings.Contains(err.Error(), "timeout") {
			t.Errorf("expected 'timeout' in error, got: %q", err.Error())
		}
	case <-time.After(10 * time.Second):
		t.Fatal("watch did not respect --timeout; command still running")
	}
}
