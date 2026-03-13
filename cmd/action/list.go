package action

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

// Integration is the API response shape for a single integration/action.
type Integration struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsManaged bool   `json:"isManaged"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func managedLabel(isManaged bool) string {
	if isManaged {
		return "yes"
	}
	return "no"
}

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List available actions",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Example: `  # List all actions
  kh a ls

  # Filter by category
  kh a ls --category web3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			category, err := cmd.Flags().GetString("category")
			if err != nil {
				return err
			}

			host := cfg.DefaultHost
			apiURL := khhttp.BuildBaseURL(host) + "/api/integrations"
			if category != "" {
				apiURL += "?category=" + category
			}

			req, err := client.NewRequest(http.MethodGet, apiURL, nil)
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

			var integrations []Integration
			if err := json.NewDecoder(resp.Body).Decode(&integrations); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(integrations, func(tw table.Writer) {
				if len(integrations) == 0 {
					fmt.Fprintln(f.IOStreams.Out, "No actions found.")
					return
				}
				tw.AppendHeader(table.Row{"NAME", "TYPE", "MANAGED"})
				for _, item := range integrations {
					tw.AppendRow(table.Row{item.Name, item.Type, managedLabel(item.IsManaged)})
				}
				tw.Render()
			})
		},
	}

	cmd.Flags().String("category", "", "Filter by category")

	return cmd
}
