package read_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/read"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReadFactory(rpcServer *httptest.Server, chainsServer *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: "test", AppVersion: "1.0.0"})
	baseURL := ""
	if chainsServer != nil {
		baseURL = chainsServer.URL
	}
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config: func() (config.Config, error) {
			return config.Config{
				RPC: map[string]string{
					"1": rpcServer.URL,
				},
			}, nil
		},
		BaseURL: func() string { return baseURL },
	}
}

func TestReadCmd_MissingChainFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer rpcServer.Close()

	f := newReadFactory(rpcServer, nil, ios)
	cmd := read.NewReadCmd(f)
	cmd.SetArgs([]string{"0xdAC17F958D2ee523a2206206994597C96e3cFa0e", "totalSupply()"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--chain is required")
}

func TestReadCmd_NoArgFunction(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	// Mock RPC server that returns a uint256 value
	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Verify method is eth_call
		assert.Equal(t, "eth_call", req["method"])

		params := req["params"].([]interface{})
		callObj := params[0].(map[string]interface{})

		// Verify the function selector for "totalSupply()"
		data := callObj["data"].(string)
		// keccak256("totalSupply()") first 4 bytes = 0x18160ddd
		assert.True(t, strings.HasPrefix(data, "0x18160ddd"), "expected totalSupply() selector, got %s", data)

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x000000000000000000000000000000000000000000000000000000174876e800",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer rpcServer.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newReadFactory(rpcServer, nil, ios)

	cmd := read.NewReadCmd(f)
	cmd.SetArgs([]string{"0xdAC17F958D2ee523a2206206994597C96e3cFa0e", "totalSupply()", "--chain", "1"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := strings.TrimSpace(outBuf.String())
	// 0x174876e800 = 100000000000 in decimal
	assert.Equal(t, "100000000000", out)
}

func TestReadCmd_AddressArgFunction(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)

		params := req["params"].([]interface{})
		callObj := params[0].(map[string]interface{})
		data := callObj["data"].(string)

		// keccak256("balanceOf(address)") first 4 bytes = 0x70a08231
		assert.True(t, strings.HasPrefix(data, "0x70a08231"), "expected balanceOf(address) selector, got %s", data)

		// Verify it has 4 bytes selector + 32 bytes address = 72 hex chars + "0x" prefix
		assert.Equal(t, 2+8+64, len(data), "expected 4-byte selector + 32-byte address arg")

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x0000000000000000000000000000000000000000000000000000000000000064",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer rpcServer.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newReadFactory(rpcServer, nil, ios)

	cmd := read.NewReadCmd(f)
	cmd.SetArgs([]string{
		"0x6B175474E89094C44Da98b954EedeAC495271d0F",
		"balanceOf(address)",
		"0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
		"--chain", "1",
	})
	err := cmd.Execute()
	require.NoError(t, err)

	out := strings.TrimSpace(outBuf.String())
	assert.Equal(t, "100", out)
}

func TestReadCmd_RawOutput(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x0000000000000000000000000000000000000000000000000000000000000006",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer rpcServer.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newReadFactory(rpcServer, nil, ios)

	cmd := read.NewReadCmd(f)
	cmd.SetArgs([]string{
		"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		"decimals()",
		"--chain", "1",
		"--raw",
	})
	err := cmd.Execute()
	require.NoError(t, err)

	out := strings.TrimSpace(outBuf.String())
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000006", out)
}

func TestReadCmd_RPCError(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"error":   map[string]interface{}{"code": -32000, "message": "execution reverted"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer rpcServer.Close()

	ios, _, _, _ := iostreams.Test()
	f := newReadFactory(rpcServer, nil, ios)

	cmd := read.NewReadCmd(f)
	cmd.SetArgs([]string{
		"0xdAC17F958D2ee523a2206206994597C96e3cFa0e",
		"totalSupply()",
		"--chain", "1",
	})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution reverted")
}

func TestReadCmd_WrongArgCount(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer rpcServer.Close()

	ios, _, _, _ := iostreams.Test()
	f := newReadFactory(rpcServer, nil, ios)

	cmd := read.NewReadCmd(f)
	cmd.SetArgs([]string{
		"0xdAC17F958D2ee523a2206206994597C96e3cFa0e",
		"balanceOf(address)",
		// Missing the address argument
		"--chain", "1",
	})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 1 argument(s)")
}

func TestReadCmd_RpcUrlFlag(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x0000000000000000000000000000000000000000000000000000000000000001",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer rpcServer.Close()

	ios, outBuf, _, _ := iostreams.Test()
	// Config has no RPC for chain 999
	f := &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) {
			return khhttp.NewClient(khhttp.ClientOptions{Host: "test", AppVersion: "1.0.0"}), nil
		},
		Config:  func() (config.Config, error) { return config.Config{}, nil },
		BaseURL: func() string { return "" },
	}

	cmd := read.NewReadCmd(f)
	cmd.SetArgs([]string{
		"0xdAC17F958D2ee523a2206206994597C96e3cFa0e",
		"totalSupply()",
		"--chain", "999",
		"--rpc-url", rpcServer.URL,
	})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "1", strings.TrimSpace(outBuf.String()))
}

func TestReadCmd_HasCallAlias(t *testing.T) {
	cmd := read.NewReadCmd(&cmdutil.Factory{IOStreams: &iostreams.IOStreams{}})
	assert.Contains(t, cmd.Aliases, "call")
}
