package wallet

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewWalletCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "wallet",
		Short:   "Manage wallets (creator-wallet REST API or agentic-wallet npm package)",
		Aliases: []string{"w"},
		Long: `Manage wallets.

Creator wallet (REST):
  kh w balance    show creator-wallet on-chain balances via KeeperHub REST API
  kh w tokens     list supported tokens

Agentic wallet (thin wrappers around npx @keeperhub/wallet):
  kh w add        provision a new agentic wallet (no account required)
  kh w info       print agentic subOrgId + walletAddress
  kh w fund       print Coinbase Onramp URL + Tempo deposit address
  kh w link       link agentic wallet to a KeeperHub account (needs KH_SESSION_COOKIE)
  kh w feedback   submit ERC-8004 feedback for a workflow execution this wallet paid for`,
		Example: `  # Creator wallet balance (REST):
  kh w balance

  # Provision an agentic wallet (npx wrapper):
  kh w add

  # Check balance on the agentic wallet:
  npx @keeperhub/wallet balance`,
	}

	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")

	// Creator wallet (REST) -- unchanged from pre-35 kh.
	cmd.AddCommand(NewBalanceCmd(f))
	cmd.AddCommand(NewTokensCmd(f))

	// Agentic wallet (npx @keeperhub/wallet) -- new in phase 35.
	cmd.AddCommand(NewAddCmd(f))
	cmd.AddCommand(NewInfoCmd(f))
	cmd.AddCommand(NewFundCmd(f))
	cmd.AddCommand(NewLinkCmd(f))
	cmd.AddCommand(NewFeedbackCmd(f))

	return cmd
}
