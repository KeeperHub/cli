package execute_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/keeperhub/cli/cmd/execute"
	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/internal/execrecovery"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func newContractCallFactory(ios *iostreams.IOStreams, srv *httptest.Server) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "test",
		IOStreams:  ios,
	})
	return &cmdutil.Factory{
		IOStreams: ios,
		HTTPClient: func() (*khhttp.Client, error) {
			return client, nil
		},
		Config: func() (config.Config, error) {
			return config.Config{DefaultHost: srv.URL}, nil
		},
	}
}

func TestContractCallCmd_SendsCorrectFields(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"100000000000000000000"}`))
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)

	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "balanceOf"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["network"] != "1" {
		t.Errorf("expected network=1, got %v", gotBody["network"])
	}
	if gotBody["contractAddress"] != "0xcontract" {
		t.Errorf("expected contractAddress=0xcontract, got %v", gotBody["contractAddress"])
	}
	if gotBody["functionName"] != "balanceOf" {
		t.Errorf("expected functionName=balanceOf, got %v", gotBody["functionName"])
	}

	out := buf.String()
	if !strings.Contains(out, "100000000000000000000") {
		t.Errorf("expected result in output, got: %q", out)
	}
}

func TestContractCallCmd_RequiredFlags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()

	tests := []struct {
		name string
		args []string
	}{
		{"missing chain", []string{"--contract", "0xcontract", "--method", "balanceOf"}},
		{"missing contract", []string{"--chain", "1", "--method", "balanceOf"}},
		{"missing method", []string{"--chain", "1", "--contract", "0xcontract"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newContractCallFactory(ios, srv)
			cmd := execute.NewContractCallCmd(f)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error for missing required flag, got nil")
			}
		})
	}
}

func TestContractCallCmd_ArgsStringType(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"42"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)

	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{
		"--chain", "1",
		"--contract", "0xcontract",
		"--method", "balanceOf",
		"--args", `["0xdef"]`,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["functionArgs"] != `["0xdef"]` {
		t.Errorf("expected functionArgs=[\"0xdef\"], got %v", gotBody["functionArgs"])
	}
}

func TestContractCallCmd_ABIFile(t *testing.T) {
	abiContent := `[{"type":"function","name":"balanceOf","inputs":[{"name":"account","type":"address"}],"outputs":[{"name":"","type":"uint256"}]}]`

	abiFile, err := os.CreateTemp("", "abi-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer os.Remove(abiFile.Name())

	if _, err := abiFile.WriteString(abiContent); err != nil {
		t.Fatalf("writing abi file: %v", err)
	}
	abiFile.Close()

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"42"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)

	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{
		"--chain", "1",
		"--contract", "0xcontract",
		"--method", "balanceOf",
		"--abi-file", abiFile.Name(),
	})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["abi"] != abiContent {
		t.Errorf("expected abi to be file content, got %v", gotBody["abi"])
	}
}

func TestContractCallCmd_ReadResponse200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"100"}`))
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)

	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "balanceOf"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "100") {
		t.Errorf("expected result value in output, got: %q", out)
	}
	if strings.Contains(out, "executionId") {
		t.Errorf("read-only call should not print executionId, got: %q", out)
	}
}

func TestContractCallCmd_WriteResponse202(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-write","status":"completed"}`))
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)

	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "transfer"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "exec-write") {
		t.Errorf("expected execution ID in output, got: %q", out)
	}
}

func TestContractCallCmd_WaitReadOnlyIsNoop(t *testing.T) {
	pollCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status") {
			pollCount++
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"42"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)

	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "balanceOf", "--wait"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pollCount > 0 {
		t.Errorf("expected no polling for read-only call, got %d polls", pollCount)
	}
}

func TestContractCallCmd_WaitWritePolls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/status") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"executionId":"exec-write","status":"completed","transactionHash":"0xtxhash"}`))
		} else {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-write","status":"pending"}`))
		}
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)

	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "transfer", "--wait", "--timeout", "10s"})

	done := make(chan error, 1)
	go func() {
		done <- cmd.Execute()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("command timed out")
	}

	out := buf.String()
	if !strings.Contains(out, "0xtxhash") {
		t.Errorf("expected tx hash in output, got: %q", out)
	}
}

func TestContractCallCmd_SendsIdempotencyKey(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get(execrecovery.IdempotencyHeader)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-cc-idem","status":"completed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)
	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "transfer", "--idempotency-key", "stable-cc-1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotKey != "stable-cc-1" {
		t.Fatalf("Idempotency-Key=%q, want stable-cc-1", gotKey)
	}
}

func TestContractCallCmd_IdempotencyKeyStableAcrossHTTPRetries(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"executionId":"exec-cc-retry","status":"completed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)
	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "transfer", "--idempotency-key", "retry-stable-cc"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() < 2 {
		t.Fatalf("expected HTTP retry, got %d calls", calls.Load())
	}
	for i, k := range keys {
		if k != "retry-stable-cc" {
			t.Fatalf("call %d Idempotency-Key=%q, want retry-stable-cc", i, k)
		}
	}
}

func TestContractCallCmd_IdempotencyKeyStableAcross504(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"executionId":"exec-cc-504","status":"completed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)
	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "transfer", "--idempotency-key", "cc-504"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() < 2 {
		t.Fatalf("expected HTTP retry, got %d calls", calls.Load())
	}
	for i, k := range keys {
		if k != "cc-504" {
			t.Fatalf("call %d key=%q", i, k)
		}
	}
}

func TestContractCallCmd_IdempotencyInProgressRetriesSameKey(t *testing.T) {
	var keys []string
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		keys = append(keys, r.Header.Get(execrecovery.IdempotencyHeader))
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":"in flight","code":"idempotency_in_progress","retryable":true}`))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-cc-inprog","status":"completed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)
	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "transfer", "--idempotency-key", "cc-inprog", "--timeout", "10s"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("got %d POSTs, want 2", calls.Load())
	}
	for i, k := range keys {
		if k != "cc-inprog" {
			t.Fatalf("call %d key=%q", i, k)
		}
	}
}

func TestContractCallCmd_IdempotencyConflictFails(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"Idempotency-Key was reused with a different request payload.","code":"idempotency_conflict","retryable":false}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)
	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "transfer", "--idempotency-key", "cc-conflict"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected conflict")
	}
	if !strings.Contains(err.Error(), "do not retry with a new key") {
		t.Fatalf("got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("got %d POSTs, want 1", calls.Load())
	}
}

func TestContractCallCmd_WaitFailsWhenWriteResponseAlreadyFailed(t *testing.T) {
	pollCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/status") {
			pollCount++
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-ccfail","status":"failed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newContractCallFactory(ios, srv)

	cmd := execute.NewContractCallCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--contract", "0xcontract", "--method", "transfer", "--wait", "--timeout", "10s"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error when the write response is already failed, got nil")
	}
	if !strings.Contains(err.Error(), "exec-ccfail") {
		t.Errorf("expected the execution id in the error, got: %q", err.Error())
	}
	if pollCount > 0 {
		t.Errorf("expected no polling when already terminal, got %d polls", pollCount)
	}
}
