package run

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewLogsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logs <run-id>",
		Short:   "Show logs for a run",
		Aliases: []string{"l"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[run logs] is not yet implemented.")
			return nil
		},
	}

	return cmd
}
