package template

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewTemplateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Short:   "Manage workflow templates",
		Aliases: []string{"tp"},
	}

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewDeployCmd(f))

	return cmd
}
