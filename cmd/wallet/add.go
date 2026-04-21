package wallet

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewAddCmd returns the `kh wallet add` subcommand -- a thin wrapper around
// `npx @keeperhub/wallet add` that provisions a new agentic wallet.
func NewAddCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Provision a new agentic wallet (no KeeperHub account required)",
		Long: `Provision a new agentic wallet by calling POST /api/agentic-wallet/provision.

This is a thin wrapper around ` + "`npx @keeperhub/wallet add`" + ` -- the npm package is the
canonical tool. Writes {subOrgId, walletAddress, hmacSecret} to ~/.keeperhub/wallet.json
(chmod 0o600) and prints subOrgId + walletAddress (hmacSecret is NEVER printed).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runNpxWallet(f, cmd, "add", nil)
		},
	}
}
