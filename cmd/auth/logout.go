package auth

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewLogoutCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "logout",
		Short: "Log out of KeeperHub",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[auth logout] is not yet implemented.")
			return nil
		},
	}

	return cmd
}
