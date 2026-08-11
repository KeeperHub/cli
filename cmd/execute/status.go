package execute

import (
	"errors"
	"fmt"
	"net/http"
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
// receiptStatus="reverted" is Failure even when status=completed.
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
				sr, fetchErr := fetchExecStatus(client, host, executionID)
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
		if sr.CreatedAt != "" {
			tw.AppendRow(table.Row{"Created", sr.CreatedAt})
		}
		if sr.CompletedAt != nil && *sr.CompletedAt != "" {
			tw.AppendRow(table.Row{"Completed", *sr.CompletedAt})
		}
		if sr.Error != nil && *sr.Error != "" {
			tw.AppendRow(table.Row{"Error", *sr.Error})
		}
		for i, r := range sr.Receipts {
			label := fmt.Sprintf("Receipt[%d]", i)
			tw.AppendRow(table.Row{label, fmt.Sprintf("%s verified=%v status=%s", r.Hash, r.Verified, r.ReceiptStatus)})
		}
		tw.Render()
	}); err != nil {
		return err
	}

	if err := execOutcomeError(sr); err != nil {
		return err
	}

	return nil
}

func watchExecStatus(f *cmdutil.Factory, client *khhttp.Client, host, executionID string, p *output.Printer) error {
	isTTY := f.IOStreams.IsTerminal()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sr, err := fetchExecStatus(client, host, executionID)
			if err != nil {
				var apiErr *khhttp.APIError
				if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
					if isTTY && !p.IsJSON() {
						fmt.Fprintf(f.IOStreams.Out, "\r%s  not_found", executionID)
					}
					continue
				}
				return err
			}

			if isTTY && !p.IsJSON() {
				fmt.Fprintf(f.IOStreams.Out, "\r%s  %s", executionID, sr.Status)
			}

			if execTerminalStatuses[sr.Status] {
				if isTTY && !p.IsJSON() {
					fmt.Fprintln(f.IOStreams.Out)
				}
				return renderExecStatus(p, f, sr)
			}
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

