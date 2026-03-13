package org

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Organization is the API response shape for a single organization.
type Organization struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Role      string                 `json:"role"`
	CreatedAt string                 `json:"createdAt"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List organizations",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Example: `  # List all organizations
  kh o ls

  # List as JSON
  kh o ls --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			url := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/organizations"

			req, err := client.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("executing request: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var orgs []Organization
			if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			if len(orgs) == 0 && !p.IsJSON() {
				fmt.Fprintln(f.IOStreams.Out, "No organizations found.")
				return nil
			}

			return p.PrintData(orgs, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"NAME", "SLUG", "ROLE", "CREATED"})
				for _, o := range orgs {
					tw.AppendRow(table.Row{o.Name, o.Slug, o.Role, o.CreatedAt})
				}
				tw.Render()
			})
		},
	}

	return cmd
}
