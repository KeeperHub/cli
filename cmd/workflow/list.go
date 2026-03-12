package workflow

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List workflows",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[workflow list] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Int("limit", 30, "Maximum number of workflows to list")
	cmd.Flags().String("status", "", "Filter by status")

	return cmd
}
