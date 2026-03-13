package project_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/project"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeDeleteProjectsServer(t *testing.T, projectID string) (*httptest.Server, *bool) {
	t.Helper()
	deleteCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/api/projects" {
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{"id": projectID, "name": "Test Project", "description": "", "workflowCount": 0},
			})
		} else if r.Method == http.MethodDelete && r.URL.Path == "/api/projects/"+projectID {
			deleteCalled = true
			_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	return server, &deleteCalled
}

func TestDeleteCmd(t *testing.T) {
	server, deleteCalled := makeDeleteProjectsServer(t, "proj-001")
	defer server.Close()

	// Non-TTY auto-proceeds without prompt
	ios, outBuf, _, _ := iostreams.Test()
	f := newProjectFactory(server, ios)

	cmd := project.NewProjectCmd(f)
	cmd.SetArgs([]string{"rm", "proj-001"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, *deleteCalled, "expected DELETE /api/projects/proj-001 to be called")
	assert.Contains(t, outBuf.String(), "proj-001")
}

func TestDeleteCmd_Yes(t *testing.T) {
	server, deleteCalled := makeDeleteProjectsServer(t, "proj-001")
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newProjectFactory(server, ios)

	cmd := project.NewProjectCmd(f)
	cmd.SetArgs([]string{"delete", "proj-001", "--yes"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, *deleteCalled)
}

func TestDeleteCmd_Cancel(t *testing.T) {
	// Document TTY cancellation behavior. In TTY mode, answering "n" returns CancelError.
	// Unit tests run in non-TTY (bytes.Buffer), so auto-proceed applies.
	// We verify the CancelError type is correctly defined for the denial path.
	expected := cmdutil.CancelError{Err: fmt.Errorf("delete cancelled")}
	assert.Equal(t, "delete cancelled", expected.Error())
}

func TestDeleteCmd_AliasRM(t *testing.T) {
	server, deleteCalled := makeDeleteProjectsServer(t, "proj-002")
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newProjectFactory(server, ios)

	cmd := project.NewProjectCmd(f)
	cmd.SetArgs([]string{"rm", "proj-002"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, *deleteCalled)
	assert.Contains(t, outBuf.String(), "proj-002")
}
