package tag

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCreateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a tag",
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[tag create] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().String("name", "", "Tag name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().String("color", "", "Tag color")

	return cmd
}
