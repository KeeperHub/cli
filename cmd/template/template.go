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
		Example: `  # List available templates
  kh tp ls

  # Deploy a template to your account
  kh tp deploy abc123`,
	}

	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewDeployCmd(f))

	return cmd
}
