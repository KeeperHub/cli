package billing

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewUsageCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "usage",
		Short:   "Show billing usage",
		Aliases: []string{"u"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[billing usage] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().String("period", "current", "Billing period (e.g. 2026-03)")

	return cmd
}
