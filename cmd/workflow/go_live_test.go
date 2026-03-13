package workflow_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/workflow"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGoLiveFactory(ios *iostreams.IOStreams, svr *httptest.Server) *cmdutil.Factory {
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		Config: func() (config.Config, error) {
			return config.Config{DefaultHost: svr.URL}, nil
		},
		HTTPClient: func() (*khhttp.Client, error) {
			return khhttp.NewClient(khhttp.ClientOptions{
				AppVersion: "1.0.0",
				IOStreams:   ios,
			}), nil
		},
	}
}

func runGoLiveViaParent(f *cmdutil.Factory, args []string) error {
	parent := workflow.NewWorkflowCmd(f)
	parent.SetArgs(append([]string{"go-live"}, args...))
	return parent.Execute()
}

func TestGoLiveSendsPUT(t *testing.T) {
	var receivedMethod, receivedPath string
	var receivedBody map[string]interface{}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"wf-abc","name":"My WF","enabled":true}`))
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runGoLiveViaParent(f, []string{"wf-abc", "--name", "My WF"})

	require.NoError(t, err)
	assert.Equal(t, "PUT", receivedMethod)
	assert.Equal(t, "/api/workflows/wf-abc/go-live", receivedPath)
	assert.Equal(t, "My WF", receivedBody["name"])
	assert.NotNil(t, receivedBody["publicTagIds"])
	assert.Contains(t, outBuf.String(), "wf-abc")
}

func TestGoLiveRequiresName(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer svr.Close()

	// In non-TTY mode, --name missing should return an error
	ios, _, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runGoLiveViaParent(f, []string{"wf-abc"})

	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "name")
}

func TestGoLiveJSON(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"wf-abc","name":"My WF","enabled":true}`))
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runGoLiveViaParent(f, []string{"wf-abc", "--name", "My WF", "--json"})

	require.NoError(t, err)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &result))
	assert.Equal(t, "wf-abc", result["id"])
}

func TestGoLive401GivesAuthHint(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newGoLiveFactory(ios, svr)

	err := runGoLiveViaParent(f, []string{"wf-abc", "--name", "My WF"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "kh auth login")
}
