package apikey

import (
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewAPIKeyCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api-key",
		Short:   "Manage API keys",
		Aliases: []string{"ak"},
	}

	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewCreateCmd(f))
	cmd.AddCommand(NewRevokeCmd(f))

	return cmd
}
