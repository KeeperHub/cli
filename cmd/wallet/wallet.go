package wallet

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewWalletCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "wallet",
		Short:   "Manage wallets",
		Aliases: []string{"w"},
		Example: `  # Show wallet balance
  kh w balance

  # List supported tokens
  kh w tokens`,
	}

	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")

	cmd.AddCommand(NewBalanceCmd(f))
	cmd.AddCommand(NewTokensCmd(f))

	return cmd
}
