package execute

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCheckAndExecuteCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "check-and-execute",
		Short:   "Check a condition and execute if true",
		Aliases: []string{"cae"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[execute check-and-execute] is not yet implemented.")
			return nil
		},
	}

	return cmd
}
