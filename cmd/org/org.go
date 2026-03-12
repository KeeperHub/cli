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
	}

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewSwitchCmd(f))
	cmd.AddCommand(NewMembersCmd(f))

	return cmd
}
