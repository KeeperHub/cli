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

	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")
	cmd.PersistentFlags().BoolP("yes", "y", false, "Skip confirmation prompts")

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewCreateCmd(f))
	cmd.AddCommand(NewGetCmd(f))
	cmd.AddCommand(NewDeleteCmd(f))

	return cmd
}
