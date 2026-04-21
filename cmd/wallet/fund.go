package wallet

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewFundCmd returns the `kh wallet fund` subcommand -- a thin wrapper around
// `npx @keeperhub/wallet fund` that prints funding instructions for the agentic wallet.
func NewFundCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "fund",
		Short: "Print Coinbase Onramp URL (Base USDC) and Tempo deposit address for the agentic wallet",
		Long: `Print a Coinbase Onramp URL for Base USDC funding plus the Tempo deposit address.

Thin wrapper around ` + "`npx @keeperhub/wallet fund`" + `. No HTTP calls, no browser launch --
prints copy-paste instructions only.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runNpxWallet(f, cmd, "fund", nil)
		},
	}
}
