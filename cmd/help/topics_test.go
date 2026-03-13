package help_test

import (
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/help"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentTopic(t *testing.T) {
	cmd := help.NewEnvironmentTopic()
	require.NotNil(t, cmd)

	assert.Equal(t, "environment", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	long := cmd.Long
	assert.Contains(t, long, "KH_HOST")
	assert.Contains(t, long, "KH_API_KEY")
	assert.Contains(t, long, "KH_CONFIG_DIR")
	assert.Contains(t, long, "XDG_CONFIG_HOME")
	assert.Contains(t, long, "XDG_STATE_HOME")
	assert.Contains(t, long, "XDG_CACHE_HOME")
	assert.Contains(t, long, "NO_COLOR")
}

func TestEnvironmentTopicIsNotRunnable(t *testing.T) {
	cmd := help.NewEnvironmentTopic()
	assert.False(t, cmd.Runnable(), "environment topic should not be runnable (no RunE)")
}

func TestExitCodesTopic(t *testing.T) {
	cmd := help.NewExitCodesTopic()
	require.NotNil(t, cmd)

	assert.Equal(t, "exit-codes", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	long := cmd.Long
	assert.Contains(t, long, "0")
	assert.Contains(t, long, "1")
	assert.Contains(t, long, "2")
	assert.Contains(t, long, "5")
}

func TestExitCodesTopicIsNotRunnable(t *testing.T) {
	cmd := help.NewExitCodesTopic()
	assert.False(t, cmd.Runnable(), "exit-codes topic should not be runnable (no RunE)")
}

func TestFormattingTopic(t *testing.T) {
	cmd := help.NewFormattingTopic()
	require.NotNil(t, cmd)

	assert.Equal(t, "formatting", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	long := cmd.Long
	assert.Contains(t, long, "--json")
	assert.Contains(t, long, "--jq")
	assert.Contains(t, long, "--no-color")
}

func TestFormattingTopicIsNotRunnable(t *testing.T) {
	cmd := help.NewFormattingTopic()
	assert.False(t, cmd.Runnable(), "formatting topic should not be runnable (no RunE)")
}

func TestTopicsAppearInHelp(t *testing.T) {
	root := &cobra.Command{
		Use:           "kh",
		Short:         "KeeperHub CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(help.NewEnvironmentTopic())
	root.AddCommand(help.NewExitCodesTopic())
	root.AddCommand(help.NewFormattingTopic())

	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"--help"})
	_ = root.Execute()

	helpOutput := sb.String()
	assert.Contains(t, helpOutput, "environment", "help output should list environment topic")
	assert.Contains(t, helpOutput, "exit-codes", "help output should list exit-codes topic")
	assert.Contains(t, helpOutput, "formatting", "help output should list formatting topic")
}
