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
	Nodes       []interface{} `json:"nodes"`
	Edges       []interface{} `json:"edges"`
	ProjectID   string        `json:"projectId,omitempty"`
	TagID       string        `json:"tagId,omitempty"`
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
		Long: `Create a workflow.

Nodes and edges can be supplied two ways, and the two take different shapes:

  --nodes-file FILE   a JSON OBJECT: {"nodes": [...], "edges": [...]}
  --nodes / --edges   bare JSON ARRAYS, passed separately

Inline flags override the file when both are given. 'kh workflow update' accepts
only --nodes-file, so a file is the portable choice if you intend to edit the
workflow later.

New workflows are created DISABLED. There is no --enabled flag; enable the
workflow in the web app, or PATCH "enabled": true against the API directly.

Node config is only lightly checked on the way in. An integrationId that
matches no integration is accepted (the API treats unknown ids as stale-but-
savable references), and "network" and "actionType" are not validated at all.
A successful create is not evidence that the workflow runs - misconfiguration
surfaces at execution time.`,
		Example: `  # Create an empty workflow
  kh wf create --name "My Workflow"

  # Create with nodes from a JSON file - the file is an object, not an array:
  #   {"nodes": [ ... ], "edges": [ ... ]}
  kh wf create --name "DeFi Monitor" --nodes-file workflow.json

  # Create with inline JSON nodes - these flags take bare arrays
  kh wf create --name "Test" --nodes '[{"id":"t1","type":"trigger","position":{"x":0,"y":0},"data":{"type":"trigger","config":{"triggerType":"Manual"}}}]'

  # Create inside a project and label it with a tag
  kh wf create --name "Payouts" --project proj_123 --tag tag_456`,
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

			project, err := cmd.Flags().GetString("project")
			if err != nil {
				return err
			}

			tag, err := cmd.Flags().GetString("tag")
			if err != nil {
				return err
			}

			body := createRequest{
				Name:        name,
				Description: description,
				Nodes:       []interface{}{},
				Edges:       []interface{}{},
				ProjectID:   project,
				TagID:       tag,
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
	cmd.Flags().String("nodes-file", "", `Path to a JSON file shaped {"nodes": [...], "edges": [...]}`)
	cmd.Flags().String("nodes", "", "Inline JSON array of nodes (overrides --nodes-file)")
	cmd.Flags().String("edges", "", "Inline JSON array of edges (overrides --nodes-file)")
	cmd.Flags().String("project", "", "Project ID to assign the workflow to")
	cmd.Flags().String("tag", "", "Tag ID to label the workflow")

	return cmd
}
