package action

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewActionCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "action",
		Short:   "Browse available actions",
		Aliases: []string{"a"},
	}

	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewGetCmd(f))

	return cmd
}
