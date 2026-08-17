package execute

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

// execTerminalStatuses are the statuses a poll loop stops on.
//
// unconfirmed is terminal for a client: the server hands it back once a transaction was broadcast
// but no receipt was observed, and nothing moves it until the reconciler runs on its own schedule,
// so polling past it only burns requests. --wait and --watch stop there and exit zero, because a
// non-zero exit invites a retry, and retrying a broadcast can double-spend.
var execTerminalStatuses = map[string]bool{
	"completed":   true,
	"failed":      true,
	"unconfirmed": true,
}

// noTransactionNote labels a completed execution that put nothing onchain.
const noTransactionNote = "none submitted"

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

			req, err := client.NewRequest(http.MethodPost, khhttp.BuildBaseURL(host)+"/api/execute/transfer", bytes.NewReader(bodyBytes))
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

			if execTerminalStatuses[execResp.Status] && !completedWithoutTransaction(execResp.Status, execResp.TransactionHash) {
				return printTransferResult(p, &execResp)
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

	for {
		statusResp, delay, serverSaysTerminal, err := fetchExecStatus(client, host, executionID)
		if err != nil {
			return err
		}

		if execTerminalStatuses[statusResp.Status] || serverSaysTerminal {
			if statusResp.Status == "failed" {
				msg := fmt.Sprintf("execution %s failed", executionID)
				if statusResp.Error != nil {
					msg = *statusResp.Error
				}
				return fmt.Errorf("%s", msg)
			}
			return printExecStatusResult(p, statusResp)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout after %s: execution %s still %s", timeout, executionID, statusResp.Status)
		}

		// Wait as long as the server asked, but never past the caller's deadline.
		if remaining := time.Until(deadline); delay > remaining {
			delay = remaining
		}
		if delay > 0 {
			time.Sleep(delay)
		}
	}
}

// fetchExecStatus returns the execution status, the delay the server asked us to wait before
// polling again, and whether the server declared the execution terminal via a hint of 0.
func fetchExecStatus(client *khhttp.Client, host, executionID string) (*ExecStatusResponse, time.Duration, bool, error) {
	url := khhttp.BuildBaseURL(host) + "/api/execute/" + executionID + "/status"
	req, err := client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, false, khhttp.NewAPIError(resp)
	}

	delay, serverSaysTerminal := nextPollDelay(resp)

	var sr ExecStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, 0, false, fmt.Errorf("decoding status response: %w", err)
	}
	return &sr, delay, serverSaysTerminal, nil
}

// completedWithoutTransaction reports an execution that reached "completed" carrying no
// transaction hash.
//
// On a direct-write response it means one status fetch is worth making: a successful
// /api/execute/contract-call broadcast can return 202 with status "completed" and no
// transactionHash, because the hash only appears on the status endpoint, so --wait would
// otherwise report a completion the caller cannot tie to a transaction.
//
// On a status response it is not an error. Any action that submits nothing onchain - a read-only
// contract call, a non-transaction plugin step - completes exactly this way by design, so the
// condition is reported in the output and the exit code is left alone. Failing on it is opt-in
// behaviour and belongs behind an explicit flag.
func completedWithoutTransaction(status string, txHash *string) bool {
	return status == "completed" && (txHash == nil || *txHash == "")
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
		if completedWithoutTransaction(sr.Status, sr.TransactionHash) {
			tw.AppendRow(table.Row{"Transaction", noTransactionNote})
		}
		tw.Render()
	})
}
