package run

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewRunCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Monitor workflow runs",
		Aliases: []string{"r"},
	}

	cmd.AddCommand(NewStatusCmd(f))
	cmd.AddCommand(NewLogsCmd(f))
	cmd.AddCommand(NewCancelCmd(f))

	return cmd
}
