package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	internalauth "github.com/keeperhub/cli/internal/auth"
	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// ResolveTokenFunc is the function used to resolve the auth token.
// Tests may override this to avoid real keyring/config access.
var ResolveTokenFunc = func(host string) (internalauth.ResolvedToken, error) {
	return internalauth.ResolveToken(host)
}

type statusOutput struct {
	User           string `json:"user"`
	Email          string `json:"email"`
	Org            string `json:"organization"`
	OrgID          string `json:"organization_id"`
	Role           string `json:"role"`
	ExpiresAt      string `json:"expires_at"`
	Method         string `json:"method"`
	Host           string `json:"host"`
}

func NewStatusCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Args:  cobra.NoArgs,
		Example: `  # Show current auth status
  kh auth status

  # Show status as JSON
  kh auth status --json`,
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

			resolved, err := ResolveTokenFunc(host)
			if err != nil {
				return err
			}

			if resolved.Method == internalauth.AuthMethodNone {
				return errors.New("Not authenticated. Run 'kh auth login' to sign in.")
			}

			info, err := FetchTokenInfoFunc(host, resolved.Token)
			if err != nil {
				return fmt.Errorf("fetching session info: %w", err)
			}

			expiresAt := ""
			if !info.ExpiresAt.IsZero() {
				expiresAt = info.ExpiresAt.Format("2006-01-02 15:04:05")
			}

			data := statusOutput{
				User:      info.Name,
				Email:     info.Email,
				Org:       info.OrgName,
				OrgID:     info.OrgID,
				Role:      info.Role,
				ExpiresAt: expiresAt,
				Method:    string(resolved.Method),
				Host:      host,
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(data, func(tw table.Writer) {
				tw.AppendRow(table.Row{"Host", host})
				tw.AppendRow(table.Row{"User", info.Email})
				tw.AppendRow(table.Row{"Organization", info.OrgName})
				tw.AppendRow(table.Row{"Role", info.Role})
				if expiresAt != "" {
					tw.AppendRow(table.Row{"Expires", expiresAt})
				}
				tw.AppendRow(table.Row{"Auth Method", string(resolved.Method)})
				tw.Render()
			})
		},
	}

	return cmd
}
