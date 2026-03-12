package workflow

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewGetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <workflow-id>",
		Short:   "Get a workflow",
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[workflow get] is not yet implemented.")
			return nil
		},
	}

	return cmd
}
