package wallet

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewInfoCmd returns the `kh wallet info` subcommand -- a thin wrapper around
// `npx @keeperhub/wallet info` that prints the local agentic wallet identity.
func NewInfoCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print subOrgId and walletAddress from local agentic wallet config",
		Long: `Print subOrgId and walletAddress from ~/.keeperhub/wallet.json.

Thin wrapper around ` + "`npx @keeperhub/wallet info`" + `. Exits non-zero if the config is missing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runNpxWallet(f, cmd, "info", nil)
		},
	}
}
