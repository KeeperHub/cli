package workflow

import (
	"bufio"
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

func NewDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <workflow-id>",
		Short:   "Delete a workflow",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Delete a workflow (will prompt for confirmation)
  kh wf delete abc123

  # Delete without prompting
  kh wf delete abc123 --yes

  # Force delete a workflow that has execution history
  kh wf delete abc123 --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID := args[0]

			force, _ := cmd.Flags().GetBool("force")
			yes, _ := cmd.Flags().GetBool("yes")

			if !yes && f.IOStreams.IsTerminal() {
				prompt := fmt.Sprintf("Delete workflow %s? This cannot be undone. (y/N) ", workflowID)
				if force {
					prompt = fmt.Sprintf("Force delete workflow %s and all its execution history? This cannot be undone. (y/N) ", workflowID)
				}
				fmt.Fprint(f.IOStreams.Out, prompt)
				scanner := bufio.NewScanner(f.IOStreams.In)
				if scanner.Scan() {
					answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
					if answer != "y" && answer != "yes" {
						return cmdutil.CancelError{Err: fmt.Errorf("delete cancelled")}
					}
				}
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

			url := khhttp.BuildBaseURL(host) + "/api/workflows/" + workflowID
			if force {
				url += "?force=true"
			}
			req, err := client.NewRequest(http.MethodDelete, url, nil)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("cannot delete workflow %s: %w", workflowID, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return cmdutil.NotFoundError{Err: fmt.Errorf("workflow %q not found", workflowID)}
			}
			if resp.StatusCode == http.StatusConflict && !force {
				fmt.Fprintf(f.IOStreams.ErrOut, "Workflow has execution history. Use --force to delete the workflow and all its runs.\n")
				return fmt.Errorf("workflow %s has existing runs that prevent deletion", workflowID)
			}
			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var result map[string]interface{}
			if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
				return fmt.Errorf("decoding delete response: %w", decodeErr)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(result, func(tw table.Writer) {
				fmt.Fprintf(f.IOStreams.Out, "Workflow %s deleted\n", workflowID)
				tw.Render()
			})
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false, "Force delete even if workflow has execution history")

	return cmd
}
