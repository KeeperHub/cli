package template_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	templatecmd "github.com/keeperhub/cli/cmd/template"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTemplateFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func runTemplateViaParent(f *cmdutil.Factory, args []string) error {
	parent := templatecmd.NewTemplateCmd(f)
	parent.SetArgs(args)
	return parent.Execute()
}

func TestListCmd(t *testing.T) {
	templates := []map[string]interface{}{
		{
			"id":          "tpl-001",
			"name":        "DeFi Monitor",
			"description": "Monitor DeFi positions",
			"visibility":  "public",
			"publicTags":  []map[string]interface{}{{"name": "defi"}},
			"createdAt":   "2026-01-01T00:00:00Z",
		},
	}

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/workflows/public" && r.URL.Query().Get("featured") == "true" {
			called = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(templates)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTemplateFactory(server, ios)

	err := runTemplateViaParent(f, []string{"ls"})
	require.NoError(t, err)
	assert.True(t, called, "expected GET /api/workflows/public?featured=true to be called")

	out := outBuf.String()
	assert.Contains(t, out, "DeFi Monitor")
	assert.Contains(t, out, "Monitor DeFi positions")
}

func TestListCmd_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]interface{}{})
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTemplateFactory(server, ios)

	err := runTemplateViaParent(f, []string{"ls"})
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "No templates found.")
}

func TestListCmd_DescriptionTruncated(t *testing.T) {
	longDesc := strings.Repeat("x", 80)
	templates := []map[string]interface{}{
		{
			"id":          "tpl-002",
			"name":        "Long Desc",
			"description": longDesc,
			"visibility":  "public",
			"publicTags":  []map[string]interface{}{},
			"createdAt":   "2026-01-01T00:00:00Z",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(templates)
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTemplateFactory(server, ios)

	err := runTemplateViaParent(f, []string{"ls"})
	require.NoError(t, err)

	out := outBuf.String()
	assert.NotContains(t, out, longDesc, "full long description should not appear")
	assert.Contains(t, out, "...", "truncated description should end with ...")
}

func TestListCmd_CategoryFromTags(t *testing.T) {
	templates := []map[string]interface{}{
		{
			"id":          "tpl-003",
			"name":        "Tagged Template",
			"description": "Has a tag",
			"visibility":  "public",
			"publicTags":  []map[string]interface{}{{"name": "web3"}},
			"createdAt":   "2026-01-01T00:00:00Z",
		},
		{
			"id":          "tpl-004",
			"name":        "Untagged Template",
			"description": "No tags",
			"visibility":  "public",
			"publicTags":  []map[string]interface{}{},
			"createdAt":   "2026-01-01T00:00:00Z",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(templates)
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTemplateFactory(server, ios)

	err := runTemplateViaParent(f, []string{"ls"})
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "web3", "expected first tag name as category")
	assert.Contains(t, out, "General", "expected 'General' fallback for untagged template")
}
