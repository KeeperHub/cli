package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			workflowID := args[0]
			host := cmdutil.ResolveHost(cmd, cfg)

			webMode, _ := cmd.Flags().GetBool("web")
			if webMode {
				webURL := khhttp.BuildBaseURL(host) + "/workflows/" + workflowID
				fmt.Fprintln(f.IOStreams.Out, webURL)
				return cmdutil.OpenBrowser(webURL)
			}

			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

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
			if p.IsJSON() {
				return p.PrintJSON(detail)
			}
			p.PrintKeyValue([][2]string{
				{"ID", detail.ID},
				{"Name", detail.Name},
				{"Status", workflowStatus(detail.Enabled)},
				{"Visibility", detail.Visibility},
				{"Created", output.TimeAgo(detail.CreatedAt)},
				{"Updated", output.TimeAgo(detail.UpdatedAt)},
				{"Nodes", fmt.Sprintf("%d", len(detail.Nodes))},
				{"Edges", fmt.Sprintf("%d", len(detail.Edges))},
			})
			return nil
		},
	}

	cmd.Flags().Bool("web", false, "Open the workflow in the browser")

	return cmd
}
