package org

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewOrgCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "org",
		Short:   "Manage organizations",
		Aliases: []string{"o"},
		Example: `  # List organizations you belong to
  kh o ls

  # Switch to a different organization
  kh o sw my-org-slug`,
	}

	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewSwitchCmd(f))
	cmd.AddCommand(NewMembersCmd(f))

	return cmd
}
