package wallet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// TokensResponse is the API response from GET /api/user/wallet/tokens.
type TokensResponse struct {
	Tokens []Token `json:"tokens"`
}

// Token holds metadata for a single supported token.
type Token struct {
	ChainID      string `json:"chainId"`
	TokenAddress string `json:"tokenAddress"`
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Decimals     int    `json:"decimals"`
}

func NewTokensCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tokens",
		Short:   "List wallet tokens",
		Aliases: []string{"tok"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return err
			}

			chain, err := cmd.Flags().GetString("chain")
			if err != nil {
				return err
			}

			host := cfg.DefaultHost
			apiURL := khhttp.BuildBaseURL(host) + "/api/user/wallet/tokens?limit=" + strconv.Itoa(limit)
			if chain != "" {
				apiURL += "&chain=" + chain
			}

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

			var tokensResp TokensResponse
			if err := json.NewDecoder(resp.Body).Decode(&tokensResp); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(tokensResp.Tokens, func(tw table.Writer) {
				if len(tokensResp.Tokens) == 0 {
					fmt.Fprintln(f.IOStreams.Out, "No tokens found.")
					return
				}
				tw.AppendHeader(table.Row{"CHAIN", "SYMBOL", "NAME", "ADDRESS"})
				for _, tok := range tokensResp.Tokens {
					tw.AppendRow(table.Row{tok.ChainID, tok.Symbol, tok.Name, tok.TokenAddress})
				}
				tw.Render()
			})
		},
	}

	cmd.Flags().Int("limit", 50, "Maximum number of tokens to list")
	cmd.Flags().String("chain", "", "Filter by chain")

	return cmd
}
