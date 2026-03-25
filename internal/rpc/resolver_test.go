package rpc_test

import (
	"encoding/json"
	"testing"

	"github.com/keeperhub/cli/internal/cache"
	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/internal/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleChains() []rpc.ChainInfo {
	return []rpc.ChainInfo{
		{ChainID: 1, Name: "Ethereum", Type: "mainnet", Status: "active", PrimaryRpcURL: "https://eth-rpc.example.com", FallbackRpcURL: "https://eth-fallback.example.com"},
		{ChainID: 137, Name: "Polygon", Type: "mainnet", Status: "active", PrimaryRpcURL: "", FallbackRpcURL: "https://polygon-fallback.example.com"},
		{ChainID: 42161, Name: "Arbitrum", Type: "mainnet", Status: "active", PrimaryRpcURL: "", FallbackRpcURL: ""},
	}
}

func TestResolve_FlagValueFirst(t *testing.T) {
	cfg := config.Config{RPC: map[string]string{"1": "https://config-rpc.example.com"}}
	chains := sampleChains()

	result, err := rpc.Resolve("1", "https://flag-rpc.example.com", cfg, chains)
	require.NoError(t, err)
	assert.Equal(t, "https://flag-rpc.example.com", result)
}

func TestResolve_EnvVarSecond(t *testing.T) {
	t.Setenv("KH_RPC_URL", "https://env-rpc.example.com")
	cfg := config.Config{RPC: map[string]string{"1": "https://config-rpc.example.com"}}
	chains := sampleChains()

	result, err := rpc.Resolve("1", "", cfg, chains)
	require.NoError(t, err)
	assert.Equal(t, "https://env-rpc.example.com", result)
}

func TestResolve_ConfigThird(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{RPC: map[string]string{"1": "https://config-rpc.example.com"}}
	chains := sampleChains()

	result, err := rpc.Resolve("1", "", cfg, chains)
	require.NoError(t, err)
	assert.Equal(t, "https://config-rpc.example.com", result)
}

func TestResolve_PrimaryRPCFourth(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{}
	chains := sampleChains()

	result, err := rpc.Resolve("1", "", cfg, chains)
	require.NoError(t, err)
	assert.Equal(t, "https://eth-rpc.example.com", result)
}

func TestResolve_FallbackRPCFifth(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{}
	chains := sampleChains()

	result, err := rpc.Resolve("137", "", cfg, chains)
	require.NoError(t, err)
	assert.Equal(t, "https://polygon-fallback.example.com", result)
}

func TestResolve_ErrorWhenNothingFound(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{}
	chains := sampleChains()

	_, err := rpc.Resolve("42161", "", cfg, chains)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no RPC endpoint found for chain 42161")
}

func TestResolve_UnknownChainError(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{}

	_, err := rpc.Resolve("999", "", cfg, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no RPC endpoint found for chain 999")
}

func TestResolve_NilConfigRPC(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{}
	chains := sampleChains()

	result, err := rpc.Resolve("1", "", cfg, chains)
	require.NoError(t, err)
	assert.Equal(t, "https://eth-rpc.example.com", result)
}

func TestLoadChains_MissingCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	chains, err := rpc.LoadChains()
	assert.Error(t, err)
	assert.Nil(t, chains)
}

func TestLoadChains_ValidCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	data, _ := json.Marshal(sampleChains())
	require.NoError(t, cache.WriteCache(rpc.ChainsCacheName, data))

	chains, err := rpc.LoadChains()
	require.NoError(t, err)
	assert.Len(t, chains, 3)
	assert.Equal(t, "Ethereum", chains[0].Name)
}

func TestCacheChains(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	data, _ := json.Marshal(sampleChains())
	err := rpc.CacheChains(data)
	require.NoError(t, err)

	// Verify we can read it back
	chains, err := rpc.LoadChains()
	require.NoError(t, err)
	assert.Len(t, chains, 3)
}
