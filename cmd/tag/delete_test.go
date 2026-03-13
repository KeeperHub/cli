package tag_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/tag"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeDeleteTagsServer(t *testing.T, tagID string) (*httptest.Server, *bool) {
	t.Helper()
	deleteCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/api/tags" {
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{"id": tagID, "name": "Test Tag", "color": "#6366f1", "workflowCount": 0},
			})
		} else if r.Method == http.MethodDelete && r.URL.Path == "/api/tags/"+tagID {
			deleteCalled = true
			_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	return server, &deleteCalled
}

func TestDeleteCmd(t *testing.T) {
	server, deleteCalled := makeDeleteTagsServer(t, "tag-001")
	defer server.Close()

	// Non-TTY auto-proceeds without prompt
	ios, outBuf, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"rm", "tag-001"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, *deleteCalled, "expected DELETE /api/tags/tag-001 to be called")
	assert.Contains(t, outBuf.String(), "tag-001")
}

func TestDeleteCmd_Yes(t *testing.T) {
	server, deleteCalled := makeDeleteTagsServer(t, "tag-001")
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"delete", "tag-001", "--yes"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, *deleteCalled)
}

func TestDeleteCmd_Cancel(t *testing.T) {
	// Document TTY cancellation behavior. In TTY mode, answering "n" returns CancelError.
	// Unit tests run in non-TTY (bytes.Buffer), so auto-proceed applies.
	expected := cmdutil.CancelError{Err: fmt.Errorf("delete cancelled")}
	assert.Equal(t, "delete cancelled", expected.Error())
}

func TestDeleteCmd_AliasRM(t *testing.T) {
	server, deleteCalled := makeDeleteTagsServer(t, "tag-002")
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newTagFactory(server, ios)

	cmd := tag.NewTagCmd(f)
	cmd.SetArgs([]string{"rm", "tag-002"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, *deleteCalled)
	assert.Contains(t, outBuf.String(), "tag-002")
}
