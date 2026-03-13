package auth

import (
	"errors"
	"fmt"
	"os"

	internalauth "github.com/keeperhub/cli/internal/auth"
	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

// BrowserLoginFunc is the function used to perform browser-based login.
// Tests may override this to avoid opening a real browser.
var BrowserLoginFunc = func(host string, ios *iostreams.IOStreams) (string, error) {
	return internalauth.BrowserLogin(host, ios)
}

// DeviceLoginFunc is the function used to perform device-code login.
// Tests may override this to avoid real network calls.
var DeviceLoginFunc = func(host string, ios *iostreams.IOStreams) (string, error) {
	return internalauth.DeviceLogin(host, ios)
}

// SetTokenFunc is the function used to store a token in the keyring.
// Tests may override this to avoid touching the real keyring.
var SetTokenFunc = func(host, token string) error {
	return internalauth.SetToken(host, token)
}

// FetchTokenInfoFunc is the function used to fetch session details from the server.
// Tests may override this to avoid real HTTP calls.
var FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
	return internalauth.FetchTokenInfo(host, token)
}

func NewLoginCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to KeeperHub",
		Args:  cobra.NoArgs,
		Long: `Authenticate with KeeperHub using the device code flow.
Opens a browser to confirm a one-time code.
Use --browser for direct browser-based OAuth (useful if device flow is unavailable).
Use --with-token to read an API key from stdin for non-interactive automation.

See also: kh auth status, kh auth logout`,
		Example: `  # Log in (device code flow)
  kh auth login

  # Log in via direct browser OAuth
  kh auth login --browser`,
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

			useBrowser, _ := cmd.Flags().GetBool("browser")
			withToken, _ := cmd.Flags().GetBool("with-token")

			var token string

			switch {
			case withToken:
				t, readErr := internalauth.ReadTokenFromStdin(f.IOStreams)
				if readErr != nil {
					return readErr
				}
				if err := SetTokenFunc(host, t); err != nil {
					return fmt.Errorf("storing token: %w", err)
				}
				token = t

			case useBrowser:
				t, loginErr := BrowserLoginFunc(host, f.IOStreams)
				if loginErr != nil {
					return loginErr
				}
				token = t

			default:
				t, loginErr := DeviceLoginFunc(host, f.IOStreams)
				if loginErr != nil {
					return loginErr
				}
				token = t
			}

			if token == "" {
				return errors.New("no token received")
			}

			info, err := FetchTokenInfoFunc(host, token)
			if err != nil {
				// Non-fatal: login succeeded but we can't fetch user details
				fmt.Fprintf(f.IOStreams.Out, "Logged in to %s\n", host)
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "Logged in to %s as %s\n", host, info.Email)
			return nil
		},
	}

	cmd.Flags().Bool("browser", false, "Use direct browser OAuth instead of device code flow")
	cmd.Flags().Bool("with-token", false, "Read token from stdin")

	return cmd
}
