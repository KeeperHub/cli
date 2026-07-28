package workflow_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/workflow"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newResumeFactory(ios *iostreams.IOStreams, svr *httptest.Server) *cmdutil.Factory {
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:  ios,
		Config: func() (config.Config, error) {
			return config.Config{DefaultHost: svr.URL}, nil
		},
		HTTPClient: func() (*khhttp.Client, error) {
			return khhttp.NewClient(khhttp.ClientOptions{
				AppVersion: "1.0.0",
				IOStreams:  ios,
			}), nil
		},
	}
}

func runResumeViaParent(f *cmdutil.Factory, args []string) error {
	parent := workflow.NewWorkflowCmd(f)
	parent.SetArgs(append([]string{"resume"}, args...))
	return parent.Execute()
}

func TestResumeSendsPATCHWithEnabledTrue(t *testing.T) {
	var receivedMethod, receivedPath string
	var receivedBody map[string]interface{}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"wf-abc","enabled":true}`))
	}))
	defer svr.Close()

	// Non-TTY: auto-skips confirmation
	ios, outBuf, _, _ := iostreams.Test()
	f := newResumeFactory(ios, svr)

	err := runResumeViaParent(f, []string{"wf-abc"})

	require.NoError(t, err)
	assert.Equal(t, "PATCH", receivedMethod)
	assert.Equal(t, "/api/workflows/wf-abc", receivedPath)
	// The inverse of disable: this is the assertion that distinguishes the two.
	assert.Equal(t, true, receivedBody["enabled"])
	assert.Contains(t, outBuf.String(), "enabled")
}

func TestResumeYesFlagSkipsConfirmation(t *testing.T) {
	var patchCalled bool
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" {
			patchCalled = true
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"wf-abc","enabled":true}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newResumeFactory(ios, svr)

	err := runResumeViaParent(f, []string{"wf-abc", "--yes"})

	require.NoError(t, err)
	assert.True(t, patchCalled, "PATCH should have been called")
}

func TestResumeActivateAlias(t *testing.T) {
	var patchCalled bool
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" {
			patchCalled = true
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"wf-abc","enabled":true}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newResumeFactory(ios, svr)

	parent := workflow.NewWorkflowCmd(f)
	parent.SetArgs([]string{"activate", "wf-abc", "--yes"})

	require.NoError(t, parent.Execute())
	assert.True(t, patchCalled, "PATCH should have been called via the activate alias")
}

func TestResumeAPIErrorSurfaces(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Workflow not found"}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newResumeFactory(ios, svr)

	err := runResumeViaParent(f, []string{"wf-missing", "--yes"})

	require.Error(t, err)
}
