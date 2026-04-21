package wallet

import (
	"fmt"
	"os"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewLinkCmd returns the `kh wallet link` subcommand -- a thin wrapper around
// `npx @keeperhub/wallet link` that links an agentic wallet to a KeeperHub account.
//
// Requires KH_SESSION_COOKIE to be set (v0.1.0 env-var contract from Phase 34 Plan 06).
func NewLinkCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Link the agentic wallet to a KeeperHub account (requires KH_SESSION_COOKIE)",
		Long: `Link the current agentic wallet to your KeeperHub account by calling POST /api/agentic-wallet/link.

Thin wrapper around ` + "`npx @keeperhub/wallet link`" + `. Requires the KH_SESSION_COOKIE env var
set to a valid kh session cookie (sign in at app.keeperhub.com, copy the session cookie, export it).

v0.1.0 does not launch a browser session handshake; this env-var contract matches the npm CLI.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if os.Getenv("KH_SESSION_COOKIE") == "" {
				return fmt.Errorf("KH_SESSION_COOKIE env var is required: sign in at app.keeperhub.com, copy the session cookie, and re-run with KH_SESSION_COOKIE='<cookie>' kh wallet link")
			}
			return runNpxWallet(f, cmd, "link", nil)
		},
	}
}
