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

func TestWorkflowCmdHas5Subcommands(t *testing.T) {
	f := newTestFactory()
	wfCmd := workflow.NewWorkflowCmd(f)
	cmds := wfCmd.Commands()
	assert.Equal(t, 5, len(cmds), "expected 5 subcommands: list, run, get, go-live, pause")
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

func TestWorkflowListNotImplementedViaLsAlias(t *testing.T) {
	ios, outBuf, _, _ := iostreams.Test()
	f := &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
	}

	wfCmd := workflow.NewWorkflowCmd(f)
	wfCmd.SetArgs([]string{"ls"})
	err := wfCmd.Execute()
	assert.NoError(t, err)
	assert.True(t, strings.Contains(outBuf.String(), "not yet implemented"),
		"expected 'not yet implemented' in output, got: %q", outBuf.String())
}
