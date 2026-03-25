package read

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"golang.org/x/crypto/sha3"

	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/internal/rpc"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewReadCmd creates the top-level read command for calling smart contract view functions.
func NewReadCmd(f *cmdutil.Factory) *cobra.Command {
	var chainID string
	var rpcURL string
	var block string
	var raw bool

	cmd := &cobra.Command{
		Use:   "read <contract-address> <function-signature> [args...]",
		Short: "Read a smart contract view function",
		Long: `Call a read-only smart contract function via eth_call.

The function signature should be in Solidity format (e.g. "balanceOf(address)").
Arguments are positional and must match the function signature types.

Supported argument types: address, uint256, bool, bytes32.
No auth required; uses public RPC endpoints by default.`,
		Aliases: []string{"call"},
		Args:    cobra.MinimumNArgs(2),
		Example: `  # Read USDT total supply
  kh read 0xdAC17F958D2ee523a2206206994597C96e3cFa0e "totalSupply()" --chain 1

  # Read ERC-20 balance
  kh read 0x6B175474E89094C44Da98b954EedeAC495271d0F "balanceOf(address)" 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045 --chain 1

  # Read token decimals
  kh read 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 "decimals()" --chain 1

  # Use a custom RPC endpoint
  kh read 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 "decimals()" --chain 1 --rpc-url https://eth.llamarpc.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if chainID == "" {
				return cmdutil.FlagError{Err: fmt.Errorf("--chain is required")}
			}

			contractAddr := args[0]
			funcSig := args[1]
			funcArgs := args[2:]

			cfg, err := f.Config()
			if err != nil {
				cfg = config.DefaultConfig()
			}

			chains := loadChainsForRPC(f)

			endpoint, err := rpc.Resolve(chainID, rpcURL, cfg, chains)
			if err != nil {
				return err
			}

			calldata, err := encodeCall(funcSig, funcArgs)
			if err != nil {
				return fmt.Errorf("encoding call: %w", err)
			}

			blockTag := "latest"
			if block != "" {
				blockTag = block
			}

			result, err := ethCall(endpoint, contractAddr, calldata, blockTag)
			if err != nil {
				return err
			}

			if raw {
				fmt.Fprintln(f.IOStreams.Out, result)
				return nil
			}

			decoded := decodeResult(funcSig, result)
			fmt.Fprintln(f.IOStreams.Out, decoded)
			return nil
		},
	}

	cmd.Flags().StringVar(&chainID, "chain", "", "Chain ID (required)")
	cmd.Flags().StringVar(&rpcURL, "rpc-url", "", "Override RPC endpoint")
	cmd.Flags().StringVar(&block, "block", "", "Block number or tag (default: latest)")
	cmd.Flags().BoolVar(&raw, "raw", false, "Output raw hex instead of decoded")

	return cmd
}

// loadChainsForRPC attempts to load cached chain data for RPC resolution.
// Returns nil on any error since chain data is optional for RPC resolution
// (the user may have --rpc-url or config entries).
func loadChainsForRPC(f *cmdutil.Factory) []rpc.ChainInfo {
	chains, err := rpc.LoadChains()
	if err != nil {
		// Try fetching from platform
		chains = fetchAndCacheChains(f)
	}
	return chains
}

// fetchAndCacheChains fetches chain data from the platform API and caches it.
// Returns nil if the fetch fails.
func fetchAndCacheChains(f *cmdutil.Factory) []rpc.ChainInfo {
	client, err := f.HTTPClient()
	if err != nil {
		return nil
	}

	baseURL := f.BaseURL()
	url := baseURL + "/api/chains"

	req, err := client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	_ = rpc.CacheChains(json.RawMessage(body))

	var chains []rpc.ChainInfo
	if err := json.Unmarshal(body, &chains); err != nil {
		return nil
	}
	return chains
}

// keccak256 computes the Keccak-256 hash of data.
func keccak256(data []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	return h.Sum(nil)
}

// functionSelector returns the 4-byte function selector for a Solidity function signature.
func functionSelector(sig string) []byte {
	hash := keccak256([]byte(sig))
	return hash[:4]
}

// parseParamTypes extracts parameter type names from a function signature.
// e.g. "balanceOf(address)" -> ["address"], "totalSupply()" -> []
func parseParamTypes(sig string) []string {
	openParen := strings.Index(sig, "(")
	closeParen := strings.LastIndex(sig, ")")
	if openParen < 0 || closeParen < 0 || closeParen <= openParen+1 {
		return nil
	}
	inner := sig[openParen+1 : closeParen]
	if strings.TrimSpace(inner) == "" {
		return nil
	}
	parts := strings.Split(inner, ",")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimSpace(p)
	}
	return result
}

// encodeCall encodes a function call into ABI-encoded calldata.
func encodeCall(funcSig string, args []string) (string, error) {
	selector := functionSelector(funcSig)
	paramTypes := parseParamTypes(funcSig)

	if len(args) != len(paramTypes) {
		return "", fmt.Errorf("expected %d argument(s) for %s, got %d", len(paramTypes), funcSig, len(args))
	}

	var buf bytes.Buffer
	buf.Write(selector)

	for i, arg := range args {
		encoded, err := abiEncodeParam(paramTypes[i], arg)
		if err != nil {
			return "", fmt.Errorf("encoding argument %d (%s): %w", i, paramTypes[i], err)
		}
		buf.Write(encoded)
	}

	return "0x" + hex.EncodeToString(buf.Bytes()), nil
}

// abiEncodeParam encodes a single parameter value as a 32-byte ABI word.
func abiEncodeParam(paramType, value string) ([]byte, error) {
	word := make([]byte, 32)

	switch {
	case paramType == "address":
		addr := strings.TrimPrefix(value, "0x")
		if len(addr) != 40 {
			return nil, fmt.Errorf("invalid address length: %s", value)
		}
		decoded, err := hex.DecodeString(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid address hex: %w", err)
		}
		copy(word[12:], decoded)

	case strings.HasPrefix(paramType, "uint"):
		n := new(big.Int)
		if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
			_, ok := n.SetString(value[2:], 16)
			if !ok {
				return nil, fmt.Errorf("invalid hex uint: %s", value)
			}
		} else {
			_, ok := n.SetString(value, 10)
			if !ok {
				return nil, fmt.Errorf("invalid uint: %s", value)
			}
		}
		b := n.Bytes()
		if len(b) > 32 {
			return nil, fmt.Errorf("uint value too large for 32 bytes")
		}
		copy(word[32-len(b):], b)

	case strings.HasPrefix(paramType, "int"):
		n := new(big.Int)
		_, ok := n.SetString(value, 10)
		if !ok {
			return nil, fmt.Errorf("invalid int: %s", value)
		}
		if n.Sign() >= 0 {
			b := n.Bytes()
			if len(b) > 32 {
				return nil, fmt.Errorf("int value too large for 32 bytes")
			}
			copy(word[32-len(b):], b)
		} else {
			// Two's complement for negative numbers
			twos := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 256), n)
			b := twos.Bytes()
			// Fill with 0xFF for sign extension
			for j := range word {
				word[j] = 0xFF
			}
			copy(word[32-len(b):], b)
		}

	case paramType == "bool":
		switch strings.ToLower(value) {
		case "true", "1":
			word[31] = 1
		case "false", "0":
			// already zero
		default:
			return nil, fmt.Errorf("invalid bool: %s", value)
		}

	case paramType == "bytes32":
		b := strings.TrimPrefix(value, "0x")
		decoded, err := hex.DecodeString(b)
		if err != nil {
			return nil, fmt.Errorf("invalid bytes32 hex: %w", err)
		}
		if len(decoded) > 32 {
			return nil, fmt.Errorf("bytes32 value too long")
		}
		copy(word[:len(decoded)], decoded)

	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", paramType)
	}

	return word, nil
}

// jsonRPCRequest is the JSON-RPC 2.0 request envelope.
type jsonRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// jsonRPCResponse is the JSON-RPC 2.0 response envelope.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  string          `json:"result"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

// jsonRPCError is a JSON-RPC error object.
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ethCallParams matches the eth_call params object.
type ethCallParams struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

// ethCall performs an eth_call JSON-RPC request.
func ethCall(rpcEndpoint, to, data, blockTag string) (string, error) {
	reqBody := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_call",
		Params: []interface{}{
			ethCallParams{To: to, Data: data},
			blockTag,
		},
		ID: 1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshalling request: %w", err)
	}

	resp, err := http.Post(rpcEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("RPC request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("RPC returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var rpcResp jsonRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return "", fmt.Errorf("decoding RPC response: %w", err)
	}

	if rpcResp.Error != nil {
		return "", fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// decodeResult attempts to decode a hex result based on the function signature's return type.
// For simple return types (single uint256, address, bool), it decodes the value.
// For complex or unknown types, it returns the raw hex.
func decodeResult(funcSig string, hexResult string) string {
	if hexResult == "" || hexResult == "0x" {
		return "0x (empty result)"
	}

	// Strip 0x prefix
	clean := strings.TrimPrefix(hexResult, "0x")
	if len(clean) == 0 {
		return "0x (empty result)"
	}

	// For a single 32-byte word, try common return types
	if len(clean) == 64 {
		// Try as uint256
		n := new(big.Int)
		n.SetString(clean, 16)
		return n.String()
	}

	// Multiple words or non-standard length: return raw hex
	return hexResult
}
