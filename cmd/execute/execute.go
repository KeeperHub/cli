package execute

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewExecuteCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute",
		Short:   "Execute direct blockchain actions",
		Aliases: []string{"ex", "exec"},
		Long: `Execute blockchain operations directly without building a full workflow.
Supports token transfers and smart contract calls. Returns execution IDs
immediately; use --wait to block until completion.

See also: kh r st, kh wf run`,
		Example: `  # Transfer ETH on a chain
  kh ex transfer --chain 1 --to 0xABCD... --amount 0.01

  # Call a smart contract method
  kh ex cc --chain 1 --contract 0x... --method balanceOf --args '["0x..."]'`,
	}

	cmd.AddCommand(NewTransferCmd(f))
	cmd.AddCommand(NewContractCallCmd(f))
	cmd.AddCommand(NewStatusCmd(f))

	return cmd
}
