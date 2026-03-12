package project

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCreateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a project",
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[project create] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().String("name", "", "Project name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().String("description", "", "Project description")

	return cmd
}
