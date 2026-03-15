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

type createRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Nodes       []interface{} `json:"nodes,omitempty"`
	Edges       []interface{} `json:"edges,omitempty"`
}

type createResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

func NewCreateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a workflow",
		Aliases: []string{"new"},
		Args:    cobra.NoArgs,
		Example: `  # Create an empty workflow
  kh wf create --name "My Workflow"

  # Create with nodes from a JSON file
  kh wf create --name "DeFi Monitor" --nodes-file workflow.json

  # Create with inline JSON nodes
  kh wf create --name "Test" --nodes '[{"id":"t1","type":"trigger","position":{"x":0,"y":0},"data":{"type":"trigger","config":{"triggerType":"Manual"}}}]'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			if name == "" {
				return cmdutil.FlagError{Err: fmt.Errorf("--name is required")}
			}

			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}

			nodesFile, err := cmd.Flags().GetString("nodes-file")
			if err != nil {
				return err
			}

			nodesInline, err := cmd.Flags().GetString("nodes")
			if err != nil {
				return err
			}

			edgesInline, err := cmd.Flags().GetString("edges")
			if err != nil {
				return err
			}

			body := createRequest{
				Name:        name,
				Description: description,
			}

			// Load nodes/edges from file if provided
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

			// Inline nodes override file
			if nodesInline != "" {
				var nodes []interface{}
				if unmarshalErr := json.Unmarshal([]byte(nodesInline), &nodes); unmarshalErr != nil {
					return fmt.Errorf("parsing --nodes JSON: %w", unmarshalErr)
				}
				body.Nodes = nodes
			}

			// Inline edges override file
			if edgesInline != "" {
				var edges []interface{}
				if unmarshalErr := json.Unmarshal([]byte(edgesInline), &edges); unmarshalErr != nil {
					return fmt.Errorf("parsing --edges JSON: %w", unmarshalErr)
				}
				body.Edges = edges
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

			url := khhttp.BuildBaseURL(host) + "/api/workflows/create"
			req, err := client.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
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
				return fmt.Errorf("HTTP 401: unauthorized, run 'kh auth login' first")
			}
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var result createResponse
			if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
				return fmt.Errorf("decoding create response: %w", decodeErr)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(result, func(tw table.Writer) {
				fmt.Fprintf(f.IOStreams.Out, "Created workflow: %s (%s)\n", result.Name, result.ID)
				tw.Render()
			})
		},
	}

	cmd.Flags().String("name", "", "Workflow name (required)")
	cmd.Flags().String("description", "", "Workflow description")
	cmd.Flags().String("nodes-file", "", "Path to JSON file with nodes and edges")
	cmd.Flags().String("nodes", "", "Inline JSON array of nodes")
	cmd.Flags().String("edges", "", "Inline JSON array of edges")

	return cmd
}
