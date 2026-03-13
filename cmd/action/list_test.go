package action_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/action"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newActionFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeIntegrationsServer(t *testing.T, integrations []map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/integrations" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(integrations)
	}))
}

func TestListCmd(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/integrations" {
			called = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":        "int-001",
					"name":      "ERC20 Transfer",
					"type":      "web3",
					"isManaged": true,
					"createdAt": "2026-01-01T00:00:00Z",
					"updatedAt": "2026-02-01T00:00:00Z",
				},
			})
		} else {
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newActionFactory(server, ios)

	actionCmd := action.NewActionCmd(f)
	actionCmd.SetArgs([]string{"ls"})
	err := actionCmd.Execute()
	require.NoError(t, err)
	assert.True(t, called, "expected GET /api/integrations to be called")

	out := outBuf.String()
	assert.Contains(t, out, "ERC20 Transfer", "expected action name in output")
	assert.Contains(t, out, "web3", "expected type in output")
	assert.Contains(t, out, "yes", "expected managed=yes in output")
}

func TestListCmd_Category(t *testing.T) {
	var gotCategory string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCategory = r.URL.Query().Get("category")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]interface{}{})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newActionFactory(server, ios)

	actionCmd := action.NewActionCmd(f)
	actionCmd.SetArgs([]string{"ls", "--category", "web3"})
	err := actionCmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "web3", gotCategory, "expected category=web3 query param")
}

func TestListCmd_Empty(t *testing.T) {
	server := makeIntegrationsServer(t, []map[string]interface{}{})
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newActionFactory(server, ios)

	actionCmd := action.NewActionCmd(f)
	actionCmd.SetArgs([]string{"ls"})
	err := actionCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "No actions found.", "expected empty hint message")
}

func TestListCmd_JSON(t *testing.T) {
	integrations := []map[string]interface{}{
		{
			"id":        "int-001",
			"name":      "ERC20 Transfer",
			"type":      "web3",
			"isManaged": true,
			"createdAt": "2026-01-01T00:00:00Z",
			"updatedAt": "2026-01-01T00:00:00Z",
		},
	}
	server := makeIntegrationsServer(t, integrations)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newActionFactory(server, ios)

	actionCmd := action.NewActionCmd(f)
	actionCmd.SetArgs([]string{"ls", "--json"})
	err := actionCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, `"id"`, "expected JSON with id field")
	assert.Contains(t, out, "int-001")
}
