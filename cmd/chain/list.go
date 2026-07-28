package chain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/internal/rpc"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewListCmd creates the chain list subcommand.
func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List supported blockchain chains",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Long: `List supported blockchain chains.

Note the type change if you are copying a chain id into a workflow. This
command reports "chainId" as a JSON number, but web3 workflow nodes expect
config.network as a STRING:

  kh chain list --json   ->   "chainId": 103
  workflow node config   ->   "network": "103"

Writing the number form into a node config is accepted by the API and fails
later at execution time, so it is worth getting right up front.

The list reflects the chains table in the target environment, which is
seeded on deploy. It is the authoritative answer for which chains are live;
the chain-config repo is only an input to that seed.`,
		Example: `  # List all chains
  kh ch ls

  # List chains as JSON
  kh ch ls --json

  # Chain id in the string form a workflow node config expects
  kh ch ls --json | jq -r '.[] | select(.name == "Solana Devnet") | .chainId | tostring'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			chains, err := fetchChains(f, cmd)
			if err != nil {
				return err
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			if len(chains) == 0 && !p.IsJSON() {
				fmt.Fprintln(f.IOStreams.Out, "No chains found.")
				return nil
			}
			return p.PrintData(chains, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"CHAIN ID", "NAME", "TYPE", "ENABLED"})
				for _, ch := range chains {
					tw.AppendRow(table.Row{ch.ChainID, ch.Name, ch.Type, ch.IsEnabled})
				}
				tw.Render()
			})
		},
	}

	return cmd
}

// fetchChains calls /api/chains and caches the result for RPC resolution.
func fetchChains(f *cmdutil.Factory, cmd *cobra.Command) ([]rpc.ChainInfo, error) {
	client, err := f.HTTPClient()
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client: %w", err)
	}

	cfg, err := f.Config()
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	host := cmdutil.ResolveHost(cmd, cfg)
	url := khhttp.BuildBaseURL(host) + "/api/chains"

	req, err := client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, khhttp.NewAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Cache the chain data for RPC resolution
	_ = rpc.CacheChains(json.RawMessage(body))

	var chains []rpc.ChainInfo
	if err := json.Unmarshal(body, &chains); err != nil {
		return nil, fmt.Errorf("decoding chains response: %w", err)
	}

	return chains, nil
}
