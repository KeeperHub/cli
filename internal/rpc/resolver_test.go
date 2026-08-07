package rpc_test

import (
	"testing"

	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/internal/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_FlagValueFirst(t *testing.T) {
	t.Setenv("KH_RPC_URL", "https://env-rpc.example.com")
	cfg := config.Config{RPC: map[string]string{"1": "https://config-rpc.example.com"}}

	result, err := rpc.Resolve("1", "https://flag-rpc.example.com", cfg)
	require.NoError(t, err)
	assert.Equal(t, "https://flag-rpc.example.com", result)
}

func TestResolve_EnvVarSecond(t *testing.T) {
	t.Setenv("KH_RPC_URL", "https://env-rpc.example.com")
	cfg := config.Config{RPC: map[string]string{"1": "https://config-rpc.example.com"}}

	result, err := rpc.Resolve("1", "", cfg)
	require.NoError(t, err)
	assert.Equal(t, "https://env-rpc.example.com", result)
}

func TestResolve_ConfigThird(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{RPC: map[string]string{"1": "https://config-rpc.example.com"}}

	result, err := rpc.Resolve("1", "", cfg)
	require.NoError(t, err)
	assert.Equal(t, "https://config-rpc.example.com", result)
}

func TestResolve_ConfigMissForRequestedChain(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{RPC: map[string]string{"1": "https://config-rpc.example.com"}}

	_, err := rpc.Resolve("137", "", cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no RPC endpoint found for chain 137")
}

func TestResolve_ErrorWhenNothingFound(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{}

	_, err := rpc.Resolve("42161", "", cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no RPC endpoint found for chain 42161")
	assert.Contains(t, err.Error(), "--rpc-url")
}

func TestResolve_NilConfigRPC(t *testing.T) {
	t.Setenv("KH_RPC_URL", "")
	cfg := config.Config{}

	_, err := rpc.Resolve("1", "", cfg)
	assert.Error(t, err)
}
