package workflow_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func makeWF(id, projectID string) map[string]interface{} {
	m := map[string]interface{}{
		"id": id, "name": id, "enabled": true, "visibility": "private",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	if projectID != "" {
		m["projectId"] = projectID
	}
	return m
}

func makeWFs(n int) []map[string]interface{} {
	out := make([]map[string]interface{}, n)
	for i := range out {
		out[i] = makeWF("wf-"+strconv.Itoa(i), "")
	}
	return out
}

// makePaginatedWorkflowsServer serves GET /api/workflows against the given
// backing dataset the way app/api/workflows/route.ts (in the keeperhub repo)
// actually behaves: it slices by &limit=&offset=, and optionally filters by
// &projectId=/&tagId= first. requests accumulates each request's raw query
// string, for assertions on how many pages a test caused.
func makePaginatedWorkflowsServer(t *testing.T, all []map[string]interface{}) (server *httptest.Server, requests *[]string) {
	t.Helper()
	reqs := []string{}
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/workflows" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		reqs = append(reqs, r.URL.RawQuery)
		q := r.URL.Query()

		filtered := all
		if pid := q.Get("projectId"); pid != "" {
			var f []map[string]interface{}
			for _, wf := range filtered {
				if wf["projectId"] == pid {
					f = append(f, wf)
				}
			}
			filtered = f
		}
		if tid := q.Get("tagId"); tid != "" {
			var f []map[string]interface{}
			for _, wf := range filtered {
				if wf["tagId"] == tid {
					f = append(f, wf)
				}
			}
			filtered = f
		}

		limit, _ := strconv.Atoi(q.Get("limit"))
		offset, _ := strconv.Atoi(q.Get("offset"))
		start := offset
		if start > len(filtered) {
			start = len(filtered)
		}
		end := start + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		page := filtered[start:end]
		if page == nil {
			page = []map[string]interface{}{}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page)
	}))
	return server, &reqs
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

func TestListCmd_PaginatesPastAPICapToSatisfyLimit(t *testing.T) {
	all := makeWFs(250)
	server, requests := makePaginatedWorkflowsServer(t, all)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls", "--limit", "220", "--json"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &got))
	assert.Len(t, got, 220, "expected pagination to satisfy a --limit above the API's 200-per-request cap")
	assert.GreaterOrEqual(t, len(*requests), 2, "expected more than one request to page past the 200 cap")
}

func TestListCmd_NotesWhenMoreExistBeyondLimit(t *testing.T) {
	all := makeWFs(10)
	server, _ := makePaginatedWorkflowsServer(t, all)
	defer server.Close()

	ios, _, errBuf, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls", "--limit", "3"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, errBuf.String(), "note: more workflows exist", "expected a note when more workflows exist beyond --limit")
}

func TestListCmd_NoNoteWhenLimitExactlyCoversAll(t *testing.T) {
	all := makeWFs(3)
	server, _ := makePaginatedWorkflowsServer(t, all)
	defer server.Close()

	ios, _, errBuf, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls", "--limit", "3"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	assert.Empty(t, errBuf.String(), "expected no note when --limit exactly covers every workflow")
}

func TestListCmd_AllPaginatesUntilExhausted(t *testing.T) {
	all := makeWFs(250)
	server, requests := makePaginatedWorkflowsServer(t, all)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls", "--all", "--json"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &got))
	assert.Len(t, got, 250, "expected --all to page through every workflow past the 200-per-request cap")
	assert.GreaterOrEqual(t, len(*requests), 2, "expected more than one request to page past the 200 cap")
}

func TestListCmd_AllIgnoresDefaultLimit(t *testing.T) {
	all := makeWFs(35)
	server, _ := makePaginatedWorkflowsServer(t, all)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	// No --limit passed: the default of 30 must not cap --all's result.
	wfCmd.SetArgs([]string{"ls", "--all", "--json"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &got))
	assert.Len(t, got, 35, "expected --all to return every workflow, not the default --limit of 30")
}

func TestListCmd_AllRespectsExplicitLimit(t *testing.T) {
	all := makeWFs(35)
	server, _ := makePaginatedWorkflowsServer(t, all)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls", "--all", "--limit", "2", "--json"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &got))
	assert.Len(t, got, 2, "expected an explicit --limit to still cap --all's result")
}

func TestListCmd_AllCanBeScopedToProject(t *testing.T) {
	var all []map[string]interface{}
	for i := 0; i < 5; i++ {
		all = append(all, makeWF("proj1-wf-"+strconv.Itoa(i), "proj-1"))
	}
	for i := 0; i < 2; i++ {
		all = append(all, makeWF("proj2-wf-"+strconv.Itoa(i), "proj-2"))
	}
	server, _ := makePaginatedWorkflowsServer(t, all)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWFListFactory(server, ios)

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls", "--all", "--project", "proj-1", "--json"})
	err := wfCmd.Execute()
	require.NoError(t, err)

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &got))
	assert.Len(t, got, 5, "expected --all combined with --project to page through only that project's workflows")
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
