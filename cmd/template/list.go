package template

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	var query string

	cmd := &cobra.Command{
		Use:     "list [query]",
		Short:   "List workflow templates",
		Aliases: []string{"ls", "search"},
		Args:    cobra.MaximumNArgs(1),
		Example: `  # List featured templates
  kh tp ls

  # Search templates by keyword
  kh tp ls defi
  kh tp ls --query monitor

  # List templates as JSON
  kh tp ls --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Positional arg takes precedence over --query flag
			if len(args) > 0 {
				query = args[0]
			}

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

			if query != "" {
				templates = filterTemplates(templates, query)
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

	cmd.Flags().StringVarP(&query, "query", "q", "", "Filter templates by name or description")

	return cmd
}

// filterTemplates returns templates whose name or description contains the query
// (case-insensitive substring match).
func filterTemplates(templates []Template, query string) []Template {
	q := strings.ToLower(query)
	var result []Template
	for _, tpl := range templates {
		if strings.Contains(strings.ToLower(tpl.Name), q) ||
			strings.Contains(strings.ToLower(tpl.Description), q) {
			result = append(result, tpl)
		}
	}
	return result
}
