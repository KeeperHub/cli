package apikey

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCreateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create an API key",
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[api-key create] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().String("name", "", "API key name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().String("expires", "", "Expiry duration (e.g. 30d)")

	return cmd
}
