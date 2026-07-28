package help

import "github.com/spf13/cobra"

// NewEnvironmentTopic returns a non-runnable help topic command for environment variables.
// Cobra displays non-runnable commands under "Additional help topics:" in kh help.
func NewEnvironmentTopic() *cobra.Command {
	return &cobra.Command{
		Use:   "environment",
		Short: "Environment variables used by kh",
		Long: `KH_HOST
  Override the KeeperHub API host. Default: app.keeperhub.com
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

// NewAuthScopeTopic returns a non-runnable help topic command describing what a
// token can and cannot reach, and how to find an integrationId.
//
// Deliberately not named "api-keys": cmd/root_test.go asserts the string
// "api-key" never appears in --help, guarding against the apikey command stubs
// removed in 9acd41d reappearing.
func NewAuthScopeTopic() *cobra.Command {
	return &cobra.Command{
		Use:   "auth-scope",
		Short: "What a token can reach, and what needs a browser session",
		Long: `Most kh commands authenticate with an API key ('kh auth login', keys start
with kh_). A few endpoints do not accept keys at all, which is why some
commands return 401 even though your key is valid and working elsewhere.

Session-only commands
---------------------

  kh org list
  kh org members

These call endpoints that resolve a browser session cookie and never inspect
the Authorization header, so an API key of any scope gets 401. This is a
property of those endpoints, not a problem with your key: if 'kh workflow
list' works, your key is fine. Use the web app for organization membership.

Finding an integrationId
------------------------

Every web3 action node needs an integrationId, and there is no
'kh integration list' command yet. The integrations endpoint does accept API
keys, so query it directly:

  curl -s -H "Authorization: Bearer $KH_TOKEN" \
    https://app.keeperhub.com/api/integrations | jq '.[] | {id, name, type}'

Pick the row with "type": "web3". An organization has at most one.

If you cannot reach that endpoint, the fallback is to read the id off an
existing workflow that already has a working web3 node:

  kh workflow get <id> --json | jq -r '.nodes[].data.config.integrationId | select(.)'

One integrationId covers every chain, EVM and Solana alike - it identifies the
organization's wallet, not a network. Seeing the same id on nodes targeting
different chains is expected. The chain is chosen by each node's own
config.network value.`,
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
      external jq binary is required - the expression is evaluated
      inside kh.
      Example: kh workflow list --json --jq '.[0].id'
      Example: kh auth status --json --jq '.email'

      The result is re-serialized as pretty-printed JSON. This is not
      jq's raw mode: there is no equivalent of jq -r, so strings keep
      their quotes and a filter matching several values prints a JSON
      array, not one value per line.

        $ kh workflow list --json --jq '.[0].id'
        "wf_abc123"

        $ kh workflow list --json --jq '.[].id'
        [
          "wf_abc123",
          "wf_def456"
        ]

      So the obvious shell idiom does not work - it iterates over the
      brackets and the trailing commas, and the next command fails with
      a not-found error naming an id like "wf_abc123," :

        for id in $(kh workflow list --json --jq '.[].id'); do ...

      To get bare values, pipe --json output to a real jq:

        kh workflow list --json | jq -r '.[].id'

  --no-color
      Disable ANSI color codes in output. Color is also disabled
      automatically when stdout is not a terminal (e.g. piped output).
      You can also set the NO_COLOR environment variable.

Pipe detection: when stdout is redirected to a file or another program,
color output is suppressed automatically. JSON output is unchanged by
pipe detection.`,
	}
}
