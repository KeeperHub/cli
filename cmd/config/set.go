package config

import (
	"fmt"

	internalconfig "github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var validConfigKeys = []string{"default_host"}

func NewSetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		Long: `Persist a configuration key-value pair to the config file. Changes take
effect immediately on the next command run. Use 'kh config list' to see
all valid keys.

See also: kh config list, kh config get`,
		Example: `  # Set the default host
  kh config set default_host app.keeperhub.io

  # Point CLI at a self-hosted instance
  kh config set default_host https://kh.mycompany.io`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			switch key {
			case "default_host":
				cfg.DefaultHost = value
			default:
				return fmt.Errorf("X Unknown config key: %s\nHint: use 'kh config list' to see available keys", key)
			}

			if err := internalconfig.WriteConfig(cfg); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Set %s to %s\n", key, value)
			return nil
		},
	}

	return cmd
}
