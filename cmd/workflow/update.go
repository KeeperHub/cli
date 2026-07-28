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

// nullableID returns nil for an empty id so the JSON body carries an
// explicit null (unassign); otherwise it returns the id unchanged.
func nullableID(id string) interface{} {
	if id == "" {
		return nil
	}
	return id
}

func NewUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <workflow-id>",
		Short: "Update a workflow",
		Args:  cobra.ExactArgs(1),
		Long: `Update a workflow.

Nodes and edges can only be supplied as a file. Unlike 'kh workflow create',
there are no inline --nodes / --edges flags here.

The file is a JSON OBJECT holding both keys, not a bare array of nodes:

  {"nodes": [...], "edges": [...]}

Both keys are sent together, so --nodes-file replaces the whole graph. To
change one node, fetch the current definition first:

  kh workflow get <id> --json > workflow.json

Node config is only lightly checked on the way in. An integrationId that
matches no integration is accepted (the API treats unknown ids as stale-but-
savable references), and "network" and "actionType" are not validated at all.
A successful update is not evidence that the workflow runs - misconfiguration
surfaces at execution time.`,
		Example: `  # Update workflow name
  kh wf update abc123 --name "New Name"

  # Update nodes from file - the file is an object, not an array:
  #   {"nodes": [ ... ], "edges": [ ... ]}
  kh wf update abc123 --nodes-file workflow.json

  # Assign the workflow to a project and tag
  kh wf update abc123 --project proj_123 --tag tag_456

  # Remove the workflow from its project (pass an empty value)
  kh wf update abc123 --project ""`,
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

			// A map lets a changed --project/--tag flag send an explicit
			// null (unassign) that a struct with omitempty could not express.
			body := map[string]interface{}{}

			if name != "" {
				body["name"] = name
			}
			if description != "" {
				body["description"] = description
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
				body["nodes"] = fileContent.Nodes
				body["edges"] = fileContent.Edges
			}

			if cmd.Flags().Changed("project") {
				project, projectErr := cmd.Flags().GetString("project")
				if projectErr != nil {
					return projectErr
				}
				body["projectId"] = nullableID(project)
			}

			if cmd.Flags().Changed("tag") {
				tag, tagErr := cmd.Flags().GetString("tag")
				if tagErr != nil {
					return tagErr
				}
				body["tagId"] = nullableID(tag)
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
	cmd.Flags().String("nodes-file", "", `Path to a JSON file shaped {"nodes": [...], "edges": [...]}; replaces the whole graph`)
	cmd.Flags().String("project", "", "Project ID to assign (empty value unassigns)")
	cmd.Flags().String("tag", "", "Tag ID to assign (empty value unassigns)")

	return cmd
}
