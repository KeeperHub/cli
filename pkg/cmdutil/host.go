package cmdutil

import (
	"os"

	"github.com/keeperhub/cli/internal/config"
	"github.com/spf13/cobra"
)

// ResolveHost resolves the target host using the priority chain:
// --host flag > KH_HOST env > cfg.DefaultHost > "app.keeperhub.com".
func ResolveHost(cmd *cobra.Command, cfg config.Config) string {
	if cmd != nil {
		root := cmd.Root()
		if root != nil {
			if fl := root.PersistentFlags().Lookup("host"); fl != nil {
				if v := fl.Value.String(); v != "" {
					return v
				}
			}
		}
	}

	if env := os.Getenv("KH_HOST"); env != "" {
		return env
	}

	if cfg.DefaultHost != "" {
		return cfg.DefaultHost
	}

	return "app.keeperhub.com"
}
