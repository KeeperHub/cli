package tag_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/tag"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTagFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeTagsServer(t *testing.T, tags []map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/tags" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tags)
	}))
}

func TestListCmd(t *testing.T) {
	tags := []map[string]interface{}{
		{
			"id":            "tag-001",
			"name":          "DeFi",
			"color":         "#6366f1",
			"workflowCount": 5,
			"createdAt":     "2026-01-01T00:00:00Z",
			"updatedAt":     "2026-02-01T00:00:00Z",
		},
	}
	server := makeTagsServer(t, tags)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"ls"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "tag-001")
	assert.Contains(t, out, "DeFi")
	assert.Contains(t, out, "#6366f1")
	assert.Contains(t, out, "5")
}

func TestListCmd_Empty(t *testing.T) {
	server := makeTagsServer(t, []map[string]interface{}{})
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"ls"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "No tags found")
	assert.Contains(t, out, "kh t create 'name'")
}

func TestListCmd_JSON(t *testing.T) {
	tags := []map[string]interface{}{
		{"id": "tag-001", "name": "DeFi", "color": "#6366f1", "workflowCount": 1},
	}
	server := makeTagsServer(t, tags)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"ls", "--json"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, `"id"`)
	assert.Contains(t, out, "tag-001")
}

func TestListCmd_JQ(t *testing.T) {
	tags := []map[string]interface{}{
		{"id": "tag-001", "name": "DeFi", "color": "#6366f1", "workflowCount": 1},
	}
	server := makeTagsServer(t, tags)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"ls", "--jq", ".[0].name"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := strings.TrimSpace(outBuf.String())
	assert.Equal(t, `"DeFi"`, out)
}
