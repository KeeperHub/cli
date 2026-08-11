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
			_, _ = w.Write([]byte(`{"error":"Execution not found","code":"not_found"}`))
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
