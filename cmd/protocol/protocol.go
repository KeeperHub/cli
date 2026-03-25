package protocol

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewProtocolCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "protocol",
		Short:   "Browse blockchain protocols",
		Aliases: []string{"pr", "proto"},
		Example: `  # List all protocols
  kh pr ls

  # Get details for a protocol
  kh pr g uniswap`,
	}

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewGetCmd(f))

	return cmd
}
