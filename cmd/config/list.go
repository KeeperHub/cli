package config

import (
	"fmt"

	internalconfig "github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all configuration values",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Example: `  # List all config values
  kh config ls`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "default_host=%s\n", cfg.DefaultHost)

			hosts, err := internalconfig.ReadHosts()
			if err != nil {
				return err
			}

			for hostname, entry := range hosts.Hosts {
				if entry.User != "" {
					fmt.Fprintf(f.IOStreams.Out, "hosts.%s.user=%s\n", hostname, entry.User)
				}
			}

			return nil
		},
	}

	return cmd
}
