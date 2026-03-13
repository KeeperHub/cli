package org

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Member is the Better Auth member shape returned by list-members.
type Member struct {
	ID        string                 `json:"id"`
	Email     string                 `json:"email"`
	Role      string                 `json:"role"`
	CreatedAt string                 `json:"createdAt"`
	User      map[string]interface{} `json:"user"`
}

// memberName extracts a display name from the member, falling back to email.
func memberName(m Member) string {
	if m.User != nil {
		if name, ok := m.User["name"].(string); ok && name != "" {
			return name
		}
	}
	return ""
}

// membersResponse wraps the Better Auth list-members response.
type membersResponse struct {
	Members []Member `json:"members"`
}

func NewMembersCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "members",
		Short:   "List organization members",
		Aliases: []string{"m"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			url := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/auth/organization/list-members"

			bodyBytes, err := json.Marshal(map[string]string{})
			if err != nil {
				return fmt.Errorf("encoding request body: %w", err)
			}

			req, err := client.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("executing request: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var wrapper membersResponse
			if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
				return fmt.Errorf("unexpected response from member list endpoint: %w", err)
			}

			members := wrapper.Members

			p := output.NewPrinter(f.IOStreams, cmd)
			if len(members) == 0 && !p.IsJSON() {
				fmt.Fprintln(f.IOStreams.Out, "No members found.")
				return nil
			}

			return p.PrintData(members, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"NAME", "EMAIL", "ROLE"})
				for _, m := range members {
					name := memberName(m)
					tw.AppendRow(table.Row{name, m.Email, m.Role})
				}
				tw.Render()
			})
		},
	}

	return cmd
}
