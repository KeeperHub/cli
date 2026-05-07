package wallet

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewFeedbackCmd returns the `kh wallet feedback` subcommand -- a thin wrapper
// around `npx @keeperhub/wallet feedback` that submits ERC-8004 ReputationRegistry
// feedback for a workflow execution this wallet paid for.
//
// Signing + broadcast happens server-side in the npm package; the wallet's
// Turnkey policy gates what can be signed (only giveFeedback() on the
// ReputationRegistry on Ethereum mainnet). Caller wallet pays gas natively.
func NewFeedbackCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		executionID string
		value       string
		decimals    string
		comment     string
		agentID     string
		chainID     string
	)
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Submit ERC-8004 feedback for a workflow execution this wallet paid for",
		Long: `Submit ERC-8004 ReputationRegistry feedback for a workflow execution this
wallet paid for. Signs giveFeedback() via Turnkey and broadcasts on Ethereum
mainnet via the KeeperHub server proxy. Caller wallet pays gas natively
(~$0.05-2 per call at typical mainnet gas).

Thin wrapper around ` + "`npx @keeperhub/wallet feedback`" + `. Defaults to rating
KeeperHub's own ERC-8004 agent (id 31875 on Ethereum); use --agent-id to rate
any other agent.`,
		Example: `  # 5-star rating for an execution this wallet paid for
  kh w feedback --execution-id c5ybokpmwxi7kiau5wxja --value 5

  # 4.5-star rating with a comment
  kh w feedback --execution-id c5ybokpmwxi7kiau5wxja --value 45 --decimals 1 --comment "very helpful"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			args := []string{"--execution-id", executionID, "--value", value}
			if decimals != "" {
				args = append(args, "--decimals", decimals)
			}
			if comment != "" {
				args = append(args, "--comment", comment)
			}
			if agentID != "" {
				args = append(args, "--agent-id", agentID)
			}
			if chainID != "" {
				args = append(args, "--chain-id", chainID)
			}
			return runNpxWallet(f, cmd, "feedback", args)
		},
	}
	cmd.Flags().StringVar(&executionID, "execution-id", "", "workflow execution id (from a previous call_workflow response) (required)")
	cmd.Flags().StringVar(&value, "value", "", "raw int128 rating value, e.g. 5 with --decimals 0 for a 5-star rating (required)")
	cmd.Flags().StringVar(&decimals, "decimals", "", "decimals for value (0..18); 0 for integer scores, 1 for 0.1-step (default 0)")
	cmd.Flags().StringVar(&comment, "comment", "", "optional plaintext comment included in the feedbackURI JSON")
	cmd.Flags().StringVar(&agentID, "agent-id", "", "rated agent NFT id (uint256 decimal); defaults to KeeperHub agent 31875")
	cmd.Flags().StringVar(&chainID, "chain-id", "", "agent chain id; defaults to 1 (Ethereum mainnet)")
	_ = cmd.MarkFlagRequired("execution-id")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}
