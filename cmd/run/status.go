package run

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// buildBaseURL normalises a host string into a base URL.
// If host already starts with http:// or https://, it is returned as-is.
// Otherwise https:// is prepended.
func buildBaseURL(host string) string {
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return strings.TrimRight(host, "/")
	}
	return "https://" + strings.TrimRight(host, "/")
}

// RunStatusResponse is the response from GET /api/workflows/executions/{id}/status.
type RunStatusResponse struct {
	Status       string       `json:"status"`
	NodeStatuses []NodeStatus `json:"nodeStatuses"`
	Progress     RunProgress  `json:"progress"`
	ErrorContext any          `json:"errorContext"`
}

// NodeStatus holds per-node execution status.
type NodeStatus struct {
	NodeID string `json:"nodeId"`
	Status string `json:"status"`
}

// RunProgress holds step-level progress counters.
type RunProgress struct {
	TotalSteps      int     `json:"totalSteps"`
	CompletedSteps  int     `json:"completedSteps"`
	RunningSteps    int     `json:"runningSteps"`
	CurrentNodeID   *string `json:"currentNodeId"`
	CurrentNodeName *string `json:"currentNodeName"`
	Percentage      int     `json:"percentage"`
}

var terminalStatuses = map[string]bool{
	"success":   true,
	"error":     true,
	"cancelled": true,
}

func NewStatusCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <run-id>",
		Short:   "Show the status of a run",
		Aliases: []string{"st"},
		Args:    cobra.ExactArgs(1),
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
				host = "app.keeperhub.io"
			}

			watch, _ := cmd.Flags().GetBool("watch")
			p := output.NewPrinter(f.IOStreams, cmd)

			fetchStatus := func() (*RunStatusResponse, error) {
				url := buildBaseURL(host) + "/api/workflows/executions/" + runID + "/status"
				req, reqErr := httpClient.NewRequest(http.MethodGet, url, nil)
				if reqErr != nil {
					return nil, fmt.Errorf("creating request: %w", reqErr)
				}
				resp, doErr := httpClient.Do(req)
				if doErr != nil {
					return nil, fmt.Errorf("fetching run status: %w", doErr)
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusUnauthorized {
					return nil, fmt.Errorf("run monitoring requires interactive login. Use 'kh auth login' first")
				}
				if resp.StatusCode != http.StatusOK {
					return nil, fmt.Errorf("unexpected status %d from server", resp.StatusCode)
				}

				body, readErr := io.ReadAll(resp.Body)
				if readErr != nil {
					return nil, fmt.Errorf("reading response: %w", readErr)
				}
				var status RunStatusResponse
				if parseErr := json.Unmarshal(body, &status); parseErr != nil {
					return nil, fmt.Errorf("parsing response: %w", parseErr)
				}
				return &status, nil
			}

			printSummary := func(status *RunStatusResponse) error {
				return p.PrintData(status, func(tw table.Writer) {
					tw.AppendHeader(table.Row{"FIELD", "VALUE"})
					tw.AppendRow(table.Row{"Status", status.Status})
					tw.AppendRow(table.Row{"Steps", fmt.Sprintf("%d/%d", status.Progress.CompletedSteps, status.Progress.TotalSteps)})
					currentStep := "-"
					if status.Progress.CurrentNodeName != nil {
						currentStep = *status.Progress.CurrentNodeName
					}
					tw.AppendRow(table.Row{"Current step", currentStep})
					tw.AppendRow(table.Row{"Percentage", fmt.Sprintf("%d%%", status.Progress.Percentage)})
					tw.Render()
				})
			}

			if !watch {
				status, fetchErr := fetchStatus()
				if fetchErr != nil {
					return fetchErr
				}
				return printSummary(status)
			}

			// Watch mode: poll until terminal status.
			// No timeout -- watch is purely observational; user presses Ctrl+C to exit.
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			isTTY := f.IOStreams.IsTerminal()

			for {
				status, fetchErr := fetchStatus()
				if fetchErr != nil {
					return fetchErr
				}

				isTerminal := terminalStatuses[status.Status]

				if !p.IsJSON() {
					nodeName := status.Status
					if status.Progress.CurrentNodeName != nil {
						nodeName = *status.Progress.CurrentNodeName
					}
					line := fmt.Sprintf("%s  %d/%d steps (%d%%)  %s",
						runID,
						status.Progress.CompletedSteps,
						status.Progress.TotalSteps,
						status.Progress.Percentage,
						nodeName,
					)
					if isTTY {
						fmt.Fprintf(f.IOStreams.Out, "\r%s", line)
					} else {
						fmt.Fprintln(f.IOStreams.Out, line)
					}
				}

				if isTerminal {
					if isTTY && !p.IsJSON() {
						fmt.Fprintln(f.IOStreams.Out, "")
					}
					if printErr := printSummary(status); printErr != nil {
						return printErr
					}
					if status.Status == "error" {
						if status.ErrorContext != nil {
							fmt.Fprintf(f.IOStreams.ErrOut, "Error: %v\n", status.ErrorContext)
						}
						return fmt.Errorf("run %s failed", runID)
					}
					return nil
				}

				<-ticker.C
			}
		},
	}

	cmd.Flags().Bool("watch", false, "Live-update until complete")

	return cmd
}
