package template

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewDeployCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deploy <template-id>",
		Short:   "Deploy a workflow template",
		Aliases: []string{"d"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[template deploy] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().String("name", "", "Workflow name")

	return cmd
}
