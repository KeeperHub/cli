package main

import (
	"fmt"
	"os"

	"github.com/keeperhub/cli/cmd"
	"github.com/keeperhub/cli/internal/auth"
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

	// resolveActiveHost returns the effective host using the priority chain:
	// --host flag > KH_HOST env > hosts.yml default > config.yml > built-in default.
	resolveActiveHost := func() string {
		var flagHost string
		if rootCmd != nil {
			if f := rootCmd.PersistentFlags().Lookup("host"); f != nil {
				flagHost = f.Value.String()
			}
		}
		envHost := os.Getenv("KH_HOST")
		hosts, err := config.ReadHosts()
		if err != nil {
			if flagHost != "" {
				return flagHost
			}
			if envHost != "" {
				return envHost
			}
			return "app.keeperhub.com"
		}
		return hosts.ActiveHost(flagHost, envHost)
	}

	// resolveOrgFlag returns the --org flag value if set, or empty string.
	resolveOrgFlag := func() string {
		if rootCmd != nil {
			if f := rootCmd.PersistentFlags().Lookup("org"); f != nil {
				return f.Value.String()
			}
		}
		return ""
	}

	f := &cmdutil.Factory{
		AppVersion: version.Version,
		IOStreams:   ios,
		OrgID:      resolveOrgFlag,
		Config: func() (config.Config, error) {
			return config.ReadConfig()
		},
		BaseURL: func() string {
			return khhttp.BuildBaseURL(resolveActiveHost())
		},
		HTTPClient: func() (*khhttp.Client, error) {
			activeHost := resolveActiveHost()

			hosts, err := config.ReadHosts()
			if err != nil {
				return nil, err
			}

			// Resolve token using the auth chain: KH_API_KEY > keyring > hosts.yml
			resolved, err := auth.ResolveToken(activeHost)
			if err != nil {
				return nil, err
			}

			entry, _ := hosts.HostEntry(activeHost)

			return khhttp.NewClient(khhttp.ClientOptions{
				Host:        activeHost,
				Token:       resolved.Token,
				Headers:     entry.Headers,
				OrgOverride: resolveOrgFlag(),
				IOStreams:    ios,
				AppVersion:  version.Version,
			}), nil
		},
	}

	rootCmd = cmd.NewRootCmd(f)

	if err := rootCmd.Execute(); err != nil {
		exitCode := cmdutil.ExitCodeForError(err)
		if exitCode != 0 {
			fmt.Fprintf(ios.ErrOut, "X %s\n", err.Error())
		}
		os.Exit(exitCode)
	}
}
