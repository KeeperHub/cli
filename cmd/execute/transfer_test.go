package execute_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/keeperhub/cli/cmd/execute"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func newTransferFactory(ios *iostreams.IOStreams, srv *httptest.Server) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "test",
		IOStreams:   ios,
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

func TestTransferCmd_SendsCorrectFields(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-123","status":"completed"}`))
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)

	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["network"] != "1" {
		t.Errorf("expected network=1, got %v", gotBody["network"])
	}
	if gotBody["recipientAddress"] != "0xabc" {
		t.Errorf("expected recipientAddress=0xabc, got %v", gotBody["recipientAddress"])
	}
	if gotBody["amount"] != "0.1" {
		t.Errorf("expected amount=0.1, got %v", gotBody["amount"])
	}
	if _, ok := gotBody["tokenAddress"]; ok {
		t.Error("tokenAddress should not be present for ETH transfer")
	}

	out := buf.String()
	if !strings.Contains(out, "exec-123") {
		t.Errorf("expected execution ID in output, got: %q", out)
	}
	if !strings.Contains(out, "completed") {
		t.Errorf("expected status in output, got: %q", out)
	}
}

func TestTransferCmd_RequiredFlags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()

	tests := []struct {
		name string
		args []string
	}{
		{"missing chain", []string{"--to", "0xabc", "--amount", "0.1"}},
		{"missing to", []string{"--chain", "1", "--amount", "0.1"}},
		{"missing amount", []string{"--chain", "1", "--to", "0xabc"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newTransferFactory(ios, srv)
			cmd := execute.NewTransferCmd(f)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error for missing required flag, got nil")
			}
		})
	}
}

func TestTransferCmd_TokenAddress(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-456","status":"completed"}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)

	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "100", "--token-address", "0xtoken"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["tokenAddress"] != "0xtoken" {
		t.Errorf("expected tokenAddress=0xtoken, got %v", gotBody["tokenAddress"])
	}
}

func TestTransferCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-789","status":"completed"}`))
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)

	cmd := execute.NewTransferCmd(f)
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"executionId"`) {
		t.Errorf("expected JSON output with executionId, got: %q", out)
	}
}

func TestTransferCmd_WaitAlreadyTerminal(t *testing.T) {
	pollCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status") {
			pollCount++
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-done","status":"completed","transactionHash":"0xtxhash"}`))
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)

	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--wait"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pollCount > 0 {
		t.Errorf("expected no polling when already terminal, got %d polls", pollCount)
	}

	out := buf.String()
	if !strings.Contains(out, "0xtxhash") {
		t.Errorf("expected tx hash in output, got: %q", out)
	}
}

func TestTransferCmd_WaitPolls(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/status") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"executionId":"exec-poll","status":"completed","transactionHash":"0xtxhash"}`))
		} else {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-poll","status":"pending"}`))
		}
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newTransferFactory(ios, srv)

	cmd := execute.NewTransferCmd(f)
	cmd.SetArgs([]string{"--chain", "1", "--to", "0xabc", "--amount", "0.1", "--wait", "--timeout", "10s"})

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

