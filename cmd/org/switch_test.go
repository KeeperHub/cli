package org_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/org"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSwitchFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeOrgSwitchServer(t *testing.T, orgs []map[string]interface{}) *httptest.Server {
	t.Helper()
	setActiveCalled := false
	_ = setActiveCalled
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/organizations":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(orgs)
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth/organization/set-active":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"session": "updated"})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
}

func TestSwitchCmd_ResolvesSlugAndCallsSetActive(t *testing.T) {
	setActiveCalled := false
	var setActiveBody map[string]string

	orgs := []map[string]interface{}{
		{
			"id":       "org-001",
			"name":     "Acme Corp",
			"slug":     "acme",
			"role":     "owner",
			"metadata": map[string]interface{}{"plan": "pro", "memberCount": 3},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/organizations":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(orgs)
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth/organization/set-active":
			setActiveCalled = true
			_ = json.NewDecoder(r.Body).Decode(&setActiveBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"session": "updated"})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newSwitchFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"switch", "acme"})
	err := orgCmd.Execute()
	require.NoError(t, err)

	assert.True(t, setActiveCalled, "expected POST /api/auth/organization/set-active to be called")
	assert.Equal(t, "org-001", setActiveBody["organizationId"], "expected organizationId in request body")
}

func TestSwitchCmd_Confirmation(t *testing.T) {
	orgs := []map[string]interface{}{
		{
			"id":       "org-001",
			"name":     "Acme Corp",
			"slug":     "acme",
			"role":     "owner",
			"metadata": map[string]interface{}{"plan": "pro", "memberCount": float64(3)},
		},
	}

	server := makeOrgSwitchServer(t, orgs)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newSwitchFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"switch", "acme"})
	err := orgCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "Switched to organization", "expected confirmation message")
	assert.Contains(t, out, "Acme Corp", "expected org name in confirmation")
}

func TestSwitchCmd_NotFound(t *testing.T) {
	orgs := []map[string]interface{}{
		{"id": "org-001", "name": "Acme Corp", "slug": "acme", "role": "owner"},
	}

	server := makeOrgSwitchServer(t, orgs)
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newSwitchFactory(server, ios)

	orgCmd := org.NewOrgCmd(f)
	orgCmd.SetArgs([]string{"switch", "nonexistent"})
	err := orgCmd.Execute()
	require.Error(t, err)

	var notFound cmdutil.NotFoundError
	require.ErrorAs(t, err, &notFound, "expected NotFoundError for unknown slug")
}
