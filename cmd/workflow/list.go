package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Workflow is the API response shape for a single workflow in the list.
type Workflow struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	Visibility string `json:"visibility"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

func workflowStatus(enabled bool) string {
	if enabled {
		return "active"
	}
	return "paused"
}

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List workflows",
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

			host := cfg.DefaultHost
			url := khhttp.BuildBaseURL(host) + "/api/workflows?limit=" + strconv.Itoa(limit)

			req, err := client.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("executing request: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return cmdutil.NotFoundError{Err: errors.New("workflows not found")}
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				return cmdutil.RateLimitError{Err: errors.New("rate limit exceeded")}
			}
			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var workflows []Workflow
			if err := json.NewDecoder(resp.Body).Decode(&workflows); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(workflows, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"ID", "NAME", "STATUS", "VISIBILITY", "UPDATED"})
				for _, wf := range workflows {
					tw.AppendRow(table.Row{wf.ID, wf.Name, workflowStatus(wf.Enabled), wf.Visibility, wf.UpdatedAt})
				}
				tw.Render()
			})
		},
	}

	cmd.Flags().Int("limit", 30, "Maximum number of workflows to list")

	return cmd
}
