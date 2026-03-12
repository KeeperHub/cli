package version

import (
	"fmt"
	"runtime"

	"github.com/keeperhub/cli/internal/version"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewVersionCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Show CLI version",
		Aliases: []string{"v"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(f.IOStreams.Out, "kh version %s\n", version.Version)
			fmt.Fprintf(f.IOStreams.Out, "%s/%s (%s)\n", runtime.GOOS, runtime.GOARCH, runtime.Version())
			return nil
		},
	}

	return cmd
}
