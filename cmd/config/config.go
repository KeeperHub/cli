package config

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewConfigCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	cmd.AddCommand(NewSetCmd(f))
	cmd.AddCommand(NewGetCmd(f))
	cmd.AddCommand(NewListCmd(f))

	return cmd
}
