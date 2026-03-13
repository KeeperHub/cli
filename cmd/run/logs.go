package run

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// LogsResponse is the response from GET /api/workflows/executions/{id}/logs.
type LogsResponse struct {
	Execution any       `json:"execution"`
	Logs      []StepLog `json:"logs"`
}

// StepLog is a single step execution log entry.
type StepLog struct {
	ID          string  `json:"id"`
	NodeID      string  `json:"nodeId"`
	NodeName    string  `json:"nodeName"`
	NodeType    string  `json:"nodeType"`
	Status      string  `json:"status"`
	Input       any     `json:"input"`
	Output      any     `json:"output"`
	Error       *string `json:"error"`
	StartedAt   string  `json:"startedAt"`
	CompletedAt *string `json:"completedAt"`
	Duration    string  `json:"duration"`
}

// formatDuration converts a millisecond string (e.g. "1234") to a human-readable
// duration like "1.2s". Values under 1000ms are shown as "Nms".
func formatDuration(ms string) string {
	n, err := strconv.ParseFloat(ms, 64)
	if err != nil || ms == "" {
		return ms
	}
	if n < 1000 {
		return fmt.Sprintf("%.0fms", n)
	}
	return fmt.Sprintf("%.1fs", n/1000)
}

// truncateJSON marshals v to JSON and truncates to maxLen characters.
// If the JSON exceeds maxLen, it is truncated with "..." appended.
func truncateJSON(v any, maxLen int) string {
	if v == nil {
		return "-"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	s := string(b)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func NewLogsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logs <run-id>",
		Short:   "Show logs for a run",
		Aliases: []string{"l"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Show step logs for a run
  kh r l abc123

  # Show logs as JSON
  kh r l abc123 --json`,
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

			url := khhttp.BuildBaseURL(host) + "/api/workflows/executions/" + runID + "/logs"
			req, err := httpClient.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("creating request: %w", err)
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("fetching run logs: %w", err)
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
			var logsResp LogsResponse
			if err := json.Unmarshal(body, &logsResp); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(logsResp, func(tw table.Writer) {
				if len(logsResp.Logs) == 0 {
					fmt.Fprintf(f.IOStreams.Out, "No logs available for run %s\n", runID)
					return
				}
				tw.AppendHeader(table.Row{"STEP", "STATUS", "DURATION", "INPUT", "OUTPUT"})
				for _, log := range logsResp.Logs {
					errSuffix := ""
					if log.Error != nil && *log.Error != "" {
						errSuffix = " [error: " + *log.Error + "]"
					}
					tw.AppendRow(table.Row{
						log.NodeName + errSuffix,
						log.Status,
						formatDuration(log.Duration),
						truncateJSON(log.Input, 80),
						truncateJSON(log.Output, 80),
					})
				}
				tw.Render()
			})
		},
	}

	return cmd
}
