package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd"
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

func TestNewRootCmdUse(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	assert.Equal(t, "kh", root.Use)
}

func TestNewRootCmdShort(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	assert.Equal(t, "KeeperHub CLI", root.Short)
}

func TestNewRootCmdSilenceErrors(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	assert.True(t, root.SilenceErrors)
}

func TestNewRootCmdSilenceUsage(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	assert.True(t, root.SilenceUsage)
}

func TestNewRootCmdHasJSONFlag(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	flag := root.PersistentFlags().Lookup("json")
	require.NotNil(t, flag)
	assert.Equal(t, "bool", flag.Value.Type())
}

func TestNewRootCmdHasJQFlag(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	flag := root.PersistentFlags().Lookup("jq")
	require.NotNil(t, flag)
	assert.Equal(t, "string", flag.Value.Type())
}

func TestNewRootCmdHasYesFlag(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	flag := root.PersistentFlags().Lookup("yes")
	require.NotNil(t, flag)
	assert.Equal(t, "bool", flag.Value.Type())
}

func TestNewRootCmdHasNoColorFlag(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	flag := root.PersistentFlags().Lookup("no-color")
	require.NotNil(t, flag)
	assert.Equal(t, "bool", flag.Value.Type())
}

func TestNewRootCmdHasHostFlag(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	flag := root.PersistentFlags().Lookup("host")
	require.NotNil(t, flag)
	assert.Equal(t, "string", flag.Value.Type())
}

func TestNewRootCmdHasOrgFlag(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	flag := root.PersistentFlags().Lookup("org")
	require.NotNil(t, flag)
	assert.Equal(t, "string", flag.Value.Type())
}

func TestRootCmdParsesOrgFlag(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	root.SetArgs([]string{"--org", "org_abc123", "--help"})
	err := root.Execute()
	assert.NoError(t, err)

	orgVal, err := root.PersistentFlags().GetString("org")
	require.NoError(t, err)
	assert.Equal(t, "org_abc123", orgVal)
}

func TestHostFlagHasShorthandH(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	flag := root.PersistentFlags().ShorthandLookup("H")
	require.NotNil(t, flag)
	assert.Equal(t, "host", flag.Name)
}

func TestRootCmdNoSubcommandsNoError(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	root.SetArgs([]string{})
	err := root.Execute()
	assert.NoError(t, err)
}

func TestRootCmdParsesHostFlag(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	root.SetArgs([]string{"-H", "app-staging.keeperhub.com", "--help"})
	err := root.Execute()
	assert.NoError(t, err)

	hostVal, err := root.PersistentFlags().GetString("host")
	require.NoError(t, err)
	assert.Equal(t, "app-staging.keeperhub.com", hostVal)
}

func TestRootCmdHas21Subcommands(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	cmds := root.Commands()
	assert.Equal(t, 21, len(cmds), "expected 21 subcommands registered on root (18 commands + 3 help topics)")
}

func TestRootCmdHelpIncludesAllCommands(t *testing.T) {
	var buf bytes.Buffer
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--help"})
	err := root.Execute()
	assert.NoError(t, err)

	helpOutput := buf.String()
	expectedCommands := []string{
		"workflow", "run", "execute", "project",
		"tag", "org", "action", "protocol",
		"wallet", "template", "billing", "doctor",
		"version", "auth", "config", "completion", "update",
	}
	for _, cmdName := range expectedCommands {
		assert.True(t, strings.Contains(helpOutput, cmdName),
			"expected --help output to contain %q", cmdName)
	}
}

func TestRootCmdHelpDoesNotIncludeAPIKey(t *testing.T) {
	var buf bytes.Buffer
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--help"})
	err := root.Execute()
	assert.NoError(t, err)

	helpOutput := buf.String()
	assert.False(t, strings.Contains(helpOutput, "api-key"),
		"api-key should not appear in --help output")
}

func TestAllNounAliasesResolve(t *testing.T) {
	aliases := []string{"wf", "r", "ex", "p", "t", "o", "a", "pr", "w", "tp", "b", "doc", "v"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			f := newTestFactory()
			root := cmd.NewRootCmd(f)
			root.SetArgs([]string{alias, "--help"})
			err := root.Execute()
			assert.NoError(t, err, "alias %q should resolve without error", alias)
		})
	}
}

func TestAKAliasDoesNotResolve(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	root.SetArgs([]string{"ak", "--help"})
	err := root.Execute()
	assert.Error(t, err, "alias 'ak' should not resolve after apikey removal")
}

func TestAllSubcommandsHaveShortDescription(t *testing.T) {
	f := newTestFactory()
	root := cmd.NewRootCmd(f)
	for _, subCmd := range root.Commands() {
		assert.NotEmpty(t, subCmd.Short,
			"command %q must have a non-empty Short description", subCmd.Use)
	}
}
