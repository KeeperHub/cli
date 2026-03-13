package config

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewConfigCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Example: `  # List all config values
  kh config ls

  # Set the default host
  kh config set default_host app.keeperhub.io`,
	}

	cmd.AddCommand(NewSetCmd(f))
	cmd.AddCommand(NewGetCmd(f))
	cmd.AddCommand(NewListCmd(f))

	return cmd
}
