package billing

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewBillingCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "billing",
		Short:   "View billing and usage",
		Aliases: []string{"b"},
		Example: `  # Show billing status
  kh b st

  # Show usage for current period
  kh b u`,
	}

	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")

	cmd.AddCommand(NewStatusCmd(f))
	cmd.AddCommand(NewUsageCmd(f))

	return cmd
}
