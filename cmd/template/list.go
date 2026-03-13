package template

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

// PublicTag is a tag attached to a public workflow template.
type PublicTag struct {
	Name string `json:"name"`
}

// Template is the API response shape for a featured public workflow acting as a template.
type Template struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Visibility  string      `json:"visibility"`
	PublicTags  []PublicTag `json:"publicTags"`
	CreatedAt   string      `json:"createdAt"`
}

var truncateMax = 50

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func categoryFromTags(tags []PublicTag) string {
	if len(tags) > 0 {
		return tags[0].Name
	}
	return "General"
}

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List workflow templates",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Example: `  # List featured templates
  kh tp ls

  # List templates as JSON
  kh tp ls --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			host := cmdutil.ResolveHost(cmd, cfg)
			url := khhttp.BuildBaseURL(host) + "/api/workflows/public?featured=true"

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

			var templates []Template
			if err := json.NewDecoder(resp.Body).Decode(&templates); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			if len(templates) == 0 && !p.IsJSON() {
				fmt.Fprintln(f.IOStreams.Out, "No templates found.")
				return nil
			}
			return p.PrintData(templates, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"NAME", "DESCRIPTION", "CATEGORY"})
				for _, tpl := range templates {
					tw.AppendRow(table.Row{
						tpl.Name,
						truncate(tpl.Description, truncateMax),
						categoryFromTags(tpl.PublicTags),
					})
				}
				tw.Render()
			})
		},
	}

	return cmd
}
