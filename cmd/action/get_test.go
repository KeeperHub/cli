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

func newActionGetFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeIntegrationsListServer(t *testing.T, integrations []map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/integrations" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(integrations)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
}

func TestGetCmd(t *testing.T) {
	integrations := []map[string]interface{}{
		{
			"id":        "int-001",
			"name":      "ERC20 Transfer",
			"type":      "web3",
			"isManaged": true,
			"createdAt": "2026-01-01T00:00:00Z",
			"updatedAt": "2026-02-01T00:00:00Z",
		},
		{
			"id":        "int-002",
			"name":      "Discord Notify",
			"type":      "discord",
			"isManaged": false,
			"createdAt": "2026-01-02T00:00:00Z",
			"updatedAt": "2026-02-02T00:00:00Z",
		},
	}
	server := makeIntegrationsListServer(t, integrations)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newActionGetFactory(server, ios)

	actionCmd := action.NewActionCmd(f)
	actionCmd.SetArgs([]string{"get", "ERC20 Transfer"})
	err := actionCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "ERC20 Transfer", "expected action name in detail output")
	assert.Contains(t, out, "web3", "expected type in detail output")
	assert.Contains(t, out, "yes", "expected managed=yes in detail output")
}

func TestGetCmd_ByID(t *testing.T) {
	integrations := []map[string]interface{}{
		{
			"id":        "int-001",
			"name":      "ERC20 Transfer",
			"type":      "web3",
			"isManaged": true,
			"createdAt": "2026-01-01T00:00:00Z",
			"updatedAt": "2026-02-01T00:00:00Z",
		},
	}
	server := makeIntegrationsListServer(t, integrations)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newActionGetFactory(server, ios)

	actionCmd := action.NewActionCmd(f)
	actionCmd.SetArgs([]string{"get", "int-001"})
	err := actionCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "ERC20 Transfer", "expected action name in detail output when searched by ID")
}

func TestGetCmd_NotFound(t *testing.T) {
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
	server := makeIntegrationsListServer(t, integrations)
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newActionGetFactory(server, ios)

	actionCmd := action.NewActionCmd(f)
	actionCmd.SetArgs([]string{"get", "nonexistent-action"})
	err := actionCmd.Execute()
	require.Error(t, err)

	var notFound cmdutil.NotFoundError
	require.ErrorAs(t, err, &notFound, "expected NotFoundError for missing action")
}
