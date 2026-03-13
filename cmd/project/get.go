package project

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

func NewGetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <project-id>",
		Short:   "Get a project",
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Get project details
  kh p g abc123

  # Get as JSON
  kh p g abc123 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]

			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			url := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/projects"
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

			var projects []Project
			if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			var found *Project
			for i := range projects {
				if projects[i].ID == projectID {
					found = &projects[i]
					break
				}
			}

			if found == nil {
				return cmdutil.NotFoundError{Err: fmt.Errorf("project %q not found", projectID)}
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(*found, func(tw table.Writer) {
				tw.AppendRow(table.Row{"ID", found.ID})
				tw.AppendRow(table.Row{"Name", found.Name})
				tw.AppendRow(table.Row{"Description", found.Description})
				tw.AppendRow(table.Row{"Color", found.Color})
				tw.AppendRow(table.Row{"Workflows", found.WorkflowCount})
				tw.AppendRow(table.Row{"Created", found.CreatedAt})
				tw.AppendRow(table.Row{"Updated", found.UpdatedAt})
				tw.Render()
			})
		},
	}

	return cmd
}
