package execute

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// defaultPollInterval is used when the server sends no polling hint.
const defaultPollInterval = 2 * time.Second

// maxPollIntervalSecs caps the server's hint. watchExecStatus has no deadline, so an oversized or
// hostile hint would otherwise park the loop for as long as the server asked.
const maxPollIntervalSecs = 30

// nextPollDelay reads the X-Poll-Interval-Hint response header, which the Direct Execution API
// documents as the number of seconds to wait before polling again. Honouring it lets the server
// pace clients instead of every client polling on its own fixed timer. The hint is clamped to
// maxPollIntervalSecs so it can slow a caller down but not stall it.
//
// A hint of 0 means the execution has reached a terminal state, reported here as (0, true) so a
// caller stops rather than sleeping for zero and spinning. A negative or unparseable hint falls
// back to the default rather than yielding a zero delay the loop would spin on.
func nextPollDelay(resp *http.Response) (time.Duration, bool) {
	raw := strings.TrimSpace(resp.Header.Get("X-Poll-Interval-Hint"))
	if raw == "" {
		return defaultPollInterval, false
	}
	secs, err := strconv.Atoi(raw)
	if err != nil || secs < 0 {
		return defaultPollInterval, false
	}
	if secs == 0 {
		return 0, true
	}
	return time.Duration(min(secs, maxPollIntervalSecs)) * time.Second, false
}

// ExecStatusResponse represents the execution status API response.
// Shared by transfer, contract-call and status commands.
type ExecStatusResponse struct {
	ExecutionID     string  `json:"executionId"`
	Status          string  `json:"status"`
	Type            string  `json:"type"`
	TransactionHash *string `json:"transactionHash"`
	TransactionLink *string `json:"transactionLink"`
	Result          any     `json:"result"`
	Error           *string `json:"error"`
	CreatedAt       string  `json:"createdAt"`
	CompletedAt     *string `json:"completedAt"`
}

func NewStatusCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <execution-id>",
		Short:   "Show the status of an execution",
		Aliases: []string{"st"},
		Args:    cobra.ExactArgs(1),
		Long: `Show the status of a direct blockchain execution (transfer or contract call).
Use --watch to poll until the execution reaches a terminal state.

See also: kh r st, kh ex transfer, kh ex cc`,
		Example: `  # Show execution status
  kh ex st abc123

  # Watch until completion
  kh ex st abc123 --watch`,
		RunE: func(cmd *cobra.Command, args []string) error {
			executionID := args[0]

			client, err := f.HTTPClient()
			if err != nil {
				return err
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			host := cmdutil.ResolveHost(cmd, cfg)

			watch, _ := cmd.Flags().GetBool("watch")

			p := output.NewPrinter(f.IOStreams, cmd)

			if !watch {
				sr, _, _, fetchErr := fetchExecStatus(client, host, executionID)
				if fetchErr != nil {
					return fetchErr
				}
				return renderExecStatus(p, f, sr)
			}

			return watchExecStatus(f, client, host, executionID, p)
		},
	}

	cmd.Flags().Bool("watch", false, "Live-update until complete")

	return cmd
}

func renderExecStatus(p *output.Printer, f *cmdutil.Factory, sr *ExecStatusResponse) error {
	if err := p.PrintData(sr, func(tw table.Writer) {
		tw.AppendRow(table.Row{"Execution", sr.ExecutionID})
		tw.AppendRow(table.Row{"Status", sr.Status})
		if sr.Type != "" {
			tw.AppendRow(table.Row{"Type", sr.Type})
		}
		if sr.TransactionHash != nil && *sr.TransactionHash != "" {
			tw.AppendRow(table.Row{"TX Hash", *sr.TransactionHash})
		}
		if sr.TransactionLink != nil && *sr.TransactionLink != "" {
			tw.AppendRow(table.Row{"TX Link", *sr.TransactionLink})
		}
		if completedWithoutTransaction(sr.Status, sr.TransactionHash) {
			tw.AppendRow(table.Row{"Transaction", noTransactionNote})
		}
		if sr.CreatedAt != "" {
			tw.AppendRow(table.Row{"Created", sr.CreatedAt})
		}
		if sr.CompletedAt != nil && *sr.CompletedAt != "" {
			tw.AppendRow(table.Row{"Completed", *sr.CompletedAt})
		}
		if sr.Error != nil && *sr.Error != "" {
			tw.AppendRow(table.Row{"Error", *sr.Error})
		}
		tw.Render()
	}); err != nil {
		return err
	}

	if sr.Status == "failed" {
		msg := fmt.Sprintf("execution %s failed", sr.ExecutionID)
		if sr.Error != nil && *sr.Error != "" {
			msg = *sr.Error
		}
		return fmt.Errorf("%s", msg)
	}

	return nil
}

func watchExecStatus(f *cmdutil.Factory, client *khhttp.Client, host, executionID string, p *output.Printer) error {
	isTTY := f.IOStreams.IsTerminal()

	for {
		sr, delay, serverSaysTerminal, err := fetchExecStatus(client, host, executionID)
		if err != nil {
			return err
		}

		if isTTY && !p.IsJSON() {
			fmt.Fprintf(f.IOStreams.Out, "\r%s  %s", executionID, sr.Status)
		}

		if execTerminalStatuses[sr.Status] || serverSaysTerminal {
			if isTTY && !p.IsJSON() {
				fmt.Fprintln(f.IOStreams.Out)
			}
			return renderExecStatus(p, f, sr)
		}

		if delay > 0 {
			time.Sleep(delay)
		}
	}
}

