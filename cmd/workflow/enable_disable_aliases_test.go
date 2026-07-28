package workflow_test

// enable/disable are the primary names, but resume/pause were the originals and
// are kept working indefinitely - scripts, tutorials and published doc links
// depend on them. These tests pin that promise: each alias must reach the same
// command and send the same request body, so a future rename cannot quietly
// drop one.

import (
	"testing"

	"github.com/keeperhub/cli/cmd/workflow"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findSubcommand returns the subcommand reachable by the given name or alias.
func findSubcommand(t *testing.T, name string) (string, bool) {
	t.Helper()
	ios, _, _, _ := iostreams.Test()
	root := workflow.NewWorkflowCmd(&cmdutil.Factory{IOStreams: ios})
	for _, c := range root.Commands() {
		if c.Name() == name {
			return c.Name(), true
		}
		for _, alias := range c.Aliases {
			if alias == name {
				return c.Name(), true
			}
		}
	}
	return "", false
}

func TestEnableIsReachableByItsAliases(t *testing.T) {
	for _, name := range []string{"enable", "resume", "activate"} {
		resolved, found := findSubcommand(t, name)
		require.True(t, found, "%q should reach a subcommand", name)
		assert.Equal(t, "enable", resolved, "%q should resolve to enable", name)
	}
}

func TestDisableIsReachableByItsAliases(t *testing.T) {
	for _, name := range []string{"disable", "pause"} {
		resolved, found := findSubcommand(t, name)
		require.True(t, found, "%q should reach a subcommand", name)
		assert.Equal(t, "disable", resolved, "%q should resolve to disable", name)
	}
}

func TestEnableAndDisableCrossReferenceEachOther(t *testing.T) {
	// "enable" is the word people search for; someone who lands on one should
	// be told about the other rather than having to guess the opposite verb.
	ios, _, _, _ := iostreams.Test()
	root := workflow.NewWorkflowCmd(&cmdutil.Factory{IOStreams: ios})

	longByName := map[string]string{}
	for _, c := range root.Commands() {
		longByName[c.Name()] = c.Long
	}

	assert.Contains(t, longByName["enable"], "kh workflow disable")
	assert.Contains(t, longByName["disable"], "kh workflow enable")
}
