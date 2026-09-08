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
			"executionsUsed": 450,
			"executionLimit": 1000,
		},
		"overageCharges": []map[string]interface{}{
			{
				"periodStart":       "2026-08-01T00:00:00.000Z",
				"periodEnd":         "2026-09-01T00:00:00.000Z",
				"overageCount":      120,
				"totalChargeCents":  350,
				"status":            "pending",
				"createdAt":         "2026-09-01T00:00:00.000Z",
				"providerInvoiceId": nil,
			},
		},
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

// realServerSubscriptionPayload mirrors GET /api/billing/subscription, where
// overageCharges is an array of recent billing line-items (not a scalar).
const realServerSubscriptionPayload = `{
  "subscription": {"plan": "Pro", "status": "active"},
  "usage": {"executionsUsed": 450, "executionLimit": 1000},
  "overageCharges": [
    {"periodStart": "2026-08-01T00:00:00.000Z", "periodEnd": "2026-09-01T00:00:00.000Z", "overageCount": 120, "totalChargeCents": 350, "status": "pending", "createdAt": "2026-09-01T00:00:00.000Z", "providerInvoiceId": null},
    {"periodStart": "2026-07-01T00:00:00.000Z", "periodEnd": "2026-08-01T00:00:00.000Z", "overageCount": 40, "totalChargeCents": 125, "status": "paid", "createdAt": "2026-08-01T00:00:00.000Z", "providerInvoiceId": "in_123"}
  ],
  "limits": {"maxWorkflows": 50}
}`

func TestSubscriptionResponse_DecodesOverageChargesArray(t *testing.T) {
	var sub billing.SubscriptionResponse
	err := json.Unmarshal([]byte(realServerSubscriptionPayload), &sub)
	require.NoError(t, err, "real server payload with overageCharges array must decode without error")

	require.Len(t, sub.OverageCharges, 2)
	assert.Equal(t, 350, sub.OverageCharges[0].TotalChargeCents)
	assert.Equal(t, "pending", sub.OverageCharges[0].Status)
	assert.Nil(t, sub.OverageCharges[0].ProviderInvoiceID)
	require.NotNil(t, sub.OverageCharges[1].ProviderInvoiceID)
	assert.Equal(t, "in_123", *sub.OverageCharges[1].ProviderInvoiceID)

	// 350 + 125 cents = $4.75
	assert.InDelta(t, 4.75, sub.TotalOverageDollars(), 1e-9)
}

func TestSubscriptionResponse_DecodesUsage(t *testing.T) {
	var sub billing.SubscriptionResponse
	err := json.Unmarshal([]byte(realServerSubscriptionPayload), &sub)
	require.NoError(t, err)

	// The server sends usage.executionsUsed / usage.executionLimit; a struct
	// tagged executions/limit silently decodes these to zero, so the command
	// reports "0 / 0" regardless of real usage.
	assert.Equal(t, 450, sub.Usage.Executions)
	assert.Equal(t, 1000, sub.Usage.Limit)
}

func TestStatusCmd_OverageChargesArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(realServerSubscriptionPayload))
	}))
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newBillingFactory(server, ios)

	err := runBillingViaParent(f, []string{"st"})
	require.NoError(t, err, "overageCharges array must not crash response decoding")
	assert.Contains(t, outBuf.String(), "Overage:     $4.75")
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
