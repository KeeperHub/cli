package project

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewProjectCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Manage projects",
		Aliases: []string{"p"},
	}

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewCreateCmd(f))
	cmd.AddCommand(NewGetCmd(f))
	cmd.AddCommand(NewDeleteCmd(f))

	return cmd
}
