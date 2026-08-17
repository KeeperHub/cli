package execute_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/keeperhub/cli/cmd/execute"
	"github.com/keeperhub/cli/internal/execrecovery"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func TestTransferCmd_SendsIdempotencyKey(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get(execrecovery.IdempotencyHeader)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-idem","status":"completed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--idempotency-key", "stable-intent-1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotKey != "stable-intent-1" {
		t.Fatalf("Idempotency-Key=%q, want stable-intent-1", gotKey)
	}
}

func TestTransferCmd_IdempotencyKeyStableAcrossHTTPRetries(t *testing.T) {
	var keys []string
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		keys = append(keys, r.Header.Get(execrecovery.IdempotencyHeader))
		if n == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-retry","status":"completed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--idempotency-key", "retry-stable"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() < 2 {
		t.Fatalf("expected HTTP retry, got %d calls", calls.Load())
	}
	for i, k := range keys {
		if k != "retry-stable" {
			t.Fatalf("call %d Idempotency-Key=%q, want retry-stable", i, k)
		}
	}
}

func TestTransferCmd_WaitToleratesInitialNotFound(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-cold","status":"pending"}`))
			return
		}
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Execution not found"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"executionId":"exec-cold","status":"completed","transactionHash":"0xabc"}`))
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--wait", "--timeout", "15s"})

	start := time.Now()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) > 14*time.Second {
		t.Fatal("cold-start wait took too long")
	}
	out := buf.String()
	if !strings.Contains(out, "exec-cold") {
		t.Fatalf("expected execution in output, got %q", out)
	}
	if calls.Load() < 2 {
		t.Fatalf("expected cold-start poll, got %d status calls", calls.Load())
	}
}

func TestTransferCmd_WaitFailsOnRevertedReceipt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-rev","status":"pending"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"executionId":     "exec-rev",
			"status":          "completed",
			"transactionHash": "0xrev",
			"receipts": []map[string]any{
				{"hash": "0xrev", "chainId": 8453, "verified": true, "receiptStatus": "reverted"},
			},
		})
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--wait", "--timeout", "10s"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected reverted receipt to fail")
	}
	if !strings.Contains(err.Error(), "reverted") {
		t.Fatalf("expected reverted error, got %v", err)
	}
}

func TestTransferCmd_IdempotencyKeyStableAcross504(t *testing.T) {
	var keys []string
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		keys = append(keys, r.Header.Get(execrecovery.IdempotencyHeader))
		if n == 1 {
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-504","status":"completed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--idempotency-key", "retry-504"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() < 2 {
		t.Fatalf("expected HTTP retry, got %d calls", calls.Load())
	}
	for i, k := range keys {
		if k != "retry-504" {
			t.Fatalf("call %d Idempotency-Key=%q, want retry-504", i, k)
		}
	}
}

func TestTransferCmd_IdempotencyInProgressRetriesSameKey(t *testing.T) {
	var keys []string
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		keys = append(keys, r.Header.Get(execrecovery.IdempotencyHeader))
		w.Header().Set("Content-Type", "application/json")
		if n < 3 {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":"A request with this Idempotency-Key is already being processed. Retry the same key shortly; do not rotate it.","code":"idempotency_in_progress","retryable":true}`))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-inprog","status":"completed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--idempotency-key", "in-progress-key", "--timeout", "10s"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 3 {
		t.Fatalf("got %d POSTs, want 3", calls.Load())
	}
	for i, k := range keys {
		if k != "in-progress-key" {
			t.Fatalf("call %d minted a new key %q", i, k)
		}
	}
}

func TestTransferCmd_504ThenInProgressReusesKey(t *testing.T) {
	var keys []string
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		keys = append(keys, r.Header.Get(execrecovery.IdempotencyHeader))
		w.Header().Set("Content-Type", "application/json")
		switch n {
		case 1:
			w.WriteHeader(http.StatusGatewayTimeout)
		case 2:
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":"in flight","code":"idempotency_in_progress","retryable":true}`))
		default:
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-combo","status":"completed"}`))
		}
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--idempotency-key", "combo-key", "--timeout", "15s"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() < 3 {
		t.Fatalf("got %d calls, want at least 3 (504 retry + in_progress + 202)", calls.Load())
	}
	for i, k := range keys {
		if k != "combo-key" {
			t.Fatalf("call %d key=%q", i, k)
		}
	}
}

func TestTransferCmd_IdempotencyConflictFailsWithoutNewKey(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"Idempotency-Key was reused with a different request payload. Use a new key for a different request.","code":"idempotency_conflict","originalExecutionId":"exec-orig","retryable":false}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--idempotency-key", "conflict-key"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "idempotency") && !strings.Contains(err.Error(), "different request payload") {
		t.Fatalf("got %v", err)
	}
	if strings.Contains(err.Error(), "rotate") && strings.Contains(err.Error(), "new key for a different") {
		// server message mentions a new key for a *different* request; we must still refuse auto-rotate
	}
	if !strings.Contains(err.Error(), "do not retry with a new key") {
		t.Fatalf("conflict must tell the user not to mint a new key: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("conflict must not retry the POST, got %d", calls.Load())
	}
}

func TestTransferCmd_WaitPersistent404TimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-missing","status":"pending"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Execution not found"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--wait", "--timeout", "3s"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected timeout")
	}
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("got %v", err)
	}
}

// An unreadable receipt must not become a non-zero exit: that is what makes a
// caller re-run and broadcast a second transaction for an intent that may
// already be on chain.
func TestTransferCmd_WaitStopsOnUnconfirmedAndExitsZero(t *testing.T) {
	var statusReads int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-unconf","status":"pending"}`))
			return
		}
		atomic.AddInt32(&statusReads, 1)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"executionId":     "exec-unconf",
			"status":          "unconfirmed",
			"transactionHash": "0xunconf",
			"receipts": []map[string]any{
				{"hash": "0xunconf", "verified": false, "receiptStatus": "not_found", "verifiedAt": "2026-08-11T00:00:00Z"},
			},
		})
	}))
	defer srv.Close()

	ios, out, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--wait", "--timeout", "10s"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unconfirmed must exit zero, got %v", err)
	}
	if got := atomic.LoadInt32(&statusReads); got != 1 {
		t.Fatalf("status reads=%d, want 1 (unconfirmed must not be polled through)", got)
	}
	if s := out.String(); !strings.Contains(s, "unconfirmed") || !strings.Contains(s, "0xunconf") {
		t.Fatalf("expected status and hash in output, got %q", s)
	}
}

func TestTransferCmd_WaitFailsOnSafeInnerFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-safe","status":"pending"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"executionId":     "exec-safe",
			"status":          "completed",
			"transactionHash": "0xsafe",
			"receipts": []map[string]any{
				{"hash": "0xsafe", "verified": false, "receiptStatus": "safe_inner_failure", "verifiedAt": "2026-08-11T00:00:00Z"},
			},
		})
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)
	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--wait", "--timeout", "10s"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected safe_inner_failure to fail")
	}
	if !strings.Contains(err.Error(), "safe_inner_failure") {
		t.Fatalf("got %v", err)
	}
}
