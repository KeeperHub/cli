package workflow

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewPauseCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "pause <workflow-id>",
		Short: "Pause a workflow",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[workflow pause] is not yet implemented.")
			return nil
		},
	}

	return cmd
}
