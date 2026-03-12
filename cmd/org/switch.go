package org

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewSwitchCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "switch <org-slug>",
		Short:   "Switch to an organization",
		Aliases: []string{"sw"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[org switch] is not yet implemented.")
			return nil
		},
	}

	return cmd
}
