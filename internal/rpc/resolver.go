package rpc

import (
	"fmt"
	"os"

	"github.com/keeperhub/cli/internal/config"
)

// ChainInfo holds a single chain as returned by /api/chains.
type ChainInfo struct {
	ChainID   int    `json:"chainId"`
	Name      string `json:"name"`
	Type      string `json:"chainType"`
	IsEnabled bool   `json:"isEnabled"`
}

// Resolve returns the RPC endpoint URL for a given chain ID.
//
// The platform does not hand out RPC endpoints: /api/chains omits the seeded
// defaults because some of them embed provider API keys. The caller supplies
// the endpoint.
//
// Resolution order:
//  1. flagValue (from --rpc-url flag)
//  2. KH_RPC_URL env var
//  3. Config file rpc.<chainID>
//  4. Error if nothing found
func Resolve(chainID string, flagValue string, cfg config.Config) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	if envURL := os.Getenv("KH_RPC_URL"); envURL != "" {
		return envURL, nil
	}

	if url := cfg.RPCEndpoint(chainID); url != "" {
		return url, nil
	}

	return "", fmt.Errorf("no RPC endpoint found for chain %s. Set one with --rpc-url, KH_RPC_URL, or config rpc.%s", chainID, chainID)
}
