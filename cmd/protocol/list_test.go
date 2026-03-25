package protocol_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keeperhub/cli/cmd/protocol"
	"github.com/keeperhub/cli/internal/cache"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProtoFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeSchemasServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func sampleSchemasResponse() map[string]interface{} {
	return map[string]interface{}{
		"actions": map[string]interface{}{
			"aave/supply": map[string]interface{}{
				"actionType":  "aave/supply",
				"label":       "Aave: Supply",
				"description": "Supply assets",
				"category":    "Aave",
				"integration": "aave",
			},
			"aave/borrow": map[string]interface{}{
				"actionType":  "aave/borrow",
				"label":       "Aave: Borrow",
				"description": "Borrow assets",
				"category":    "Aave",
				"integration": "aave",
			},
			"uniswap/swap": map[string]interface{}{
				"actionType":  "uniswap/swap",
				"label":       "Uniswap: Swap",
				"description": "Swap tokens",
				"category":    "Uniswap",
				"integration": "uniswap",
			},
		},
	}
}

func TestListCmd_CacheMiss(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	called := false
	server := makeSchemasServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/mcp/schemas" {
			called = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sampleSchemasResponse())
			return
		}
		http.Error(w, "unexpected", http.StatusBadRequest)
	})
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newProtoFactory(server, ios)

	cmd := protocol.NewProtocolCmd(f)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, called, "expected GET /api/mcp/schemas to be called on cache miss")
	out := outBuf.String()
	assert.Contains(t, out, "aave")
	assert.Contains(t, out, "uniswap")
}

func TestListCmd_CacheHit(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	// Pre-write cache
	data, _ := json.Marshal(sampleSchemasResponse())
	require.NoError(t, cache.WriteCache(cache.ProtocolCacheName, data))

	networkCalled := false
	server := makeSchemasServer(t, func(w http.ResponseWriter, r *http.Request) {
		networkCalled = true
		http.Error(w, "should not be called", http.StatusInternalServerError)
	})
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newProtoFactory(server, ios)

	cmd := protocol.NewProtocolCmd(f)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.False(t, networkCalled, "expected no network request on cache hit")
	assert.Contains(t, outBuf.String(), "aave")
}

func TestListCmd_Refresh(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	// Pre-write cache
	data, _ := json.Marshal(sampleSchemasResponse())
	require.NoError(t, cache.WriteCache(cache.ProtocolCacheName, data))

	networkCalled := false
	server := makeSchemasServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/mcp/schemas" {
			networkCalled = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sampleSchemasResponse())
			return
		}
		http.Error(w, "unexpected", http.StatusBadRequest)
	})
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newProtoFactory(server, ios)

	cmd := protocol.NewProtocolCmd(f)
	cmd.SetArgs([]string{"list", "--refresh"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, networkCalled, "expected network request with --refresh even when cache exists")
}

func TestListCmd_StaleWithError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	// Write stale cache (2 hours ago)
	staleEntry := cache.CacheEntry{
		FetchedAt: time.Now().UTC().Add(-2 * time.Hour),
		Data:      mustMarshal(sampleSchemasResponse()),
	}
	staleBytes, _ := json.Marshal(staleEntry)
	require.NoError(t, cache.WriteRaw(cache.ProtocolCacheName, staleBytes))

	// Server returns 500
	server := makeSchemasServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	})
	defer server.Close()

	ios, outBuf, errBuf, _ := iostreams.Test()
	f := newProtoFactory(server, ios)

	cmd := protocol.NewProtocolCmd(f)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	require.NoError(t, err, "stale cache with error should not return error")

	assert.Contains(t, outBuf.String(), "aave", "expected stale data to be served")
	assert.Contains(t, errBuf.String(), "Warning", "expected warning on stderr")
}

func TestListCmd_NoCacheWithError(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	server := makeSchemasServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	})
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newProtoFactory(server, ios)

	cmd := protocol.NewProtocolCmd(f)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	assert.Error(t, err, "expected error when no cache and API fails")
	assert.Contains(t, err.Error(), "could not fetch protocols")
}

func TestListCmd_Table(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	server := makeSchemasServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/mcp/schemas" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sampleSchemasResponse())
			return
		}
		http.Error(w, "unexpected", http.StatusBadRequest)
	})
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newProtoFactory(server, ios)

	cmd := protocol.NewProtocolCmd(f)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	// Test streams are non-TTY, so tsvWriter outputs data rows without headers.
	// Verify protocol names and action counts are present.
	assert.Contains(t, out, "aave")
	assert.Contains(t, out, "2")
	assert.Contains(t, out, "uniswap")
	assert.Contains(t, out, "1")
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
