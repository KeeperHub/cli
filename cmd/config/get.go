package config

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewGetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			var value string
			switch key {
			case "default_host":
				value = cfg.DefaultHost
			default:
				return fmt.Errorf("X Config key not set: %s", key)
			}

			if value == "" {
				return fmt.Errorf("X Config key not set: %s", key)
			}

			fmt.Fprintln(f.IOStreams.Out, value)
			return nil
		},
	}

	return cmd
}
