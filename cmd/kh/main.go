package main

import (
	"fmt"
	"os"

	"github.com/keeperhub/cli/cmd/auth"
	"github.com/keeperhub/cli/cmd/completion"
	"github.com/keeperhub/cli/cmd/config"
	"github.com/keeperhub/cli/cmd/doctor"
	"github.com/keeperhub/cli/cmd/execute"
	cmdrun "github.com/keeperhub/cli/cmd/run"
	"github.com/keeperhub/cli/cmd/serve"
	cmdversion "github.com/keeperhub/cli/cmd/version"
	"github.com/keeperhub/cli/cmd/workflow"
	internalconfig "github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/internal/version"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

func main() {
	streams := iostreams.System()

	f := &cmdutil.Factory{
		AppVersion: version.Version,
		IOStreams:   streams,
		Config: func() (internalconfig.Config, error) {
			return internalconfig.ReadConfig()
		},
	}

	rootCmd := &cobra.Command{
		Use:   "kh",
		Short: "KeeperHub CLI",
		Long:  "KeeperHub CLI - manage workflows, runs, and blockchain actions.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(workflow.NewWorkflowCmd(f))
	rootCmd.AddCommand(cmdrun.NewRunCmd(f))
	rootCmd.AddCommand(execute.NewExecuteCmd(f))
	rootCmd.AddCommand(auth.NewAuthCmd(f))
	rootCmd.AddCommand(config.NewConfigCmd(f))
	rootCmd.AddCommand(serve.NewServeCmd(f))
	rootCmd.AddCommand(cmdversion.NewVersionCmd(f))
	rootCmd.AddCommand(doctor.NewDoctorCmd(f))
	rootCmd.AddCommand(completion.NewCompletionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(streams.ErrOut, "Error: %s\n", err)
		os.Exit(1)
	}
}
