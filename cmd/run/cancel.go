package run

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCancelCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "cancel <run-id>",
		Short: "Cancel a run",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[run cancel] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation")

	return cmd
}
