package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// duplicateResponse is the API response shape for a duplicated workflow.
type duplicateResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewDeployCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deploy <template-id>",
		Short:   "Deploy a workflow template",
		Aliases: []string{"d"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Deploy a template using its ID
  kh tp deploy abc123

  # Deploy and give it a custom name
  kh tp deploy abc123 --name "My Uniswap Workflow"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			host := cfg.DefaultHost
			url := khhttp.BuildBaseURL(host) + "/api/workflows/" + templateID + "/duplicate"

			var bodyData map[string]interface{}
			if name != "" {
				bodyData = map[string]interface{}{"name": name}
			} else {
				bodyData = map[string]interface{}{}
			}

			bodyBytes, err := json.Marshal(bodyData)
			if err != nil {
				return fmt.Errorf("encoding request body: %w", err)
			}

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

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				return khhttp.NewAPIError(resp)
			}

			var result duplicateResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			if p.IsJSON() {
				return p.PrintJSON(result)
			}

			fmt.Fprintf(f.IOStreams.Out, "Created workflow '%s' (%s)\n", result.Name, result.ID)
			return nil
		},
	}

	cmd.Flags().String("name", "", "Workflow name")

	return cmd
}
