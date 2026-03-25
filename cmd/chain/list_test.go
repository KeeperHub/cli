package chain_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/chain"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newChainFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func sampleChainsResponse() []map[string]interface{} {
	return []map[string]interface{}{
		{"chainId": 1, "name": "Ethereum", "type": "mainnet", "status": "active", "primaryRpcUrl": "https://eth.example.com", "fallbackRpcUrl": ""},
		{"chainId": 137, "name": "Polygon", "type": "mainnet", "status": "active", "primaryRpcUrl": "https://polygon.example.com", "fallbackRpcUrl": ""},
		{"chainId": 42161, "name": "Arbitrum", "type": "mainnet", "status": "active", "primaryRpcUrl": "", "fallbackRpcUrl": ""},
	}
}

func TestChainListCmd(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/chains" {
			called = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sampleChainsResponse())
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newChainFactory(server, ios)

	cmd := chain.NewChainCmd(f)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, called, "expected GET /api/chains to be called")
	out := outBuf.String()
	assert.Contains(t, out, "Ethereum")
	assert.Contains(t, out, "Polygon")
	assert.Contains(t, out, "Arbitrum")
}

func TestChainListCmd_Empty(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]interface{}{})
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newChainFactory(server, ios)

	cmd := chain.NewChainCmd(f)
	cmd.SetArgs([]string{"ls"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "No chains found.")
}

func TestChainListCmd_ServerError(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newChainFactory(server, ios)

	cmd := chain.NewChainCmd(f)
	cmd.SetArgs([]string{"ls"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestChainCmd_HasAlias(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{AppVersion: "1.0.0", IOStreams: ios}
	cmd := chain.NewChainCmd(f)
	assert.Contains(t, cmd.Aliases, "ch")
}
