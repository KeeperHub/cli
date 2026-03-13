package wallet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// BalancesResponse is the API response from GET /api/user/wallet/balances.
type BalancesResponse struct {
	WalletAddress string         `json:"walletAddress"`
	Balances      []ChainBalance `json:"balances"`
}

// ChainBalance holds the balance for a single chain.
type ChainBalance struct {
	ChainID       string         `json:"chainId"`
	ChainName     string         `json:"chainName"`
	NativeBalance string         `json:"nativeBalance"`
	Tokens        []TokenBalance `json:"tokens"`
}

// TokenBalance holds the balance for a single token.
type TokenBalance struct {
	Symbol   string `json:"symbol"`
	Balance  string `json:"balance"`
	Decimals int    `json:"decimals"`
}

var zeroValues = map[string]bool{"0": true, "0.0": true, "0.00": true}

func isZeroBalance(b string) bool {
	return zeroValues[strings.TrimSpace(b)]
}

func hasNonZeroTokens(tokens []TokenBalance) bool {
	for _, t := range tokens {
		if !isZeroBalance(t.Balance) {
			return true
		}
	}
	return false
}

func NewBalanceCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "balance",
		Short:   "Show wallet balance",
		Aliases: []string{"bal"},
		Args:    cobra.NoArgs,
		Example: `  # Show balances for all chains
  kh w balance

  # Filter to a specific chain
  kh w balance --chain Ethereum`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			chainFilter, err := cmd.Flags().GetString("chain")
			if err != nil {
				return err
			}

			host := cfg.DefaultHost
			apiURL := khhttp.BuildBaseURL(host) + "/api/user/wallet/balances"

			req, err := client.NewRequest(http.MethodGet, apiURL, nil)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("executing request: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var balances BalancesResponse
			if err := json.NewDecoder(resp.Body).Decode(&balances); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)

			// JSON mode: return full unfiltered response.
			if p.IsJSON() {
				return p.PrintData(balances, nil)
			}

			// Table mode: filter to non-zero chains, apply --chain flag.
			fmt.Fprintf(f.IOStreams.Out, "Wallet: %s\n\n", balances.WalletAddress)

			visible := make([]ChainBalance, 0, len(balances.Balances))
			for _, chain := range balances.Balances {
				if chainFilter != "" && !strings.EqualFold(chain.ChainName, chainFilter) {
					continue
				}
				if isZeroBalance(chain.NativeBalance) && !hasNonZeroTokens(chain.Tokens) {
					continue
				}
				visible = append(visible, chain)
			}

			if len(visible) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No balances found.")
				return nil
			}

			for _, chain := range visible {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", chain.ChainName)
				tw := output.NewTable(f.IOStreams.Out, false)
				tw.AppendHeader(table.Row{"TOKEN", "BALANCE"})
				tw.AppendRow(table.Row{"ETH", chain.NativeBalance})
				for _, tok := range chain.Tokens {
					if !isZeroBalance(tok.Balance) {
						tw.AppendRow(table.Row{tok.Symbol, tok.Balance})
					}
				}
				tw.Render()
				fmt.Fprintln(f.IOStreams.Out)
			}

			return nil
		},
	}

	cmd.Flags().String("chain", "", "Filter by chain")

	return cmd
}
