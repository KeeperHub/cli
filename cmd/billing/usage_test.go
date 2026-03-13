package billing_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func TestUsageCmd(t *testing.T) {
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

	err := runBillingViaParent(f, []string{"usage"})
	require.NoError(t, err)
	assert.True(t, called, "expected GET /api/billing/subscription to be called")

	out := outBuf.String()
	assert.Contains(t, out, "450")
	assert.Contains(t, out, "1000")
}

func TestUsageCmd_Period(t *testing.T) {
	var gotPeriod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPeriod = r.URL.Query().Get("period")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(makeSubscriptionResponse())
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newBillingFactory(server, ios)

	err := runBillingViaParent(f, []string{"usage", "--period", "2026-02"})
	require.NoError(t, err)
	assert.Equal(t, "2026-02", gotPeriod, "expected period query param to be sent")
}

func TestUsageCmd_NotEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newBillingFactory(server, ios)

	err := runBillingViaParent(f, []string{"usage"})
	require.NoError(t, err, "404 should not return an error")
	assert.Contains(t, outBuf.String(), "Billing is not enabled for this instance.")
}

func TestUsageCmd_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(makeSubscriptionResponse())
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newBillingFactory(server, ios)

	err := runBillingViaParent(f, []string{"usage", "--json"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &result))
	assert.Contains(t, result, "subscription")
	assert.Contains(t, result, "usage")
}
