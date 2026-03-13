package wallet_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keeperhub/cli/cmd/wallet"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWalletTokensFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeTokensServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestTokensCmd(t *testing.T) {
	called := false
	server := makeTokensServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/user/wallet/tokens" {
			called = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"tokens": []map[string]interface{}{
					{
						"chainId":      "1",
						"tokenAddress": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
						"symbol":       "USDC",
						"name":         "USD Coin",
						"decimals":     6,
					},
				},
			})
		} else {
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	})
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWalletTokensFactory(server, ios)

	walletCmd := wallet.NewWalletCmd(f)
	walletCmd.SetArgs([]string{"tokens"})
	err := walletCmd.Execute()
	require.NoError(t, err)
	assert.True(t, called, "expected GET /api/user/wallet/tokens to be called")

	out := outBuf.String()
	assert.Contains(t, out, "USDC", "expected token symbol in output")
	assert.Contains(t, out, "USD Coin", "expected token name in output")
}

func TestTokensCmd_Limit(t *testing.T) {
	var gotLimit string
	server := makeTokensServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"tokens": []interface{}{}})
	})
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newWalletTokensFactory(server, ios)

	walletCmd := wallet.NewWalletCmd(f)
	walletCmd.SetArgs([]string{"tokens", "--limit", "10"})
	err := walletCmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "10", gotLimit, "expected limit=10 query param")
}

func TestTokensCmd_Chain(t *testing.T) {
	var gotChain string
	server := makeTokensServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotChain = r.URL.Query().Get("chain")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"tokens": []interface{}{}})
	})
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newWalletTokensFactory(server, ios)

	walletCmd := wallet.NewWalletCmd(f)
	walletCmd.SetArgs([]string{"tokens", "--chain", "ethereum"})
	err := walletCmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "ethereum", gotChain, "expected chain=ethereum query param")
}
