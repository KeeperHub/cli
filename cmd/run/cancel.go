package run

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// CancelResponse is the response from POST /api/executions/{id}/cancel.
type CancelResponse struct {
	Success bool `json:"success"`
}

func NewCancelCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <run-id>",
		Short: "Cancel a run",
		Args:  cobra.ExactArgs(1),
		Example: `  # Cancel a run (will prompt for confirmation)
  kh r cancel abc123

  # Cancel without prompting
  kh r cancel abc123 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]

			httpClient, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			host, _ := cmd.Flags().GetString("host")
			if host == "" {
				host = cfg.DefaultHost
			}
			if host == "" {
				host = "app.keeperhub.com"
			}

			// Confirmation: skip if --yes or non-TTY (auto-proceed in non-interactive mode).
			yes, _ := cmd.Flags().GetBool("yes")
			isTTY := f.IOStreams.IsTerminal()
			if !yes && isTTY {
				fmt.Fprintf(f.IOStreams.Out, "Cancel run %s? (y/N) ", runID)
				scanner := bufio.NewScanner(f.IOStreams.In)
				scanner.Scan()
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					return cmdutil.CancelError{Err: fmt.Errorf("cancelled")}
				}
			}

			url := khhttp.BuildBaseURL(host) + "/api/executions/" + runID + "/cancel"
			req, err := httpClient.NewRequest(http.MethodPost, url, nil)
			if err != nil {
				return fmt.Errorf("creating request: %w", err)
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("cancelling run: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				return fmt.Errorf("run monitoring requires interactive login. Use 'kh auth login' first")
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status %d from server", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}
			var cancelResp CancelResponse
			if err := json.Unmarshal(body, &cancelResp); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(cancelResp, func(tw table.Writer) {
				fmt.Fprintf(f.IOStreams.Out, "Run %s cancelled\n", runID)
			})
		},
	}

	return cmd
}
