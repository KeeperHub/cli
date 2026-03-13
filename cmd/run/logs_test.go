package run_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/run"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func makeLogsServer(t *testing.T, response map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/logs") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
}

func makeLogsWith401(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
}

func TestLogsCmd_RendersStepTable(t *testing.T) {
	durStr := "1234"
	errStr := (*string)(nil)
	completedAt := "2026-03-13T00:05:00Z"

	srv := makeLogsServer(t, map[string]any{
		"execution": map[string]any{"id": "run-abc123"},
		"logs": []map[string]any{
			{
				"id":          "log-1",
				"nodeId":      "node-1",
				"nodeName":    "Transfer ETH",
				"nodeType":    "action",
				"status":      "success",
				"input":       map[string]any{"amount": "0.1"},
				"output":      map[string]any{"txHash": "0xabc"},
				"error":       errStr,
				"startedAt":   "2026-03-13T00:04:58Z",
				"completedAt": completedAt,
				"duration":    durStr,
			},
		},
	})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewLogsCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Transfer ETH") {
		t.Errorf("expected node name in output, got: %q", out)
	}
	if !strings.Contains(out, "success") {
		t.Errorf("expected status in output, got: %q", out)
	}
}

func TestLogsCmd_JSONOutput(t *testing.T) {
	srv := makeLogsServer(t, map[string]any{
		"execution": map[string]any{"id": "run-abc123"},
		"logs":      []map[string]any{},
	})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewLogsCmd(f)
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().String("jq", "", "")
	cmd.SetArgs([]string{"run-abc123", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"logs"`) {
		t.Errorf("expected JSON with 'logs' field, got: %q", out)
	}
}

func TestLogsCmd_EmptyLogs(t *testing.T) {
	srv := makeLogsServer(t, map[string]any{
		"execution": map[string]any{"id": "run-abc123"},
		"logs":      []map[string]any{},
	})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewLogsCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No logs found.") {
		t.Errorf("expected empty logs message, got: %q", out)
	}
}

func TestLogsCmd_401AuthHint(t *testing.T) {
	srv := makeLogsWith401(t)
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewLogsCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "kh auth login") {
		t.Errorf("expected auth hint in error, got: %q", err.Error())
	}
}
