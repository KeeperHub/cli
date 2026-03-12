package wallet

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewTokensCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tokens",
		Short:   "List wallet tokens",
		Aliases: []string{"tok"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[wallet tokens] is not yet implemented.")
			return nil
		},
	}

	return cmd
}
