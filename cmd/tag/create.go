package tag

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
		Short:   "Create a tag",
		Aliases: []string{"c"},
		Args:    cobra.MinimumNArgs(1),
		Example: `  # Create a tag with default color
  kh t create "defi"

  # Create a tag with a custom color
  kh t create "urgent" --color "#ef4444"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			color, _ := cmd.Flags().GetString("color")

			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			bodyBytes, err := json.Marshal(map[string]string{
				"name":  name,
				"color": color,
			})
			if err != nil {
				return fmt.Errorf("building request body: %w", err)
			}

			url := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/tags"
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

			var created Tag
			if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(created, func(tw table.Writer) {
				tw.AppendRow(table.Row{"ID", created.ID})
				tw.AppendRow(table.Row{"Name", created.Name})
				tw.AppendRow(table.Row{"Color", created.Color})
				tw.AppendRow(table.Row{"Workflows", created.WorkflowCount})
				tw.AppendRow(table.Row{"Created", created.CreatedAt})
				tw.Render()
			})
		},
	}

	cmd.Flags().String("color", "#6366f1", "Tag color (default: #6366f1)")

	return cmd
}
