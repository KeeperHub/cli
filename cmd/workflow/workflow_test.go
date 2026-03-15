package workflow_test

import (
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/workflow"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFactory() *cmdutil.Factory {
	ios, _, _, _ := iostreams.Test()
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
	}
}

func TestWorkflowCmdHas8Subcommands(t *testing.T) {
	f := newTestFactory()
	wfCmd := workflow.NewWorkflowCmd(f)
	cmds := wfCmd.Commands()
	assert.Equal(t, 8, len(cmds), "expected 8 subcommands: list, run, get, go-live, pause, create, delete, update")
}

func TestWorkflowCmdHasAlias(t *testing.T) {
	f := newTestFactory()
	wfCmd := workflow.NewWorkflowCmd(f)
	require.Contains(t, wfCmd.Aliases, "wf", "workflow command must have 'wf' alias")
}

func TestWorkflowListHasLsAlias(t *testing.T) {
	f := newTestFactory()
	wfCmd := workflow.NewWorkflowCmd(f)
	for _, c := range wfCmd.Commands() {
		if c.Use == "list" {
			require.Contains(t, c.Aliases, "ls", "list subcommand must have 'ls' alias")
			return
		}
	}
	t.Fatal("list subcommand not found")
}

func TestWorkflowRunHasWaitFlag(t *testing.T) {
	f := newTestFactory()
	wfCmd := workflow.NewWorkflowCmd(f)
	for _, c := range wfCmd.Commands() {
		if strings.HasPrefix(c.Use, "run") {
			flag := c.Flags().Lookup("wait")
			require.NotNil(t, flag, "run subcommand must have --wait flag")
			assert.Equal(t, "bool", flag.Value.Type())
			return
		}
	}
	t.Fatal("run subcommand not found")
}

func TestWorkflowListHasLimitFlag(t *testing.T) {
	f := newTestFactory()
	wfCmd := workflow.NewWorkflowCmd(f)
	for _, c := range wfCmd.Commands() {
		if c.Use == "list" {
			flag := c.Flags().Lookup("limit")
			require.NotNil(t, flag, "list subcommand must have --limit flag")
			assert.Equal(t, "int", flag.Value.Type())
			return
		}
	}
	t.Fatal("list subcommand not found")
}

