package org_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/org"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMembersFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeMembersServer(t *testing.T, members []map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/auth/organization/list-members" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{"members": members}
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func TestMembersCmd_UsesGETMethod(t *testing.T) {
	method := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/organization/list-members" {
			method = r.Method
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"members": []interface{}{}})
		} else {
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newMembersFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"members"})
	err := orgCmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, method, "expected GET method for list-members endpoint")
}

func TestMembersCmd_RendersTableWithColumns(t *testing.T) {
	members := []map[string]interface{}{
		{
			"id":        "mem-001",
			"email":     "alice@acme.com",
			"role":      "owner",
			"createdAt": "2026-01-01T00:00:00Z",
			"user": map[string]interface{}{
				"name": "Alice Smith",
			},
		},
	}
	server := makeMembersServer(t, members)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newMembersFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"members"})
	err := orgCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "alice@acme.com", "expected email in output")
	assert.Contains(t, out, "owner", "expected role in output")
}

func TestMembersCmd_Empty(t *testing.T) {
	server := makeMembersServer(t, []map[string]interface{}{})
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newMembersFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"members"})
	err := orgCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "No members found", "expected empty hint message")
}

func TestMembersCmd_JSON(t *testing.T) {
	members := []map[string]interface{}{
		{"id": "mem-001", "email": "alice@acme.com", "role": "owner"},
	}
	server := makeMembersServer(t, members)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newMembersFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"members", "--json"})
	err := orgCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	var result interface{}
	err = json.Unmarshal([]byte(strings.TrimSpace(out)), &result)
	require.NoError(t, err, "expected valid JSON output")
	assert.Contains(t, out, "alice@acme.com", "expected member email in JSON output")
}
