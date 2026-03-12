package workflow

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewGoLiveCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "go-live <workflow-id>",
		Short:   "Activate a workflow",
		Aliases: []string{"live"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[workflow go-live] is not yet implemented.")
			return nil
		},
	}

	return cmd
}
