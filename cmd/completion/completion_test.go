package completion_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/completion"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRootWithCompletion() (*cobra.Command, *bytes.Buffer) {
	var buf bytes.Buffer
	root := &cobra.Command{Use: "kh"}
	root.SetOut(&buf)
	root.AddCommand(completion.NewCompletionCmd())
	return root, &buf
}

func TestBashCompletion(t *testing.T) {
	root, buf := newRootWithCompletion()
	root.SetArgs([]string{"completion", "bash"})
	err := root.Execute()
	require.NoError(t, err)
	out := buf.String()
	assert.NotEmpty(t, out)
	assert.True(t, strings.Contains(out, "bash") || strings.Contains(out, "complete"),
		"bash completion output should contain 'bash' or 'complete'")
}

func TestZshCompletion(t *testing.T) {
	root, buf := newRootWithCompletion()
	root.SetArgs([]string{"completion", "zsh"})
	err := root.Execute()
	require.NoError(t, err)
	out := buf.String()
	assert.NotEmpty(t, out)
	assert.True(t, strings.Contains(out, "#compdef") || strings.Contains(out, "zsh"),
		"zsh completion output should contain '#compdef' or 'zsh'")
}

func TestFishCompletion(t *testing.T) {
	root, buf := newRootWithCompletion()
	root.SetArgs([]string{"completion", "fish"})
	err := root.Execute()
	require.NoError(t, err)
	out := buf.String()
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "complete", "fish completion output should contain 'complete'")
}

func TestPowershellCompletion(t *testing.T) {
	root, buf := newRootWithCompletion()
	root.SetArgs([]string{"completion", "powershell"})
	err := root.Execute()
	require.NoError(t, err)
	out := buf.String()
	assert.NotEmpty(t, out)
}

func TestCompletionNoArgs(t *testing.T) {
	root, _ := newRootWithCompletion()
	root.SetArgs([]string{"completion"})
	err := root.Execute()
	assert.Error(t, err, "completion with no args should return an error")
}

func TestCompletionInvalidShell(t *testing.T) {
	root, _ := newRootWithCompletion()
	root.SetArgs([]string{"completion", "invalid-shell"})
	err := root.Execute()
	assert.Error(t, err, "completion with invalid shell should return an error")
}

func TestCompletionCmdStructure(t *testing.T) {
	cmd := completion.NewCompletionCmd()
	assert.Equal(t, "completion <shell>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.ElementsMatch(t, []string{"bash", "zsh", "fish", "powershell"}, cmd.ValidArgs)
}

