package execute

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewStatusCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <execution-id>",
		Short:   "Show the status of an execution",
		Aliases: []string{"st"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[execute status] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Bool("watch", false, "Live-update until complete")

	return cmd
}
