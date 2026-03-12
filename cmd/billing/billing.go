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
	}

	cmd.AddCommand(NewStatusCmd(f))
	cmd.AddCommand(NewUsageCmd(f))

	return cmd
}
