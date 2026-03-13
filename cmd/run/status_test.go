package run_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/run"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func makeStatusServer(t *testing.T, responses []map[string]any) *httptest.Server {
	t.Helper()
	callCount := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/status") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		idx := callCount
		if idx >= len(responses) {
			idx = len(responses) - 1
		}
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responses[idx]); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
}

func makeStatusServerWith401(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
}

func makeRunFactory(ios *iostreams.IOStreams, host string) *cmdutil.Factory {
	return &cmdutil.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return config.Config{DefaultHost: host}, nil
		},
		HTTPClient: func() (*khhttp.Client, error) {
			return khhttp.NewClient(khhttp.ClientOptions{
				Host:      host,
				IOStreams:  ios,
			}), nil
		},
	}
}

func TestStatusCmd_SingleShot(t *testing.T) {
	srv := makeStatusServer(t, []map[string]any{
		{
			"status": "success",
			"nodeStatuses": []map[string]any{
				{"nodeId": "node1", "status": "success"},
			},
			"progress": map[string]any{
				"totalSteps":      3,
				"completedSteps":  3,
				"runningSteps":    0,
				"currentNodeId":   nil,
				"currentNodeName": nil,
				"percentage":      100,
			},
			"errorContext": nil,
		},
	})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewStatusCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "success") {
		t.Errorf("expected 'success' in output, got: %q", out)
	}
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	srv := makeStatusServer(t, []map[string]any{
		{
			"status": "running",
			"nodeStatuses": []map[string]any{},
			"progress": map[string]any{
				"totalSteps":     2,
				"completedSteps": 1,
				"runningSteps":   1,
				"percentage":     50,
			},
			"errorContext": nil,
		},
	})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewStatusCmd(f)
	// Need to add json flag inherited from parent - add it directly
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().String("jq", "", "")
	cmd.SetArgs([]string{"run-abc123", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"status"`) {
		t.Errorf("expected JSON with 'status' field, got: %q", out)
	}
	if !strings.Contains(out, "running") {
		t.Errorf("expected 'running' in JSON output, got: %q", out)
	}
}

func TestStatusCmd_WatchSucceeds(t *testing.T) {
	srv := makeStatusServer(t, []map[string]any{
		{
			"status": "running",
			"nodeStatuses": []map[string]any{},
			"progress": map[string]any{
				"totalSteps":      2,
				"completedSteps":  1,
				"runningSteps":    1,
				"currentNodeId":   strPtr("node1"),
				"currentNodeName": strPtr("Transfer ETH"),
				"percentage":      50,
			},
			"errorContext": nil,
		},
		{
			"status": "success",
			"nodeStatuses": []map[string]any{},
			"progress": map[string]any{
				"totalSteps":      2,
				"completedSteps":  2,
				"runningSteps":    0,
				"currentNodeId":   nil,
				"currentNodeName": nil,
				"percentage":      100,
			},
			"errorContext": nil,
		},
	})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewStatusCmd(f)
	cmd.SetArgs([]string{"run-abc123", "--watch"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "success") {
		t.Errorf("expected 'success' in watch output, got: %q", out)
	}
}

func TestStatusCmd_WatchError(t *testing.T) {
	srv := makeStatusServer(t, []map[string]any{
		{
			"status": "error",
			"nodeStatuses": []map[string]any{},
			"progress": map[string]any{
				"totalSteps":     1,
				"completedSteps": 0,
				"runningSteps":   0,
				"percentage":     0,
			},
			"errorContext": "RPC timeout after 30s",
		},
	})
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewStatusCmd(f)
	cmd.SetArgs([]string{"run-abc123", "--watch"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for error status in watch mode, got nil")
	}
}

func TestStatusCmd_WatchNonTTY(t *testing.T) {
	srv := makeStatusServer(t, []map[string]any{
		{
			"status": "running",
			"nodeStatuses": []map[string]any{},
			"progress": map[string]any{
				"totalSteps":      2,
				"completedSteps":  1,
				"runningSteps":    1,
				"currentNodeId":   strPtr("node1"),
				"currentNodeName": strPtr("My Step"),
				"percentage":      50,
			},
			"errorContext": nil,
		},
		{
			"status": "success",
			"nodeStatuses": []map[string]any{},
			"progress": map[string]any{
				"totalSteps":     2,
				"completedSteps": 2,
				"runningSteps":   0,
				"percentage":     100,
			},
			"errorContext": nil,
		},
	})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	// Non-TTY: Out is a bytes.Buffer, not *os.File, so IsTerminal() returns false
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewStatusCmd(f)
	cmd.SetArgs([]string{"run-abc123", "--watch"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	// In non-TTY mode, each update should be on its own line
	if !strings.Contains(out, "\n") {
		t.Errorf("expected newlines in non-TTY watch output, got: %q", out)
	}
}

func TestStatusCmd_WatchJSONMode(t *testing.T) {
	srv := makeStatusServer(t, []map[string]any{
		{
			"status": "running",
			"nodeStatuses": []map[string]any{},
			"progress": map[string]any{
				"totalSteps":     1,
				"completedSteps": 0,
				"runningSteps":   1,
				"percentage":     0,
			},
			"errorContext": nil,
		},
		{
			"status": "success",
			"nodeStatuses": []map[string]any{},
			"progress": map[string]any{
				"totalSteps":     1,
				"completedSteps": 1,
				"runningSteps":   0,
				"percentage":     100,
			},
			"errorContext": nil,
		},
	})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewStatusCmd(f)
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().String("jq", "", "")
	cmd.SetArgs([]string{"run-abc123", "--watch", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	// JSON mode: should output valid JSON for final result only
	if !strings.Contains(out, `"status"`) {
		t.Errorf("expected JSON output in watch+json mode, got: %q", out)
	}
	if !strings.Contains(out, "success") {
		t.Errorf("expected 'success' in watch+json output, got: %q", out)
	}
}

func TestStatusCmd_401AuthHint(t *testing.T) {
	srv := makeStatusServerWith401(t)
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewStatusCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "kh auth login") {
		t.Errorf("expected auth hint in error, got: %q", err.Error())
	}
}

func strPtr(s string) *string { return &s }
