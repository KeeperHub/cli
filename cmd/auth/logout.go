package auth

import (
	"fmt"
	"os"

	internalauth "github.com/keeperhub/cli/internal/auth"
	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// DeleteTokenFunc is the function used to remove the keyring token.
// Tests may override this to avoid touching the real keyring.
var DeleteTokenFunc = func(host string) error {
	return internalauth.DeleteToken(host)
}

// ClearHostTokenFunc is the function used to clear the token from hosts.yml.
// Tests may override this to avoid touching the real config file.
var ClearHostTokenFunc = func(host string) error {
	return config.ClearHostToken(host)
}

func NewLogoutCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out of KeeperHub",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			hosts, err := config.ReadHosts()
			if err != nil {
				return err
			}

			var flagHost string
			if root := cmd.Root(); root != nil {
				if fl := root.PersistentFlags().Lookup("host"); fl != nil {
					flagHost = fl.Value.String()
				}
			}
			envHost := os.Getenv("KH_HOST")
			host := hosts.ActiveHost(flagHost, envHost)

			if err := DeleteTokenFunc(host); err != nil {
				return fmt.Errorf("removing token from keyring: %w", err)
			}

			if err := ClearHostTokenFunc(host); err != nil {
				return fmt.Errorf("clearing token from config: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "Logged out of %s\n", host)
			return nil
		},
	}

	return cmd
}
