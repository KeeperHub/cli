package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type updateRequest struct {
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Nodes       []interface{} `json:"nodes,omitempty"`
	Edges       []interface{} `json:"edges,omitempty"`
}

func NewUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <workflow-id>",
		Short: "Update a workflow",
		Args:  cobra.ExactArgs(1),
		Example: `  # Update workflow name
  kh wf update abc123 --name "New Name"

  # Update nodes from file
  kh wf update abc123 --nodes-file workflow.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID := args[0]

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}
			nodesFile, err := cmd.Flags().GetString("nodes-file")
			if err != nil {
				return err
			}

			body := updateRequest{}

			if name != "" {
				body.Name = name
			}
			if description != "" {
				body.Description = description
			}

			if nodesFile != "" {
				fileData, readErr := os.ReadFile(nodesFile)
				if readErr != nil {
					return fmt.Errorf("reading nodes file: %w", readErr)
				}

				var fileContent struct {
					Nodes []interface{} `json:"nodes"`
					Edges []interface{} `json:"edges"`
				}
				if unmarshalErr := json.Unmarshal(fileData, &fileContent); unmarshalErr != nil {
					return fmt.Errorf("parsing nodes file: %w", unmarshalErr)
				}
				body.Nodes = fileContent.Nodes
				body.Edges = fileContent.Edges
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

			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return err
			}

			url := khhttp.BuildBaseURL(host) + "/api/workflows/" + workflowID
			req, err := client.NewRequest(http.MethodPatch, url, bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return cmdutil.NotFoundError{Err: fmt.Errorf("workflow %q not found", workflowID)}
			}
			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var result map[string]interface{}
			if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
				return fmt.Errorf("decoding update response: %w", decodeErr)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(result, func(tw table.Writer) {
				fmt.Fprintf(f.IOStreams.Out, "Workflow %s updated\n", workflowID)
				tw.Render()
			})
		},
	}

	cmd.Flags().String("name", "", "New workflow name")
	cmd.Flags().String("description", "", "New workflow description")
	cmd.Flags().String("nodes-file", "", "Path to JSON file with nodes and edges")

	return cmd
}
