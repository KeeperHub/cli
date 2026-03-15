package workflow

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewWorkflowCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workflow",
		Short:   "Manage workflows",
		Aliases: []string{"wf"},
		Example: `  # List workflows
  kh wf ls

  # Run a workflow
  kh wf run abc123`,
	}

	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewRunCmd(f))
	cmd.AddCommand(NewGetCmd(f))
	cmd.AddCommand(NewCreateCmd(f))
	cmd.AddCommand(NewDeleteCmd(f))
	cmd.AddCommand(NewUpdateCmd(f))
	cmd.AddCommand(NewGoLiveCmd(f))
	cmd.AddCommand(NewPauseCmd(f))

	return cmd
}
