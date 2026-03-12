package serve

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewServeCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "serve",
		Short: "Start a server",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[serve] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Bool("mcp", false, "Start MCP stdio server")

	return cmd
}
