package doctor

import (
	"fmt"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewDoctorCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "doctor",
		Short:   "Check CLI health",
		Aliases: []string{"doc"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "[doctor] is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Structured output for CI")

	return cmd
}
