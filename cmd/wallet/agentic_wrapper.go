package wallet

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// execCommand is a package-level var so tests can override with a fake builder.
// It wraps stdlib os/exec.Command: argv is a fixed slice, NO shell interpolation.
var execCommand = exec.Command

// lookPath is a package-level var so tests can override the PATH probe.
var lookPath = exec.LookPath

// runNpxWallet invokes `npx @keeperhub/wallet <subcmd> [args...]` via the Go
// stdlib os/exec package. Argv is a fixed slice built from hard-coded strings,
// so there is no command-injection vector (exec.Command is the execFile-equivalent
// in Go -- it does not spawn a shell).
//
// Forwards stdio, injects KEEPERHUB_API_URL derived from --host / hosts.yml, and
// returns any child process exit code as the Cobra command's error.
//
// Does NOT parse the child's stdout -- the npm package is the canonical tool;
// this wrapper treats it as an opaque CLI to avoid reimplementing signing logic.
func runNpxWallet(f *cmdutil.Factory, cmd *cobra.Command, subcmd string, args []string) error {
	if _, err := lookPath("npx"); err != nil {
		return fmt.Errorf("`npx` not found on PATH: install Node.js 20+ from https://nodejs.org and retry (npx ships with npm)")
	}

	cfg, cfgErr := f.Config()
	if cfgErr != nil {
		return fmt.Errorf("reading kh config: %w", cfgErr)
	}
	host := cmdutil.ResolveHost(cmd, cfg)
	baseURL := khhttp.BuildBaseURL(host)

	childArgs := append([]string{"@keeperhub/wallet", subcmd}, args...)
	child := execCommand("npx", childArgs...)
	child.Stdin = f.IOStreams.In
	child.Stdout = f.IOStreams.Out
	child.Stderr = f.IOStreams.ErrOut

	// Base env: inherit from parent so KH_SESSION_COOKIE / HOME / PATH propagate.
	// Append KEEPERHUB_API_URL last so it wins over any pre-existing env entry.
	env := os.Environ()
	env = append(env, "KEEPERHUB_API_URL="+baseURL)
	child.Env = env

	if runErr := child.Run(); runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return fmt.Errorf("npx @keeperhub/wallet %s exited with code %d", subcmd, exitErr.ExitCode())
		}
		return fmt.Errorf("failed to run npx @keeperhub/wallet %s: %w", subcmd, runErr)
	}
	return nil
}
