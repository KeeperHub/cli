package project

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Project is the API response shape for a single project in the list.
type Project struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Color         string `json:"color"`
	WorkflowCount int    `json:"workflowCount"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List projects",
		Aliases: []string{"ls"},
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

			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return err
			}

			url := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/projects?limit=" + strconv.Itoa(limit)

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

			p := output.NewPrinter(f.IOStreams, cmd)
			if len(projects) == 0 && !p.IsJSON() {
				fmt.Fprintln(f.IOStreams.Out, "No projects found. Create one with: kh p create 'name'")
				return nil
			}

			return p.PrintData(projects, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"ID", "NAME", "DESCRIPTION", "WORKFLOWS"})
				for _, proj := range projects {
					desc := proj.Description
					if len(desc) > 40 {
						desc = desc[:37] + "..."
					}
					tw.AppendRow(table.Row{proj.ID, proj.Name, desc, proj.WorkflowCount})
				}
				tw.Render()
			})
		},
	}

	cmd.Flags().Int("limit", 30, "Maximum number of projects to list")

	return cmd
}
