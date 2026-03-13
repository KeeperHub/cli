package project_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/keeperhub/cli/pkg/iostreams"
)

func makeProjectCreateServer(t *testing.T, created map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/projects" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	}))
}

func TestCreateCmd(t *testing.T) {
	created := map[string]interface{}{
		"id":            "proj-new",
		"name":          "My New Project",
		"description":   "",
		"workflowCount": 0,
		"createdAt":     "2026-01-01T00:00:00Z",
		"updatedAt":     "2026-01-01T00:00:00Z",
	}
	var gotName, gotDesc string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/projects" {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotName = body["name"]
			gotDesc = body["description"]
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newProjectFactory(server, ios)

	cmd := project.NewProjectCmd(f)
	cmd.SetArgs([]string{"create", "My New Project", "--description", "A description"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "My New Project", gotName)
	assert.Equal(t, "A description", gotDesc)
	assert.Contains(t, outBuf.String(), "proj-new")
}

func TestCreateCmd_NoArgs(t *testing.T) {
	server := makeProjectCreateServer(t, map[string]interface{}{})
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newProjectFactory(server, ios)

	cmd := project.NewProjectCmd(f)
	cmd.SetArgs([]string{"create"})
	err := cmd.Execute()
	assert.Error(t, err)
}
