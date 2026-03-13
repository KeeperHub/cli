package auth

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewAuthCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with KeeperHub",
		Example: `  # Log in via browser
  kh auth login

  # Check current auth status
  kh auth status`,
	}

	cmd.AddCommand(NewLoginCmd(f))
	cmd.AddCommand(NewLogoutCmd(f))
	cmd.AddCommand(NewStatusCmd(f))

	return cmd
}
