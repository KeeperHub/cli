package workflow

import (
	"bufio"
	"bytes"
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

func NewDisableCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "disable <workflow-id>",
		// pause is the original name and stays working indefinitely; scripts
		// and published links depend on it.
		Aliases: []string{"pause"},
		Short:   "Disable a workflow so it stops running",
		Long: `Disable a workflow so it stops running.

Turns off a workflow without deleting it. Runs already in flight are unaffected;
the trigger simply stops firing new ones.

See also: kh workflow enable`,
		Args: cobra.ExactArgs(1),
		Example: `  # Disable a workflow (will prompt for confirmation)
  kh wf disable abc123

  # Disable without prompting
  kh wf disable abc123 --yes

  # pause is an alias and still works
  kh wf pause abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID := args[0]

			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				// --yes not defined locally; check parent persistent flags
				yes = false
			}

			if !yes && f.IOStreams.IsTerminal() {
				fmt.Fprintf(f.IOStreams.Out, "Pause workflow %s? (y/N) ", workflowID)
				scanner := bufio.NewScanner(f.IOStreams.In)
				if scanner.Scan() {
					answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
					if answer != "y" && answer != "yes" {
						return cmdutil.CancelError{Err: fmt.Errorf("pause cancelled")}
					}
				}
			}
			// Non-TTY or --yes: proceed without prompting

			client, err := f.HTTPClient()
			if err != nil {
				return err
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			host := cmdutil.ResolveHost(cmd, cfg)

			bodyBytes, err := json.Marshal(map[string]bool{"enabled": false})
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

			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decoding pause response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(result, func(tw table.Writer) {
				fmt.Fprintf(f.IOStreams.Out, "Workflow %s disabled\n", workflowID)
				tw.Render()
			})
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	return cmd
}
