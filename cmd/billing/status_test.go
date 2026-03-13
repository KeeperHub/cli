package billing_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/billing"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBillingFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func runBillingViaParent(f *cmdutil.Factory, args []string) error {
	parent := billing.NewBillingCmd(f)
	parent.SetArgs(args)
	return parent.Execute()
}

func makeSubscriptionResponse() map[string]interface{} {
	return map[string]interface{}{
		"subscription": map[string]interface{}{
			"plan":   "Pro",
			"status": "active",
		},
		"usage": map[string]interface{}{
			"executions": 450,
			"limit":      1000,
		},
		"overageCharges": 0.0,
		"limits": map[string]interface{}{
			"maxWorkflows": 50,
		},
	}
}

func TestStatusCmd(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/billing/subscription" {
			called = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(makeSubscriptionResponse())
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newBillingFactory(server, ios)

	err := runBillingViaParent(f, []string{"st"})
	require.NoError(t, err)
	assert.True(t, called, "expected GET /api/billing/subscription to be called")

	out := outBuf.String()
	assert.Contains(t, out, "Pro")
	assert.Contains(t, out, "active")
}

func TestStatusCmd_NotEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newBillingFactory(server, ios)

	err := runBillingViaParent(f, []string{"st"})
	require.NoError(t, err, "404 should not return an error")
	assert.Contains(t, outBuf.String(), "Billing is not enabled for this instance.")
}

func TestStatusCmd_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(makeSubscriptionResponse())
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newBillingFactory(server, ios)

	err := runBillingViaParent(f, []string{"st", "--json"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &result))
	assert.Contains(t, result, "subscription")
	assert.Contains(t, result, "usage")
}
