package main

import (
	"fmt"
	"os"

	"github.com/keeperhub/cli/cmd"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/version"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

func main() {
	ios := iostreams.System()

	// rootCmd is created first so the HTTPClient closure can read --host after flag parsing.
	var rootCmd *cobra.Command

	f := &cmdutil.Factory{
		AppVersion: version.Version,
		IOStreams:   ios,
		Config: func() (config.Config, error) {
			return config.ReadConfig()
		},
		HTTPClient: func() (*khhttp.Client, error) {
			hosts, err := config.ReadHosts()
			if err != nil {
				return nil, err
			}

			// Priority: --host flag > KH_HOST env > hosts.yml default > built-in default
			var flagHost string
			if rootCmd != nil {
				if f := rootCmd.PersistentFlags().Lookup("host"); f != nil {
					flagHost = f.Value.String()
				}
			}
			envHost := os.Getenv("KH_HOST")
			activeHost := hosts.ActiveHost(flagHost, envHost)

			entry, _ := hosts.HostEntry(activeHost)

			return khhttp.NewClient(khhttp.ClientOptions{
				Host:       activeHost,
				Token:      entry.Token,
				Headers:    entry.Headers,
				IOStreams:   ios,
				AppVersion: version.Version,
			}), nil
		},
	}

	rootCmd = cmd.NewRootCmd(f)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(ios.ErrOut, "X %s\n", err.Error())
		os.Exit(1)
	}
}
