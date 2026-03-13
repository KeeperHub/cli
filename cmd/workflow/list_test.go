package workflow_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/workflow"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWFListFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeWorkflowsServer(t *testing.T, workflows []map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/workflows" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(workflows)
	}))
}

func TestListCmd_SendsGETWorkflows(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/workflows" {
			called = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]interface{}{})
		} else {
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls"})
	err := wfCmd.Execute()
	require.NoError(t, err)
	assert.True(t, called, "expected GET /api/workflows to be called")
}

func TestListCmd_RendersTableWithColumns(t *testing.T) {
	workflows := []map[string]interface{}{
		{
			"id":         "wf-001",
			"name":       "My Workflow",
			"enabled":    true,
			"visibility": "private",
			"createdAt":  "2026-01-01T00:00:00Z",
			"updatedAt":  "2026-02-01T00:00:00Z",
		},
	}
	server := makeWorkflowsServer(t, workflows)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "wf-001", "expected workflow ID in output")
	assert.Contains(t, out, "My Workflow", "expected workflow name in output")
	assert.Contains(t, out, "active", "expected 'active' status for enabled workflow")
	assert.Contains(t, out, "private", "expected visibility in output")
}

func TestListCmd_LimitSendsQueryParam(t *testing.T) {
	var gotLimit string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]interface{}{})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls", "--limit", "5"})
	err := wfCmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "5", gotLimit, "expected limit=5 query param")
}

func TestListCmd_JSONOutput(t *testing.T) {
	workflows := []map[string]interface{}{
		{"id": "wf-001", "name": "Alpha", "enabled": true, "visibility": "private", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z"},
		{"id": "wf-002", "name": "Beta", "enabled": false, "visibility": "public", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z"},
	}
	server := makeWorkflowsServer(t, workflows)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls", "--json"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, `"id"`, "expected JSON with id field")
	assert.Contains(t, out, "wf-001")
	assert.Contains(t, out, "wf-002")
}

func TestListCmd_JQFilter(t *testing.T) {
	workflows := []map[string]interface{}{
		{"id": "wf-001", "name": "Alpha", "enabled": true, "visibility": "private", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z"},
		{"id": "wf-002", "name": "Beta", "enabled": false, "visibility": "public", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z"},
	}
	server := makeWorkflowsServer(t, workflows)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls", "--jq", ".[0].name"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	out := strings.TrimSpace(outBuf.String())
	assert.Equal(t, `"Alpha"`, out, "expected jq filter to return first workflow name")
}

func TestListCmd_EmptyResponsePrintsEmptyTable(t *testing.T) {
	server := makeWorkflowsServer(t, []map[string]interface{}{})
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls"})
	err := wfCmd.Execute()
	assert.NoError(t, err, "empty list should not return error")
}

func TestListCmd_DisabledWorkflowShowsPaused(t *testing.T) {
	workflows := []map[string]interface{}{
		{"id": "wf-002", "name": "Paused One", "enabled": false, "visibility": "private", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z"},
	}
	server := makeWorkflowsServer(t, workflows)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "paused", "expected 'paused' status for disabled workflow")
}
