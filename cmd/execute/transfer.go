package execute

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/keeperhub/cli/internal/execrecovery"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type transferRequest struct {
	Network          string `json:"network"`
	RecipientAddress string `json:"recipientAddress"`
	Amount           string `json:"amount"`
	TokenAddress     string `json:"tokenAddress,omitempty"`
}

type transferResponse struct {
	ExecutionID     string  `json:"executionId"`
	Status          string  `json:"status"`
	TransactionHash *string `json:"transactionHash,omitempty"`
}

// execStatusUnconfirmed is terminal for the CLI wait/watch loops.
//
// It is not terminal on the server: a reconciliation sweep still settles that
// row to completed or failed once the chain answers. The CLI stops on it
// anyway and reports it, rather than polling to a non-zero timeout, because a
// non-zero exit invites a re-run that broadcasts a second transaction for an
// intent that may already be on chain. Read the settled status later with
// `kh ex st <id>`.
const execStatusUnconfirmed = "unconfirmed"

var execTerminalStatuses = map[string]bool{
	"completed":           true,
	"failed":              true,
	execStatusUnconfirmed: true,
}

// printUnconfirmedNotice reports an unconfirmed execution on stderr: which
// transaction was broadcast, and that the reconciler is still watching it so
// the execution can be re-checked later.
func printUnconfirmedNotice(f *cmdutil.Factory, executionID string, txHash *string) {
	hash := "not reported"
	if txHash != nil && *txHash != "" {
		hash = *txHash
	}
	fmt.Fprintf(f.IOStreams.ErrOut,
		"execution %s is unconfirmed: transaction %s was broadcast but no receipt could be read yet.\nThe reconciler is still watching it - re-check later with: kh ex st %s\n",
		executionID, hash, executionID)
}

// terminalExecError reports a terminal status that did not succeed on the
// write-response path, which carries no receipt or error detail.
func terminalExecError(executionID, status string, apiErr *string) error {
	if status != "failed" {
		return nil
	}
	if apiErr != nil && *apiErr != "" {
		return fmt.Errorf("%s", *apiErr)
	}
	return fmt.Errorf("execution %s failed", executionID)
}

func NewTransferCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfer",
		Short:   "Transfer tokens",
		Aliases: []string{"t"},
		Example: `  # Transfer ETH and wait for completion
  kh ex t --chain 1 --to 0xABCD... --amount 0.01 --wait

  # Transfer an ERC-20 token
  kh ex t --chain 1 --to 0xABCD... --amount 100 --token-address 0xUSDC...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return err
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			host := cmdutil.ResolveHost(cmd, cfg)

			chain, _ := cmd.Flags().GetString("chain")
			to, _ := cmd.Flags().GetString("to")
			amount, _ := cmd.Flags().GetString("amount")
			token, _ := cmd.Flags().GetString("token")
			tokenAddress, _ := cmd.Flags().GetString("token-address")
			wait, _ := cmd.Flags().GetBool("wait")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			idemKeyFlag, _ := cmd.Flags().GetString("idempotency-key")

			body := transferRequest{
				Network:          chain,
				RecipientAddress: to,
				Amount:           amount,
			}

			if tokenAddress != "" {
				body.TokenAddress = tokenAddress
			} else if token != "ETH" {
				return cmdutil.FlagError{Err: fmt.Errorf("--token-address is required when --token is not ETH")}
			}

			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("marshalling request: %w", err)
			}

			idemKey, err := execrecovery.ResolveIdempotencyKey(idemKeyFlag)
			if err != nil {
				return err
			}

			deadline := time.Now().Add(timeout)
			resp, err := postIdempotentJSON(client, khhttp.BuildBaseURL(host)+"/api/execute/transfer", bodyBytes, idemKey, deadline)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusAccepted {
				return khhttp.NewAPIError(resp)
			}

			var execResp transferResponse
			if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)

			if !wait {
				return p.PrintData(execResp, func(tw table.Writer) {
					tw.AppendRow(table.Row{"Execution", execResp.ExecutionID})
					tw.AppendRow(table.Row{"Status", execResp.Status})
					tw.Render()
				})
			}

			if execTerminalStatuses[execResp.Status] {
				if err := terminalExecError(execResp.ExecutionID, execResp.Status, nil); err != nil {
					return err
				}
				if err := printTransferResult(p, &execResp); err != nil {
					return err
				}
				if execResp.Status == execStatusUnconfirmed {
					printUnconfirmedNotice(f, execResp.ExecutionID, execResp.TransactionHash)
				}
				return nil
			}

			return pollExecStatus(f, client, host, execResp.ExecutionID, timeout, p)
		},
	}

	cmd.Flags().String("chain", "", "Chain ID (required)")
	cmd.Flags().String("to", "", "Recipient address (required)")
	cmd.Flags().String("amount", "", "Amount to transfer (required)")
	cmd.Flags().String("token", "ETH", "Token symbol")
	cmd.Flags().String("token-address", "", "ERC-20 token contract address")
	cmd.Flags().Bool("wait", false, "Wait for completion")
	cmd.Flags().Duration("timeout", 5*time.Minute, "Timeout when using --wait")
	cmd.Flags().String("idempotency-key", "", "Stable Idempotency-Key for this write intent (auto-generated if empty)")

	_ = cmd.MarkFlagRequired("chain")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("amount")

	return cmd
}

func printTransferResult(p *output.Printer, execResp *transferResponse) error {
	return p.PrintData(execResp, func(tw table.Writer) {
		tw.AppendRow(table.Row{"Execution", execResp.ExecutionID})
		tw.AppendRow(table.Row{"Status", execResp.Status})
		if execResp.TransactionHash != nil && *execResp.TransactionHash != "" {
			tw.AppendRow(table.Row{"TX Hash", *execResp.TransactionHash})
		}
		tw.Render()
	})
}

func pollExecStatus(f *cmdutil.Factory, client *khhttp.Client, host, executionID string, timeout time.Duration, p *output.Printer) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			statusResp, err := fetchExecStatus(client, host, executionID)
			if err != nil {
				var apiErr *khhttp.APIError
				// R6: tolerate cold-start 404 until the wait deadline.
				if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
					if time.Now().After(deadline) {
						return fmt.Errorf("timeout after %s: execution %s not found", timeout, executionID)
					}
					continue
				}
				return err
			}

			if execTerminalStatuses[statusResp.Status] {
				if err := execOutcomeError(statusResp); err != nil {
					_ = printExecStatusResult(p, statusResp)
					return err
				}
				if err := printExecStatusResult(p, statusResp); err != nil {
					return err
				}
				if statusResp.Status == execStatusUnconfirmed {
					printUnconfirmedNotice(f, executionID, statusResp.TransactionHash)
				}
				return nil
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("timeout after %s: execution %s still %s", timeout, executionID, statusResp.Status)
			}
		default:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout after %s: execution %s timed out", timeout, executionID)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func fetchExecStatus(client *khhttp.Client, host, executionID string) (*ExecStatusResponse, error) {
	url := khhttp.BuildBaseURL(host) + "/api/execute/" + executionID + "/status"
	req, err := client.NewRequest(http.MethodGet, url, nil)
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

	var sr ExecStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("decoding status response: %w", err)
	}
	return &sr, nil
}

func printExecStatusResult(p *output.Printer, sr *ExecStatusResponse) error {
	return p.PrintData(sr, func(tw table.Writer) {
		tw.AppendRow(table.Row{"Execution", sr.ExecutionID})
		tw.AppendRow(table.Row{"Status", sr.Status})
		if sr.TransactionHash != nil && *sr.TransactionHash != "" {
			tw.AppendRow(table.Row{"TX Hash", *sr.TransactionHash})
		}
		if sr.TransactionLink != nil && *sr.TransactionLink != "" {
			tw.AppendRow(table.Row{"TX Link", *sr.TransactionLink})
		}
		for i, r := range sr.Receipts {
			tw.AppendRow(table.Row{fmt.Sprintf("Receipt[%d]", i), fmt.Sprintf("%s verified=%v status=%s", r.Hash, r.Verified, r.ReceiptStatus)})
		}
		tw.Render()
	})
}

// execOutcomeError returns a non-nil error for failed terminal states and for
// any receipt that is not explicitly successful when the run has completed, or
// for a conclusive on-chain failure (reverted / safe_inner_failure) at any
// status. not_found / timeout receipts leave `unconfirmed` a zero-exit
// outcome: the server treats those as unread, not failed, so erroring here
// would invite the re-run that double-broadcasts.
func execOutcomeError(sr *ExecStatusResponse) error {
	if sr.Status == "failed" {
		msg := fmt.Sprintf("execution %s failed", sr.ExecutionID)
		if sr.Error != nil && *sr.Error != "" {
			msg = *sr.Error
		}
		return fmt.Errorf("%s", msg)
	}
	for _, r := range sr.Receipts {
		if execrecovery.ConclusiveFailedReceipt(r.ReceiptStatus) {
			return fmt.Errorf("execution %s receipt %s status=%s", sr.ExecutionID, r.Hash, r.ReceiptStatus)
		}
		if sr.Status == "completed" && execrecovery.NonSuccessReceipt(r) {
			return fmt.Errorf("execution %s completed with non-success receipt %s status=%s", sr.ExecutionID, r.Hash, r.ReceiptStatus)
		}
	}
	return nil
}
