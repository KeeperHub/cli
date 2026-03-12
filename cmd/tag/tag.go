package tag

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewTagCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tag",
		Short:   "Manage tags",
		Aliases: []string{"t"},
	}

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewCreateCmd(f))
	cmd.AddCommand(NewGetCmd(f))
	cmd.AddCommand(NewDeleteCmd(f))

	return cmd
}
