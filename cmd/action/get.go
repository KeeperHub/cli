package action

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

func NewGetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <action-name>",
		Short:   "Get an action",
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Get action by name
  kh a g ethereum-transfer

  # Get action details as JSON
  kh a g uniswap-swap --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			query := args[0]
			host := cfg.DefaultHost
			apiURL := khhttp.BuildBaseURL(host) + "/api/integrations"

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

			var found *Integration
			for i := range integrations {
				item := &integrations[i]
				if item.ID == query || strings.EqualFold(item.Name, query) {
					found = item
					break
				}
			}

			if found == nil {
				return cmdutil.NotFoundError{Err: fmt.Errorf("action %q not found", query)}
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(found, func(tw table.Writer) {
				tw.AppendRow(table.Row{"Name", found.Name})
				tw.AppendRow(table.Row{"Type", found.Type})
				tw.AppendRow(table.Row{"Managed", managedLabel(found.IsManaged)})
				tw.AppendRow(table.Row{"Created", found.CreatedAt})
				tw.AppendRow(table.Row{"Updated", found.UpdatedAt})
				tw.Render()
			})
		},
	}

	return cmd
}

