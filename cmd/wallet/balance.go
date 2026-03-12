package wallet

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewBalanceCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "balance",
		Short:   "Show wallet balance",
		Aliases: []string{"bal"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[wallet balance] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().String("chain", "", "Filter by chain")

	return cmd
}
