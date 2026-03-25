package chain

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewChainCmd creates the top-level chain command group.
func NewChainCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chain",
		Short:   "Manage blockchain chains",
		Aliases: []string{"ch"},
		Example: `  # List supported chains
  kh ch ls`,
	}

	cmd.AddCommand(NewListCmd(f))

	return cmd
}
