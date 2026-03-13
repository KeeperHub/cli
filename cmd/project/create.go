package project

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

func NewCreateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name>",
		Short:   "Create a project",
		Aliases: []string{"c"},
		Args:    cobra.MinimumNArgs(1),
		Example: `  # Create a project
  kh p create "My Project"

  # Create with a description
  kh p create "DeFi Automations" --description "Uniswap and Aave workflows"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			description, _ := cmd.Flags().GetString("description")

			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			bodyBytes, err := json.Marshal(map[string]string{
				"name":        name,
				"description": description,
			})
			if err != nil {
				return fmt.Errorf("building request body: %w", err)
			}

			url := khhttp.BuildBaseURL(cmdutil.ResolveHost(cmd, cfg)) + "/api/projects"
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

			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var created Project
			if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(created, func(tw table.Writer) {
				tw.AppendRow(table.Row{"ID", created.ID})
				tw.AppendRow(table.Row{"Name", created.Name})
				tw.AppendRow(table.Row{"Description", created.Description})
				tw.AppendRow(table.Row{"Workflows", created.WorkflowCount})
				tw.AppendRow(table.Row{"Created", created.CreatedAt})
				tw.Render()
			})
		},
	}

	cmd.Flags().String("description", "", "Project description")

	return cmd
}
