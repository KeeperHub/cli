package execute

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewContractCallCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contract-call",
		Short:   "Call a smart contract method",
		Aliases: []string{"cc"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[execute contract-call] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().String("chain", "", "Chain ID")
	cmd.Flags().String("contract", "", "Contract address")
	cmd.Flags().String("method", "", "Method name")
	cmd.Flags().StringSlice("args", nil, "Method arguments")

	return cmd
}
