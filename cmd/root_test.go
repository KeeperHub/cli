package cmd_test

import (
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
