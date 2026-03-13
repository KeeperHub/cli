package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type executeResponse struct {
	ExecutionID string `json:"executionId"`
	Status      string `json:"status"`
}

type statusResponse struct {
	Status   string `json:"status"`
	Progress struct {
		TotalSteps      int    `json:"totalSteps"`
		CompletedSteps  int    `json:"completedSteps"`
		CurrentNodeName string `json:"currentNodeName"`
		Percentage      int    `json:"percentage"`
	} `json:"progress"`
}

var terminalStatuses = map[string]bool{
	"success":   true,
	"error":     true,
	"cancelled": true,
}

func NewRunCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run <workflow-id>",
		Short:   "Run a workflow",
		Aliases: []string{"r"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID := args[0]

			wait, err := cmd.Flags().GetBool("wait")
			if err != nil {
				return err
			}
			timeout, err := cmd.Flags().GetDuration("timeout")
			if err != nil {
				return err
			}

			client, err := f.HTTPClient()
			if err != nil {
				return err
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			host := cfg.DefaultHost

			execURL := host + "/api/workflow/" + workflowID + "/execute"
			req, err := client.NewRequest(http.MethodPost, execURL, bytes.NewBufferString("{}"))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusAccepted {
				return khhttp.NewAPIError(resp)
			}

			var execResp executeResponse
			if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
				return fmt.Errorf("decoding execute response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)

			if !wait {
				if p.IsJSON() {
					return p.PrintJSON(execResp)
				}
				fmt.Fprintf(f.IOStreams.Out, "Triggered run: %s\n", execResp.ExecutionID)
				return nil
			}

			return runWaitLoop(f, client, host, execResp.ExecutionID, timeout, p)
		},
	}

	cmd.Flags().Bool("wait", false, "Wait for completion")
	cmd.Flags().Duration("timeout", 5*time.Minute, "Timeout when using --wait")

	return cmd
}

func runWaitLoop(f *cmdutil.Factory, client *khhttp.Client, host, executionID string, timeout time.Duration, p *output.Printer) error {
	deadline := time.Now().Add(timeout)
	pollInterval := 2 * time.Second

	isTTY := f.IOStreams.IsTerminal()

	var finalStatus statusResponse

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout after %s: run %s timed out", timeout, executionID)
		}

		sr, err := fetchStatus(client, host, executionID)
		if err != nil {
			return err
		}

		if isTTY && !p.IsJSON() {
			fmt.Fprintf(f.IOStreams.Out, "\r%s  %d/%d steps (%d%%)",
				executionID,
				sr.Progress.CompletedSteps,
				sr.Progress.TotalSteps,
				sr.Progress.Percentage,
			)
		}

		if terminalStatuses[sr.Status] {
			finalStatus = *sr
			break
		}

		// Sleep with deadline awareness
		sleepUntil := time.Now().Add(pollInterval)
		if sleepUntil.After(deadline) {
			sleepUntil = deadline
		}
		time.Sleep(time.Until(sleepUntil))
	}

	if isTTY && !p.IsJSON() {
		fmt.Fprintln(f.IOStreams.Out)
	}

	if p.IsJSON() {
		return p.PrintJSON(finalStatus)
	}

	if finalStatus.Status == "error" {
		return fmt.Errorf("run %s finished with status: error", executionID)
	}

	p.PrintTable(func(tw table.Writer) {
		tw.AppendHeader(table.Row{"EXECUTION ID", "STATUS", "STEPS"})
		tw.AppendRow(table.Row{
			executionID,
			finalStatus.Status,
			fmt.Sprintf("%d/%d", finalStatus.Progress.CompletedSteps, finalStatus.Progress.TotalSteps),
		})
		tw.Render()
	})

	return nil
}

func fetchStatus(client *khhttp.Client, host, executionID string) (*statusResponse, error) {
	statusURL := host + "/api/workflows/executions/" + executionID + "/status"
	req, err := client.NewRequest(http.MethodGet, statusURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, khhttp.NewAPIError(resp)
	}

	var sr statusResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("decoding status response: %w", err)
	}
	return &sr, nil
}
