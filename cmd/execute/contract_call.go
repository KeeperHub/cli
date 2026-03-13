package execute

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type contractCallRequest struct {
	ContractAddress string `json:"contractAddress"`
	Network         string `json:"network"`
	FunctionName    string `json:"functionName"`
	FunctionArgs    string `json:"functionArgs,omitempty"`
	ABI             string `json:"abi,omitempty"`
}

type contractCallReadResponse struct {
	Result any `json:"result"`
}

type contractCallWriteResponse struct {
	ExecutionID     string  `json:"executionId"`
	Status          string  `json:"status"`
	TransactionHash *string `json:"transactionHash,omitempty"`
}

func NewContractCallCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contract-call",
		Short:   "Call a smart contract method",
		Aliases: []string{"cc"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return err
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			host := cfg.DefaultHost

			chain, _ := cmd.Flags().GetString("chain")
			contract, _ := cmd.Flags().GetString("contract")
			method, _ := cmd.Flags().GetString("method")
			argsStr, _ := cmd.Flags().GetString("args")
			abiFile, _ := cmd.Flags().GetString("abi-file")
			wait, _ := cmd.Flags().GetBool("wait")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			reqBody := contractCallRequest{
				ContractAddress: contract,
				Network:         chain,
				FunctionName:    method,
			}

			if argsStr != "" {
				reqBody.FunctionArgs = argsStr
			}

			if abiFile != "" {
				abiBytes, readErr := os.ReadFile(abiFile)
				if readErr != nil {
					return fmt.Errorf("reading abi file: %w", readErr)
				}
				if !json.Valid(abiBytes) {
					return cmdutil.FlagError{Err: fmt.Errorf("--abi-file does not contain valid JSON")}
				}
				reqBody.ABI = string(abiBytes)
			}

			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("marshalling request: %w", err)
			}

			req, err := client.NewRequest(http.MethodPost, host+"/api/execute/contract-call", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			p := output.NewPrinter(f.IOStreams, cmd)

			switch resp.StatusCode {
			case http.StatusOK:
				var readResp contractCallReadResponse
				if err := json.NewDecoder(resp.Body).Decode(&readResp); err != nil {
					return fmt.Errorf("decoding response: %w", err)
				}
				return printContractCallReadResult(p, &readResp)

			case http.StatusAccepted:
				var writeResp contractCallWriteResponse
				if err := json.NewDecoder(resp.Body).Decode(&writeResp); err != nil {
					return fmt.Errorf("decoding response: %w", err)
				}

				if !wait {
					return p.PrintData(writeResp, func(tw table.Writer) {
						tw.AppendRow(table.Row{"Execution", writeResp.ExecutionID})
						tw.AppendRow(table.Row{"Status", writeResp.Status})
						tw.Render()
					})
				}

				if execTerminalStatuses[writeResp.Status] {
					return printContractCallWriteResult(p, &writeResp)
				}

				return pollExecStatus(f, client, host, writeResp.ExecutionID, timeout, p)

			default:
				return khhttp.NewAPIError(resp)
			}
		},
	}

	cmd.Flags().String("chain", "", "Chain ID (required)")
	cmd.Flags().String("contract", "", "Contract address (required)")
	cmd.Flags().String("method", "", "Method name (required)")
	cmd.Flags().String("args", "", `Method arguments as JSON array: '["arg1","arg2"]'`)
	cmd.Flags().String("abi-file", "", "Path to local ABI JSON file")
	cmd.Flags().Bool("wait", false, "Wait for completion")
	cmd.Flags().Duration("timeout", 5*time.Minute, "Timeout when using --wait")

	_ = cmd.MarkFlagRequired("chain")
	_ = cmd.MarkFlagRequired("contract")
	_ = cmd.MarkFlagRequired("method")

	return cmd
}

func printContractCallReadResult(p *output.Printer, r *contractCallReadResponse) error {
	return p.PrintData(r, func(tw table.Writer) {
		tw.AppendRow(table.Row{"Result", fmt.Sprintf("%v", r.Result)})
		tw.Render()
	})
}

func printContractCallWriteResult(p *output.Printer, r *contractCallWriteResponse) error {
	return p.PrintData(r, func(tw table.Writer) {
		tw.AppendRow(table.Row{"Execution", r.ExecutionID})
		tw.AppendRow(table.Row{"Status", r.Status})
		if r.TransactionHash != nil && *r.TransactionHash != "" {
			tw.AppendRow(table.Row{"TX Hash", *r.TransactionHash})
		}
		tw.Render()
	})
}
