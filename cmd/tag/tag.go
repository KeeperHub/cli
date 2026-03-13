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
		Example: `  # List all tags
  kh t ls

  # Create a new tag
  kh t create "my-tag"`,
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
