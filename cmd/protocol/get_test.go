package protocol_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/protocol"
	"github.com/keeperhub/cli/internal/cache"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProtoGetFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func TestGetCmd(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	// Pre-write cache
	data, _ := json.Marshal(sampleSchemasResponse())
	require.NoError(t, cache.WriteCache(cache.ProtocolCacheName, data))

	server := makeSchemasServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Should not need network if cache is fresh
		http.Error(w, "unexpected", http.StatusInternalServerError)
	})
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newProtoGetFactory(server, ios)

	cmd := protocol.NewProtocolCmd(f)
	cmd.SetArgs([]string{"get", "aave"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "aave", "expected protocol name in output")
	assert.Contains(t, out, "aave/supply", "expected action type in output")
	assert.Contains(t, out, "Supply", "expected action label in output")
}

func TestGetCmd_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	// Pre-write cache
	data, _ := json.Marshal(sampleSchemasResponse())
	require.NoError(t, cache.WriteCache(cache.ProtocolCacheName, data))

	server := makeSchemasServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unexpected", http.StatusInternalServerError)
	})
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newProtoGetFactory(server, ios)

	cmd := protocol.NewProtocolCmd(f)
	cmd.SetArgs([]string{"get", "nonexistent-slug"})
	err := cmd.Execute()
	assert.Error(t, err)

	var notFound cmdutil.NotFoundError
	require.ErrorAs(t, err, &notFound)
}
