package rpc

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/keeperhub/cli/internal/cache"
	"github.com/keeperhub/cli/internal/config"
)

const (
	// ChainsCacheName is the filename for the cached /api/chains response.
	ChainsCacheName = "chains.json"

	// ChainsCacheTTL is the TTL for chain cache entries.
	ChainsCacheTTL = 1 * time.Hour
)

// ChainInfo holds the RPC endpoint data for a single chain as returned by /api/chains.
type ChainInfo struct {
	ChainID        int    `json:"chainId"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	PrimaryRpcURL  string `json:"primaryRpcUrl"`
	FallbackRpcURL string `json:"fallbackRpcUrl"`
}

// Resolve returns the RPC endpoint URL for a given chain ID.
// Resolution order:
//  1. flagValue (from --rpc-url flag)
//  2. KH_RPC_URL env var
//  3. Config file rpc.<chainID>
//  4. Platform /api/chains primaryRpcUrl (from cached chain data)
//  5. Platform /api/chains fallbackRpcUrl
//  6. Error if nothing found
func Resolve(chainID string, flagValue string, cfg config.Config, chains []ChainInfo) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	if envURL := os.Getenv("KH_RPC_URL"); envURL != "" {
		return envURL, nil
	}

	if url := cfg.RPCEndpoint(chainID); url != "" {
		return url, nil
	}

	for _, ch := range chains {
		if fmt.Sprintf("%d", ch.ChainID) == chainID {
			if ch.PrimaryRpcURL != "" {
				return ch.PrimaryRpcURL, nil
			}
			if ch.FallbackRpcURL != "" {
				return ch.FallbackRpcURL, nil
			}
		}
	}

	return "", fmt.Errorf("no RPC endpoint found for chain %s. Set one with --rpc-url, KH_RPC_URL, or config rpc.%s", chainID, chainID)
}

// LoadChains reads chain data from cache, returning nil if no cached data is available.
// The caller is responsible for fetching and caching fresh data when this returns nil.
func LoadChains() ([]ChainInfo, error) {
	entry, err := cache.ReadCache(ChainsCacheName)
	if err != nil {
		return nil, err
	}
	if cache.IsStale(entry, ChainsCacheTTL) {
		return nil, fmt.Errorf("chain cache is stale")
	}
	return unmarshalChains(entry.Data)
}

// CacheChains writes chain data to the local cache.
func CacheChains(data json.RawMessage) error {
	return cache.WriteCache(ChainsCacheName, data)
}

// unmarshalChains parses the raw JSON array of chain objects.
func unmarshalChains(raw json.RawMessage) ([]ChainInfo, error) {
	var chains []ChainInfo
	if err := json.Unmarshal(raw, &chains); err != nil {
		return nil, fmt.Errorf("decoding chains data: %w", err)
	}
	return chains, nil
}
