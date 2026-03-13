package tag

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

// Tag is the API response shape for a single tag in the list.
type Tag struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Color         string `json:"color"`
	WorkflowCount int    `json:"workflowCount"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tags",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Example: `  # List all tags
  kh t ls

  # List with a higher limit
  kh t ls --limit 50`,
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

			url := khhttp.BuildBaseURL(cmdutil.ResolveHost(cmd, cfg)) + "/api/tags?limit=" + strconv.Itoa(limit)

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

			var tags []Tag
			if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			if len(tags) == 0 && !p.IsJSON() {
				fmt.Fprintln(f.IOStreams.Out, "No tags found. Create one with: kh t create 'name'")
				return nil
			}

			return p.PrintData(tags, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"ID", "NAME", "COLOR", "WORKFLOWS"})
				for _, t := range tags {
					tw.AppendRow(table.Row{t.ID, t.Name, t.Color, t.WorkflowCount})
				}
				tw.Render()
			})
		},
	}

	cmd.Flags().Int("limit", 30, "Maximum number of tags to list")

	return cmd
}
