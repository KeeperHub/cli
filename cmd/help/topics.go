package help

import "github.com/spf13/cobra"

// NewEnvironmentTopic returns a non-runnable help topic command for environment variables.
// Cobra displays non-runnable commands under "Additional help topics:" in kh help.
func NewEnvironmentTopic() *cobra.Command {
	return &cobra.Command{
		Use:   "environment",
		Short: "Environment variables used by kh",
		Long: `KH_HOST
  Override the KeeperHub API host. Default: app.keeperhub.io
  Example: KH_HOST=https://kh.mycompany.io kh workflow list

KH_API_KEY
  API key for non-interactive authentication. Overrides any stored
  credentials from kh auth login.
  Example: KH_API_KEY=sk_live_... kh workflow list

KH_CONFIG_DIR
  Override the configuration directory where kh stores config.yml and
  hosts.yml. Default: $XDG_CONFIG_HOME/kh or ~/.config/kh

XDG_CONFIG_HOME
  Base directory for configuration files when KH_CONFIG_DIR is not set.
  Default: ~/.config

XDG_STATE_HOME
  Base directory for state files (e.g. device auth state).
  Default: ~/.local/state

XDG_CACHE_HOME
  Base directory for cache files (e.g. protocol schema cache).
  Default: ~/.cache

NO_COLOR
  Disable color output. Also available as the --no-color flag.
  Example: NO_COLOR=1 kh workflow list`,
	}
}

// NewExitCodesTopic returns a non-runnable help topic command for exit codes.
func NewExitCodesTopic() *cobra.Command {
	return &cobra.Command{
		Use:   "exit-codes",
		Short: "Exit codes returned by kh",
		Long: `kh commands exit with one of the following codes:

  0   Success. The command completed without errors.

  1   General error. An unexpected error occurred (network failure,
      server error, auth failure, invalid config).

  2   Resource not found or invalid argument. The requested resource
      does not exist or a required flag was missing or malformed.

  5   Rate limit exceeded. The API rate limit was hit; retry after
      a short delay.`,
	}
}

// NewFormattingTopic returns a non-runnable help topic command for output formatting.
func NewFormattingTopic() *cobra.Command {
	return &cobra.Command{
		Use:   "formatting",
		Short: "Output formatting options",
		Long: `By default, kh commands display output as a table formatted for terminal
readability. You can change the output format with the following flags:

  --json
      Output the raw API response as machine-readable JSON. Useful for
      scripting and piping to other tools.
      Example: kh workflow list --json

  --jq EXPR
      Filter or transform the JSON output using a jq expression. No
      external jq binary is required -- the expression is evaluated
      inside kh.
      Example: kh workflow list --json --jq '.[0].id'
      Example: kh auth status --json --jq '.email'

  --no-color
      Disable ANSI color codes in output. Color is also disabled
      automatically when stdout is not a terminal (e.g. piped output).
      You can also set the NO_COLOR environment variable.

Pipe detection: when stdout is redirected to a file or another program,
color output is suppressed automatically. JSON output is unchanged by
pipe detection.`,
	}
}
