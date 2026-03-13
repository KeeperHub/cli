package workflow

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

type goLiveRequest struct {
	Name         string   `json:"name"`
	PublicTagIDs []string `json:"publicTagIds"`
}

func NewGoLiveCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "go-live <workflow-id>",
		Short:   "Publish a workflow",
		Aliases: []string{"live"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Publish a workflow as a template
  kh wf go-live abc123 --name "My DeFi Template"

  # Publish with public tags
  kh wf go-live abc123 --name "Uniswap Swap" --tags tag1,tag2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID := args[0]

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			if name == "" {
				return cmdutil.FlagError{Err: fmt.Errorf("--name is required")}
			}

			tags, err := cmd.Flags().GetStringSlice("tags")
			if err != nil {
				return err
			}
			if tags == nil {
				tags = []string{}
			}

			client, err := f.HTTPClient()
			if err != nil {
				return err
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			host := cmdutil.ResolveHost(cmd, cfg)

			body := goLiveRequest{
				Name:         name,
				PublicTagIDs: tags,
			}
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return err
			}

			url := khhttp.BuildBaseURL(host) + "/api/workflows/" + workflowID + "/go-live"
			req, err := client.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				return fmt.Errorf("HTTP 401: unauthorized, this command requires interactive login, run 'kh auth login' first")
			}
			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decoding go-live response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(result, func(tw table.Writer) {
				fmt.Fprintf(f.IOStreams.Out, "Workflow %s is now live\n", workflowID)
				tw.Render()
			})
		},
	}

	cmd.Flags().String("name", "", "Name for the published workflow (required)")
	cmd.Flags().StringSlice("tags", nil, "Public tag IDs to attach")

	return cmd
}
