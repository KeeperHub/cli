package workflow_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/keeperhub/cli/cmd/workflow"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRunFactory(ios *iostreams.IOStreams, svr *httptest.Server) *cmdutil.Factory {
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

// runViaParent runs a workflow subcommand through the parent command so
// persistent flags (--json, --jq) are inherited.
func runViaParent(f *cmdutil.Factory, args []string) error {
	parent := workflow.NewWorkflowCmd(f)
	parent.SetArgs(append([]string{"run"}, args...))
	return parent.Execute()
}

func TestRunFireAndForget(t *testing.T) {
	var receivedMethod, receivedPath string
	var receivedBody []byte

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		receivedBody, _ = readBody(r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-123","status":"running"}`))
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newRunFactory(ios, svr)

	err := runViaParent(f, []string{"wf-abc"})

	require.NoError(t, err)
	assert.Equal(t, "POST", receivedMethod)
	assert.Equal(t, "/api/workflow/wf-abc/execute", receivedPath)
	assert.Equal(t, `{}`, strings.TrimSpace(string(receivedBody)))
	assert.Contains(t, outBuf.String(), "exec-123")
	assert.Contains(t, outBuf.String(), "Triggered run:")
}

func TestRunFireAndForgetJSON(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"executionId":"exec-456","status":"running"}`))
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newRunFactory(ios, svr)

	err := runViaParent(f, []string{"wf-abc", "--json"})

	require.NoError(t, err)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &result))
	assert.Equal(t, "exec-456", result["executionId"])
	assert.Equal(t, "running", result["status"])
}

func TestRunWaitSuccess(t *testing.T) {
	callCount := 0
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-789","status":"running"}`))
			return
		}
		// GET status
		callCount++
		w.WriteHeader(http.StatusOK)
		if callCount < 2 {
			_, _ = w.Write([]byte(`{"status":"running","progress":{"totalSteps":3,"completedSteps":1,"currentNodeName":"Step 1","percentage":33}}`))
		} else {
			_, _ = w.Write([]byte(`{"status":"success","progress":{"totalSteps":3,"completedSteps":3,"currentNodeName":"","percentage":100}}`))
		}
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newRunFactory(ios, svr)

	err := runViaParent(f, []string{"wf-abc", "--wait", "--timeout", "30s"})

	require.NoError(t, err)
	out := outBuf.String()
	assert.Contains(t, out, "success")
	assert.Contains(t, out, "3")
}

func TestRunWaitError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-err","status":"running"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"error","progress":{"totalSteps":2,"completedSteps":1,"currentNodeName":"","percentage":50}}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newRunFactory(ios, svr)

	err := runViaParent(f, []string{"wf-abc", "--wait"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestRunWaitTimeout(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-timeout","status":"running"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"running","progress":{"totalSteps":5,"completedSteps":1,"currentNodeName":"Step 1","percentage":20}}`))
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newRunFactory(ios, svr)

	start := time.Now()
	err := runViaParent(f, []string{"wf-abc", "--wait", "--timeout", "200ms"})
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
	assert.Less(t, elapsed, 5*time.Second, "should have timed out quickly")
}

func TestRunWaitNoTTYSuppressesProgress(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"executionId":"exec-notty","status":"running"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","progress":{"totalSteps":2,"completedSteps":2,"currentNodeName":"","percentage":100}}`))
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newRunFactory(ios, svr)

	err := runViaParent(f, []string{"wf-abc", "--wait"})

	require.NoError(t, err)
	out := outBuf.String()
	assert.NotContains(t, out, "\r", "should not contain carriage return in non-TTY mode")
	assert.Contains(t, out, "success")
}

func TestRunHasTimeoutFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{AppVersion: "1.0.0", IOStreams: ios}
	cmd := workflow.NewRunCmd(f)
	flag := cmd.Flags().Lookup("timeout")
	require.NotNil(t, flag, "run command must have --timeout flag")
	assert.Equal(t, "duration", flag.Value.Type())
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()
	buf := &bytes.Buffer{}
	_, err := buf.ReadFrom(r.Body)
	return buf.Bytes(), err
}
