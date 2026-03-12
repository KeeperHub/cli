package execute

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewTransferCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfer",
		Short:   "Transfer tokens",
		Aliases: []string{"t"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[execute transfer] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().String("chain", "", "Chain ID (required)")
	cmd.Flags().String("to", "", "Recipient address (required)")
	cmd.Flags().String("amount", "", "Amount to transfer (required)")
	cmd.Flags().String("token", "ETH", "Token symbol")

	return cmd
}
