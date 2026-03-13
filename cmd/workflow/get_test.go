package workflow_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/workflow"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWFGetFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
	}
}

func makeGetServer(t *testing.T, workflowID string, detail map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/workflows/"+workflowID {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(detail)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	}))
}

func TestGetCmd_SendsGETWorkflowByID(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/workflows/wf-123" {
			called = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":        "wf-123",
				"name":      "Test WF",
				"enabled":   true,
				"visibility": "private",
				"createdAt": "2026-01-01T00:00:00Z",
				"updatedAt": "2026-01-01T00:00:00Z",
				"nodes":     []interface{}{},
				"edges":     []interface{}{},
			})
		} else {
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newWFGetFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"get", "wf-123"})
	err := wfCmd.Execute()
	require.NoError(t, err)
	assert.True(t, called, "expected GET /api/workflows/wf-123 to be called")
}

func TestGetCmd_RendersDetailOutput(t *testing.T) {
	detail := map[string]interface{}{
		"id":         "wf-123",
		"name":       "Token Transfer",
		"enabled":    true,
		"visibility": "private",
		"createdAt":  "2026-01-01T00:00:00Z",
		"updatedAt":  "2026-02-15T00:00:00Z",
		"nodes":      []interface{}{"n1", "n2", "n3"},
		"edges":      []interface{}{"e1", "e2"},
	}
	server := makeGetServer(t, "wf-123", detail)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFGetFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"get", "wf-123"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "wf-123", "expected workflow ID in output")
	assert.Contains(t, out, "Token Transfer", "expected workflow name in output")
	assert.Contains(t, out, "active", "expected status in output")
	assert.Contains(t, out, "private", "expected visibility in output")
	assert.Contains(t, out, "3", "expected node count in output")
	assert.Contains(t, out, "2", "expected edge count in output")
}

func TestGetCmd_JSONOutput(t *testing.T) {
	detail := map[string]interface{}{
		"id":         "wf-123",
		"name":       "Token Transfer",
		"enabled":    true,
		"visibility": "private",
		"createdAt":  "2026-01-01T00:00:00Z",
		"updatedAt":  "2026-01-01T00:00:00Z",
		"nodes":      []interface{}{},
		"edges":      []interface{}{},
	}
	server := makeGetServer(t, "wf-123", detail)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFGetFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	wfCmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")
	wfCmd.SetArgs([]string{"get", "wf-123", "--json"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, `"id"`)
	assert.Contains(t, out, "wf-123")
}

func TestGetCmd_NotFoundReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"workflow not found"}`))
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newWFGetFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"get", "nonexistent-id"})
	err := wfCmd.Execute()
	require.Error(t, err)

	var notFound cmdutil.NotFoundError
	require.ErrorAs(t, err, &notFound, "expected NotFoundError for 404 response")
	assert.Equal(t, 2, cmdutil.ExitCodeForError(err), "expected exit code 2 for NotFoundError")
}

func TestGetCmd_RequiresWorkflowID(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
	}

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"get"})
	err := wfCmd.Execute()
	assert.Error(t, err, "get without ID should return error")
}

func TestGetCmd_AliasGResolvesCorrectly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/workflows/wf-abc" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         "wf-abc",
				"name":       "Aliased",
				"enabled":    false,
				"visibility": "public",
				"createdAt":  "2026-01-01T00:00:00Z",
				"updatedAt":  "2026-01-01T00:00:00Z",
				"nodes":      []interface{}{},
				"edges":      []interface{}{},
			})
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFGetFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"g", "wf-abc"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.True(t, strings.Contains(out, "wf-abc") || strings.Contains(out, "Aliased"),
		"expected workflow detail in output via alias 'g'")
}
