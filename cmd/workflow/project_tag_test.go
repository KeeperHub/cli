package workflow_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/keeperhub/cli/cmd/workflow"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runWorkflowSubcmd(f *cmdutil.Factory, args []string) error {
	parent := workflow.NewWorkflowCmd(f)
	parent.SetArgs(args)
	return parent.Execute()
}

func TestCreateSendsProjectAndTag(t *testing.T) {
	var body map[string]interface{}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"wf-1","name":"P","createdAt":"2026-01-01"}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runWorkflowSubcmd(f, []string{"create", "--name", "P", "--project", "proj_123", "--tag", "tag_456"})

	require.NoError(t, err)
	assert.Equal(t, "proj_123", body["projectId"])
	assert.Equal(t, "tag_456", body["tagId"])
}

func TestCreateOmitsProjectAndTagWhenUnset(t *testing.T) {
	var body map[string]interface{}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"wf-1","name":"P","createdAt":"2026-01-01"}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runWorkflowSubcmd(f, []string{"create", "--name", "P"})

	require.NoError(t, err)
	_, hasProject := body["projectId"]
	_, hasTag := body["tagId"]
	assert.False(t, hasProject, "projectId must be omitted when --project is unset")
	assert.False(t, hasTag, "tagId must be omitted when --tag is unset")
}

func TestUpdateAssignsProjectAndTag(t *testing.T) {
	var body map[string]interface{}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"wf-1"}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runWorkflowSubcmd(f, []string{"update", "wf-1", "--project", "proj_123", "--tag", "tag_456"})

	require.NoError(t, err)
	assert.Equal(t, "proj_123", body["projectId"])
	assert.Equal(t, "tag_456", body["tagId"])
}

func TestUpdateUnassignsWithEmptyValue(t *testing.T) {
	var body map[string]interface{}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"wf-1"}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runWorkflowSubcmd(f, []string{"update", "wf-1", "--project", ""})

	require.NoError(t, err)
	val, has := body["projectId"]
	assert.True(t, has, "projectId key must be present to send an explicit null")
	assert.Nil(t, val, "an empty --project must serialize to JSON null (unassign)")
}

func TestUpdateOmitsProjectAndTagWhenUnset(t *testing.T) {
	var body map[string]interface{}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"wf-1"}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runWorkflowSubcmd(f, []string{"update", "wf-1", "--name", "Renamed"})

	require.NoError(t, err)
	_, hasProject := body["projectId"]
	_, hasTag := body["tagId"]
	assert.False(t, hasProject, "an unrelated update must not touch projectId")
	assert.False(t, hasTag, "an unrelated update must not touch tagId")
}

func TestListFiltersByProjectAndTag(t *testing.T) {
	var query url.Values
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runWorkflowSubcmd(f, []string{"list", "--project", "proj_123", "--tag", "tag_456"})

	require.NoError(t, err)
	assert.Equal(t, "proj_123", query.Get("projectId"))
	assert.Equal(t, "tag_456", query.Get("tagId"))
}
