package protocol

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewProtocolCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Short:   "Browse available plugins and integrations",
		Aliases: []string{"plugins", "protocol", "pr", "proto"},
		Example: `  # List all plugins
  kh plugin ls

  # Get details for a plugin
  kh plugin g aave`,
	}

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewGetCmd(f))

	return cmd
}
