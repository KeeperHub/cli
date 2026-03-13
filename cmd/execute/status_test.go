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

func newStatusFactory(ios *iostreams.IOStreams, srv *httptest.Server) *cmdutil.Factory {
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

func TestExecStatusCmd_ShowsStatusTable(t *testing.T) {
	txHash := "0xtxhash123"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := execute.ExecStatusResponse{
			ExecutionID:     "exec-abc",
			Status:          "completed",
			Type:            "transfer",
			TransactionHash: &txHash,
			CreatedAt:       "2026-03-13T00:00:00Z",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newStatusFactory(ios, srv)

	cmd := execute.NewStatusCmd(f)
	cmd.SetArgs([]string{"exec-abc"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "exec-abc") {
		t.Errorf("expected execution ID in output, got: %q", out)
	}
	if !strings.Contains(out, "completed") {
		t.Errorf("expected status in output, got: %q", out)
	}
	if !strings.Contains(out, "0xtxhash123") {
		t.Errorf("expected tx hash in output, got: %q", out)
	}
}

func TestExecStatusCmd_JSONOutput(t *testing.T) {
	txHash := "0xtxhash456"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := execute.ExecStatusResponse{
			ExecutionID:     "exec-json",
			Status:          "completed",
			TransactionHash: &txHash,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newStatusFactory(ios, srv)

	cmd := execute.NewStatusCmd(f)
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.SetArgs([]string{"exec-json", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"executionId"`) {
		t.Errorf("expected JSON with executionId, got: %q", out)
	}
	if !strings.Contains(out, "exec-json") {
		t.Errorf("expected execution ID in JSON, got: %q", out)
	}
}

func TestExecStatusCmd_FailedStatus_ReturnsError(t *testing.T) {
	errMsg := "insufficient funds"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := execute.ExecStatusResponse{
			ExecutionID: "exec-failed",
			Status:      "failed",
			Error:       &errMsg,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := newStatusFactory(ios, srv)

	cmd := execute.NewStatusCmd(f)
	cmd.SetArgs([]string{"exec-failed"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed execution, got nil")
	}
	if !strings.Contains(err.Error(), "insufficient funds") {
		t.Errorf("expected error message to contain 'insufficient funds', got: %q", err.Error())
	}

	out := buf.String()
	if !strings.Contains(out, "failed") {
		t.Errorf("expected 'failed' in output, got: %q", out)
	}
	if !strings.Contains(out, "insufficient funds") {
		t.Errorf("expected error message in output, got: %q", out)
	}
}

func TestExecStatusCmd_Watch_PollsUntilTerminal(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		txHash := "0xtxhash789"
		var resp execute.ExecStatusResponse
		if callCount >= 2 {
			resp = execute.ExecStatusResponse{
				ExecutionID:     "exec-watch",
				Status:          "completed",
				TransactionHash: &txHash,
			}
		} else {
			resp = execute.ExecStatusResponse{
				ExecutionID: "exec-watch",
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
	cmd.SetArgs([]string{"exec-watch", "--watch"})

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

	if callCount < 2 {
		t.Errorf("expected at least 2 status polls, got %d", callCount)
	}

	out := buf.String()
	if !strings.Contains(out, "0xtxhash789") {
		t.Errorf("expected tx hash in final output, got: %q", out)
	}
}
