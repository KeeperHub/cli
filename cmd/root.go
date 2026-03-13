package cmd

import (
	"github.com/keeperhub/cli/cmd/action"
	"github.com/keeperhub/cli/cmd/auth"
	"github.com/keeperhub/cli/cmd/billing"
	"github.com/keeperhub/cli/cmd/completion"
	"github.com/keeperhub/cli/cmd/config"
	"github.com/keeperhub/cli/cmd/doctor"
	"github.com/keeperhub/cli/cmd/execute"
	"github.com/keeperhub/cli/cmd/help"
	"github.com/keeperhub/cli/cmd/org"
	"github.com/keeperhub/cli/cmd/project"
	"github.com/keeperhub/cli/cmd/protocol"
	"github.com/keeperhub/cli/cmd/run"
	"github.com/keeperhub/cli/cmd/serve"
	"github.com/keeperhub/cli/cmd/tag"
	"github.com/keeperhub/cli/cmd/template"
	"github.com/keeperhub/cli/cmd/update"
	"github.com/keeperhub/cli/cmd/version"
	"github.com/keeperhub/cli/cmd/wallet"
	"github.com/keeperhub/cli/cmd/workflow"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root cobra command for the kh CLI.
// All global flags are registered as persistent flags so every subcommand inherits them.
func NewRootCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "kh",
		Short:         "KeeperHub CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")
	cmd.PersistentFlags().BoolP("yes", "y", false, "Skip confirmation prompts")
	cmd.PersistentFlags().Bool("no-color", false, "Disable color output")
	cmd.PersistentFlags().StringP("host", "H", "", "KeeperHub host (default: app.keeperhub.io)")

	cmd.AddCommand(action.NewActionCmd(f))
	cmd.AddCommand(auth.NewAuthCmd(f))
	cmd.AddCommand(billing.NewBillingCmd(f))
	cmd.AddCommand(completion.NewCompletionCmd())
	cmd.AddCommand(config.NewConfigCmd(f))
	cmd.AddCommand(doctor.NewDoctorCmd(f))
	cmd.AddCommand(execute.NewExecuteCmd(f))
	cmd.AddCommand(org.NewOrgCmd(f))
	cmd.AddCommand(project.NewProjectCmd(f))
	cmd.AddCommand(protocol.NewProtocolCmd(f))
	cmd.AddCommand(run.NewRunCmd(f))
	cmd.AddCommand(serve.NewServeCmd(f))
	cmd.AddCommand(tag.NewTagCmd(f))
	cmd.AddCommand(template.NewTemplateCmd(f))
	cmd.AddCommand(update.NewUpdateCmd(f))
	cmd.AddCommand(version.NewVersionCmd(f))
	cmd.AddCommand(wallet.NewWalletCmd(f))
	cmd.AddCommand(workflow.NewWorkflowCmd(f))
	cmd.AddCommand(help.NewEnvironmentTopic())
	cmd.AddCommand(help.NewExitCodesTopic())
	cmd.AddCommand(help.NewFormattingTopic())

	return cmd
}
