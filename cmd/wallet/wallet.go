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
	}

	cmd.AddCommand(NewBalanceCmd(f))
	cmd.AddCommand(NewTokensCmd(f))

	return cmd
}
