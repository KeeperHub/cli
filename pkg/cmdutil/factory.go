package cmdutil

import (
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/iostreams"
)

// Factory is the dependency injection container passed to every command.
// Commands receive a *Factory and call its func fields to obtain dependencies
// lazily; this allows tests to inject mocks without a full runtime setup.
type Factory struct {
	// AppVersion is the binary version string, injected at build time via ldflags.
	AppVersion string

	// Config returns the parsed application configuration.
	Config func() (config.Config, error)

	// HTTPClient returns a configured KeeperHub HTTP client for API requests.
	// The client automatically injects version headers and per-host credentials.
	HTTPClient func() (*khhttp.Client, error)

	// IOStreams provides the standard input/output streams.
	IOStreams *iostreams.IOStreams
}
