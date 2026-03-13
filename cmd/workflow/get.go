package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// WorkflowDetail extends Workflow with node/edge counts and tags.
type WorkflowDetail struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Enabled    bool          `json:"enabled"`
	Visibility string        `json:"visibility"`
	CreatedAt  string        `json:"createdAt"`
	UpdatedAt  string        `json:"updatedAt"`
	Nodes      []interface{} `json:"nodes"`
	Edges      []interface{} `json:"edges"`
	PublicTags []string      `json:"publicTags,omitempty"`
}

func NewGetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <workflow-id>",
		Short:   "Get a workflow",
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Get workflow details
  kh wf g abc123

  # Get as JSON
  kh wf g abc123 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			workflowID := args[0]
			host := cfg.DefaultHost
			url := khhttp.BuildBaseURL(host) + "/api/workflows/" + workflowID

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
				return cmdutil.NotFoundError{Err: fmt.Errorf("workflow %q not found", workflowID)}
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				return cmdutil.RateLimitError{Err: errors.New("rate limit exceeded")}
			}
			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var detail WorkflowDetail
			if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(detail, func(tw table.Writer) {
				tw.AppendRow(table.Row{"ID", detail.ID})
				tw.AppendRow(table.Row{"Name", detail.Name})
				tw.AppendRow(table.Row{"Status", workflowStatus(detail.Enabled)})
				tw.AppendRow(table.Row{"Visibility", detail.Visibility})
				tw.AppendRow(table.Row{"Created", detail.CreatedAt})
				tw.AppendRow(table.Row{"Updated", detail.UpdatedAt})
				tw.AppendRow(table.Row{"Nodes", len(detail.Nodes)})
				tw.AppendRow(table.Row{"Edges", len(detail.Edges)})
				tw.Render()
			})
		},
	}

	return cmd
}
