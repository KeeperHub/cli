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

func newWalletFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
	}
}

func makeBalancesServer(t *testing.T, response map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/user/wallet/balances" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
}

func TestBalanceCmd(t *testing.T) {
	response := map[string]interface{}{
		"walletAddress": "0xABCD1234",
		"balances": []map[string]interface{}{
			{
				"chainId":       "1",
				"chainName":     "Ethereum",
				"nativeBalance": "1.5",
				"tokens": []map[string]interface{}{
					{"symbol": "USDC", "balance": "100.0", "decimals": 6},
				},
			},
		},
	}
	server := makeBalancesServer(t, response)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWalletFactory(server, ios)

	walletCmd := wallet.NewWalletCmd(f)
	walletCmd.SetArgs([]string{"bal"})
	err := walletCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "0xABCD1234", "expected wallet address in output")
	assert.Contains(t, out, "Ethereum", "expected chain name in output")
	assert.Contains(t, out, "1.5", "expected native balance in output")
	assert.Contains(t, out, "USDC", "expected token symbol in output")
}

func TestBalanceCmd_ZeroFiltered(t *testing.T) {
	response := map[string]interface{}{
		"walletAddress": "0xABCD1234",
		"balances": []map[string]interface{}{
			{
				"chainId":       "1",
				"chainName":     "Ethereum",
				"nativeBalance": "1.5",
				"tokens":        []map[string]interface{}{},
			},
			{
				"chainId":       "137",
				"chainName":     "Polygon",
				"nativeBalance": "0",
				"tokens":        []map[string]interface{}{},
			},
		},
	}
	server := makeBalancesServer(t, response)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWalletFactory(server, ios)

	walletCmd := wallet.NewWalletCmd(f)
	walletCmd.SetArgs([]string{"bal"})
	err := walletCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "Ethereum", "expected non-zero chain to be shown")
	assert.NotContains(t, out, "Polygon", "expected zero-balance chain to be hidden")
}

func TestBalanceCmd_ChainFilter(t *testing.T) {
	response := map[string]interface{}{
		"walletAddress": "0xABCD1234",
		"balances": []map[string]interface{}{
			{
				"chainId":       "1",
				"chainName":     "Ethereum",
				"nativeBalance": "1.5",
				"tokens":        []map[string]interface{}{},
			},
			{
				"chainId":       "8453",
				"chainName":     "Base",
				"nativeBalance": "0.5",
				"tokens":        []map[string]interface{}{},
			},
		},
	}
	server := makeBalancesServer(t, response)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWalletFactory(server, ios)

	walletCmd := wallet.NewWalletCmd(f)
	walletCmd.SetArgs([]string{"bal", "--chain", "ethereum"})
	err := walletCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "Ethereum", "expected filtered chain in output")
	assert.NotContains(t, out, "Base", "expected non-matching chain to be hidden")
}

func TestBalanceCmd_Empty(t *testing.T) {
	response := map[string]interface{}{
		"walletAddress": "0xABCD1234",
		"balances": []map[string]interface{}{
			{
				"chainId":       "1",
				"chainName":     "Ethereum",
				"nativeBalance": "0",
				"tokens":        []map[string]interface{}{},
			},
		},
	}
	server := makeBalancesServer(t, response)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWalletFactory(server, ios)

	walletCmd := wallet.NewWalletCmd(f)
	walletCmd.SetArgs([]string{"bal"})
	err := walletCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "No balances found.", "expected empty hint when all balances are zero")
}

func TestBalanceCmd_JSON(t *testing.T) {
	response := map[string]interface{}{
		"walletAddress": "0xABCD1234",
		"balances": []map[string]interface{}{
			{
				"chainId":       "1",
				"chainName":     "Ethereum",
				"nativeBalance": "0",
				"tokens":        []map[string]interface{}{},
			},
		},
	}
	server := makeBalancesServer(t, response)
	defer server.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newWalletFactory(server, ios)

	walletCmd := wallet.NewWalletCmd(f)
	walletCmd.SetArgs([]string{"bal", "--json"})
	err := walletCmd.Execute()
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, `"walletAddress"`, "expected JSON with walletAddress field")
	assert.Contains(t, out, "0xABCD1234", "expected wallet address in JSON output")
	assert.Contains(t, out, "Ethereum", "expected chain in JSON output (no zero-filtering in JSON mode)")
}
