package run

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewStatusCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <run-id>",
		Short:   "Show the status of a run",
		Aliases: []string{"st"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[run status] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Bool("watch", false, "Live-update until complete")

	return cmd
}
