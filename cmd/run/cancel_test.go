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

func makeCancelServer(t *testing.T, statusCode int, response map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/cancel") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("failed to encode response: %v", err)
			}
		}
	}))
}

// TestCancelCmd_WithYesFlag verifies --yes skips confirmation and cancels.
// Note: in implemented cancel command, --yes is inherited from root as persistent flag.
// For unit tests, we verify non-TTY auto-proceeds (same effect as --yes).
func TestCancelCmd_WithYesFlag(t *testing.T) {
	srv := makeCancelServer(t, http.StatusOK, map[string]any{"success": true})
	defer srv.Close()

	// Non-TTY auto-proceeds without prompt, equivalent to --yes
	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewCancelCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "cancelled") {
		t.Errorf("expected 'cancelled' in output, got: %q", out)
	}
}

func TestCancelCmd_PromptYes(t *testing.T) {
	srv := makeCancelServer(t, http.StatusOK, map[string]any{"success": true})
	defer srv.Close()

	ios, buf, _, inBuf := iostreams.Test()
	inBuf.WriteString("y\n")
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewCancelCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "cancelled") {
		t.Errorf("expected 'cancelled' in output, got: %q", out)
	}
}

func TestCancelCmd_PromptNo(t *testing.T) {
	srv := makeCancelServer(t, http.StatusOK, map[string]any{"success": true})
	defer srv.Close()

	ios, _, _, inBuf := iostreams.Test()
	inBuf.WriteString("n\n")
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewCancelCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	err := cmd.Execute()
	// Answering "n" returns CancelError which exits 0 (not nil error but CancelError)
	if err == nil {
		t.Fatal("expected CancelError when user says n, got nil")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected 'cancelled' in error, got: %q", err.Error())
	}
}

func TestCancelCmd_NonTTYAutoProceeds(t *testing.T) {
	srv := makeCancelServer(t, http.StatusOK, map[string]any{"success": true})
	defer srv.Close()

	// Non-TTY: Out is bytes.Buffer so IsTerminal() is false
	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewCancelCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error in non-TTY auto-proceed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "cancelled") {
		t.Errorf("expected 'cancelled' in output, got: %q", out)
	}
}

func TestCancelCmd_JSONOutput(t *testing.T) {
	srv := makeCancelServer(t, http.StatusOK, map[string]any{"success": true})
	defer srv.Close()

	ios, buf, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewCancelCmd(f)
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().String("jq", "", "")
	cmd.SetArgs([]string{"run-abc123", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"success"`) {
		t.Errorf("expected JSON with 'success' field, got: %q", out)
	}
	if !strings.Contains(out, "true") {
		t.Errorf("expected true in JSON output, got: %q", out)
	}
}

func TestCancelCmd_401AuthHint(t *testing.T) {
	srv := makeCancelServer(t, http.StatusUnauthorized, nil)
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := makeRunFactory(ios, srv.URL)

	cmd := run.NewCancelCmd(f)
	cmd.SetArgs([]string{"run-abc123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "kh auth login") {
		t.Errorf("expected auth hint in error, got: %q", err.Error())
	}
}
