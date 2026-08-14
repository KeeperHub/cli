package execute

import (
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// ExecStatusResponse represents the execution status API response.
// Shared by transfer, contract-call and status commands.
type ExecStatusResponse struct {
	ExecutionID     string        `json:"executionId"`
	Status          string        `json:"status"`
	Type            string        `json:"type"`
	TransactionHash *string       `json:"transactionHash"`
	TransactionLink *string       `json:"transactionLink"`
	Result          any           `json:"result"`
	Error           *string       `json:"error"`
	CreatedAt       string        `json:"createdAt"`
	CompletedAt     *string       `json:"completedAt"`
	Receipts        []ExecReceipt `json:"receipts,omitempty"`
}

// ExecReceipt is a chain-re-fetched proof entry attached to an execution.
// A transactionHash alone proves a transaction was submitted; a receipt with
// verified=true and receiptStatus="success" proves it landed onchain.
type ExecReceipt struct {
	Hash          string  `json:"hash"`
	ChainID       int64   `json:"chainId"`
	Verified      bool    `json:"verified"`
	ReceiptStatus string  `json:"receiptStatus"`
	BlockNumber   *int64  `json:"blockNumber,omitempty"`
	GasUsed       *string `json:"gasUsed,omitempty"`
	VerifiedAt    *string `json:"verifiedAt,omitempty"`
}

func NewStatusCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <execution-id>",
		Short:   "Show the status of an execution",
		Aliases: []string{"st"},
		Args:    cobra.ExactArgs(1),
		Long: `Show the status of a direct blockchain execution (transfer or contract call).
Use --watch to poll until the execution reaches a terminal state.

Use --require-verified to fail unless the execution completed AND every
onchain receipt is chain-verified with receiptStatus "success". A completed
status without receipts exits non-zero: submitted is not the same as landed.

See also: kh r st, kh ex transfer, kh ex cc`,
		Example: `  # Show execution status
  kh ex st abc123

  # Watch until completion
  kh ex st abc123 --watch

  # Gate a script on chain-verified success. --watch polls until the
  # execution reaches a terminal status and has no deadline of its own,
  # so bound it externally when running unattended.
  kh ex st abc123 --watch --require-verified && ./next-step.sh`,
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
			requireVerified, _ := cmd.Flags().GetBool("require-verified")

			p := output.NewPrinter(f.IOStreams, cmd)

			if !watch {
				sr, fetchErr := fetchExecStatus(client, host, executionID)
				if fetchErr != nil {
					return fetchErr
				}
				return renderExecStatusChecked(p, f, sr, requireVerified)
			}

			return watchExecStatus(f, client, host, executionID, requireVerified, p)
		},
	}

	cmd.Flags().Bool("watch", false, "Live-update until complete")
	cmd.Flags().Bool("require-verified", false, "Exit non-zero unless completed with chain-verified success receipts")

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
		if sr.CreatedAt != "" {
			tw.AppendRow(table.Row{"Created", sr.CreatedAt})
		}
		if sr.CompletedAt != nil && *sr.CompletedAt != "" {
			tw.AppendRow(table.Row{"Completed", *sr.CompletedAt})
		}
		for _, r := range sr.Receipts {
			state := r.ReceiptStatus
			if r.Verified {
				state += ", verified"
			} else {
				state += ", unverified"
			}
			tw.AppendRow(table.Row{"Receipt", fmt.Sprintf("%s (%s)", r.Hash, state)})
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

// renderExecStatusChecked renders the status and, when requireVerified is set,
// additionally enforces chain-verified success receipts on completed executions.
func renderExecStatusChecked(p *output.Printer, f *cmdutil.Factory, sr *ExecStatusResponse, requireVerified bool) error {
	if err := renderExecStatus(p, f, sr); err != nil {
		return err
	}
	if requireVerified {
		return verifyExecReceipts(sr)
	}
	return nil
}

// verifyExecReceipts fails closed: a completed status only counts as landed
// when at least one receipt exists and every receipt is chain-verified with
// receiptStatus "success". not_found and timeout receipts are treated as
// unproven, not as success.
func verifyExecReceipts(sr *ExecStatusResponse) error {
	if sr.Status != "completed" {
		return fmt.Errorf("execution %s is %s, not completed", sr.ExecutionID, sr.Status)
	}
	if len(sr.Receipts) == 0 {
		return fmt.Errorf("execution %s completed with no receipts: there is nothing to chain-verify", sr.ExecutionID)
	}
	for _, r := range sr.Receipts {
		if !r.Verified {
			return fmt.Errorf("receipt %s is not chain-verified (verified=false)", r.Hash)
		}
		if r.ReceiptStatus != "success" {
			return fmt.Errorf("receipt %s has receiptStatus %q, want \"success\"", r.Hash, r.ReceiptStatus)
		}
	}
	return nil
}

func watchExecStatus(f *cmdutil.Factory, client *khhttp.Client, host, executionID string, requireVerified bool, p *output.Printer) error {
	isTTY := f.IOStreams.IsTerminal()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sr, err := fetchExecStatus(client, host, executionID)
			if err != nil {
				return err
			}

			if isTTY && !p.IsJSON() {
				fmt.Fprintf(f.IOStreams.Out, "\r%s  %s", executionID, sr.Status)
			}

			if execTerminalStatuses[sr.Status] {
				if isTTY && !p.IsJSON() {
					fmt.Fprintln(f.IOStreams.Out)
				}
				return renderExecStatusChecked(p, f, sr, requireVerified)
			}
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}
