package tag

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
		Use:     "get <tag-id>",
		Short:   "Get a tag",
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Get tag details
  kh t g abc123

  # Get as JSON
  kh t g abc123 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tagID := args[0]

			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			url := khhttp.BuildBaseURL(cmdutil.ResolveHost(cmd, cfg)) + "/api/tags"
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

			var found *Tag
			for i := range tags {
				if tags[i].ID == tagID {
					found = &tags[i]
					break
				}
			}

			if found == nil {
				return cmdutil.NotFoundError{Err: fmt.Errorf("tag %q not found", tagID)}
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(*found, func(tw table.Writer) {
				tw.AppendRow(table.Row{"ID", found.ID})
				tw.AppendRow(table.Row{"Name", found.Name})
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
