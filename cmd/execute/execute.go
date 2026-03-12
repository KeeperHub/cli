package execute

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewExecuteCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute",
		Short:   "Execute direct blockchain actions",
		Aliases: []string{"ex"},
	}

	cmd.AddCommand(NewTransferCmd(f))
	cmd.AddCommand(NewContractCallCmd(f))
	cmd.AddCommand(NewCheckAndExecuteCmd(f))
	cmd.AddCommand(NewStatusCmd(f))

	return cmd
}
