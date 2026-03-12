package workflow

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewRunCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run <workflow-id>",
		Short:   "Run a workflow",
		Aliases: []string{"r"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[workflow run] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Bool("wait", false, "Wait for completion")

	return cmd
}
