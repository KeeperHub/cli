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

// apiKeyName mirrors the label the server attaches to keys minted by the
// device-authorization flow, so the CLI can name the credential it just
// created without a second round trip.
const apiKeyName = "kh CLI"

// DeviceLoginFunc is the function used to perform device-code login.
// Tests may override this to avoid real network calls.
var DeviceLoginFunc = func(host string, ios *iostreams.IOStreams) (string, error) {
	return internalauth.DeviceLogin(host, ios)
}

// SetTokenFunc is the function used to store a token in hosts.yml.
// Tests may override this to avoid touching the real config.
var SetTokenFunc = func(host, token string) error {
	return config.SetHostToken(host, token)
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
Use --with-token to read an API key from stdin for non-interactive automation.

See also: kh auth status, kh auth logout`,
		Example: `  # Log in (device code flow)
  kh auth login

  # Log in with an API key (non-interactive)
  echo "kh_xxx" | kh auth login --with-token`,
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

			withToken, _ := cmd.Flags().GetBool("with-token")
			force, _ := cmd.Flags().GetBool("force")

			// Completing the device flow mints a new organization API key
			// server-side, and the plaintext key is returned exactly once.
			// Re-running login on an already-authenticated host would strand
			// the stored key and leave a new one behind on every invocation,
			// so stop early unless the caller asked for a fresh credential.
			if !withToken && !force {
				if entry, ok := hosts.HostEntry(host); ok && entry.Token != "" {
					if info, infoErr := FetchTokenInfoFunc(host, entry.Token); infoErr == nil {
						fmt.Fprintf(f.IOStreams.Out,
							"Already logged in to %s as %s\nUse --force to replace the stored API key.\n",
							host, info.Email)
						return nil
					}
				}
			}

			var token string

			if withToken {
				t, readErr := internalauth.ReadTokenFromStdin(f.IOStreams)
				if readErr != nil {
					return readErr
				}
				if err := SetTokenFunc(host, t); err != nil {
					return fmt.Errorf("storing token: %w", err)
				}
				token = t
			} else {
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
				fmt.Fprintf(f.IOStreams.Out, "Logged in to %s\n", host)
			} else {
				fmt.Fprintf(f.IOStreams.Out, "Logged in to %s as %s\n", host, info.Email)
			}

			// The device flow creates a real organization API key rather than a
			// browser session. Say so plainly: it is a durable credential the
			// user can see and revoke in the dashboard, not an invisible login.
			if !withToken {
				fmt.Fprintf(f.IOStreams.Out,
					"Created a new organization API key (%q) and stored it in %s\nRevoke it from the KeeperHub dashboard under API keys.\n",
					apiKeyName, config.HostsFile())
			}
			return nil
		},
	}

	cmd.Flags().Bool("with-token", false, "Read token from stdin")
	cmd.Flags().Bool(
		"force",
		false,
		"Log in again even if a valid credential is stored, creating a new API key",
	)

	return cmd
}
