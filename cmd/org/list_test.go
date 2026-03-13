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

func newOrgFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeOrgsServer(t *testing.T, orgs []map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/organizations" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(orgs)
	}))
}

func TestListCmd_SendsGETOrganizations(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/organizations" {
			called = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]interface{}{})
		} else {
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newOrgFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"ls"})
	err := orgCmd.Execute()
	require.NoError(t, err)
	assert.True(t, called, "expected GET /api/organizations to be called")
}

func TestListCmd_RendersTableWithColumns(t *testing.T) {
	orgs := []map[string]interface{}{
		{
			"id":        "org-001",
			"name":      "Acme Corp",
			"slug":      "acme",
			"role":      "owner",
			"createdAt": "2026-01-01T00:00:00Z",
		},
	}
	server := makeOrgsServer(t, orgs)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newOrgFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"ls"})
	err := orgCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "Acme Corp", "expected org name in output")
	assert.Contains(t, out, "acme", "expected org slug in output")
	assert.Contains(t, out, "owner", "expected role in output")
}

func TestListCmd_Empty(t *testing.T) {
	server := makeOrgsServer(t, []map[string]interface{}{})
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newOrgFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"ls"})
	err := orgCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "No organizations found", "expected empty hint message")
}

func TestListCmd_JSON(t *testing.T) {
	orgs := []map[string]interface{}{
		{"id": "org-001", "name": "Acme Corp", "slug": "acme", "role": "owner", "createdAt": "2026-01-01T00:00:00Z"},
	}
	server := makeOrgsServer(t, orgs)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newOrgFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"ls", "--json"})
	err := orgCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	var result []interface{}
	err = json.Unmarshal([]byte(strings.TrimSpace(out)), &result)
	require.NoError(t, err, "expected valid JSON array output")
	assert.Len(t, result, 1)
}
