package auth

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewLoginCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "login",
		Short: "Log in to KeeperHub",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[auth login] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Bool("no-browser", false, "Do not open a browser window")
	cmd.Flags().Bool("with-token", false, "Read token from stdin")

	return cmd
}
