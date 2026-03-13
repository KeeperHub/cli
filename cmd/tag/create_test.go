package tag_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/tag"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCmd(t *testing.T) {
	created := map[string]interface{}{
		"id":            "tag-new",
		"name":          "DeFi",
		"color":         "#6366f1",
		"workflowCount": 0,
		"createdAt":     "2026-01-01T00:00:00Z",
		"updatedAt":     "2026-01-01T00:00:00Z",
	}
	var gotName, gotColor string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/tags" {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotName = body["name"]
			gotColor = body["color"]
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"create", "DeFi"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "DeFi", gotName)
	assert.Equal(t, "#6366f1", gotColor, "expected default color #6366f1 when --color omitted")
	assert.Contains(t, outBuf.String(), "tag-new")
}

func TestCreateCmd_CustomColor(t *testing.T) {
	var gotColor string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/tags" {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotColor = body["color"]
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "tag-001", "name": "Web3", "color": gotColor, "workflowCount": 0,
			})
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"create", "Web3", "--color", "#ff0000"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "#ff0000", gotColor)
}

func TestCreateCmd_NoArgs(t *testing.T) {
	server := makeTagsServer(t, []map[string]interface{}{})
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"create"})
	err := cmd.Execute()
	assert.Error(t, err)
}
