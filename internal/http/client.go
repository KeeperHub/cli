package khhttp

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/keeperhub/cli/pkg/iostreams"
)

// ClientOptions configures a new Client.
type ClientOptions struct {
	// Host is the base URL for this client (used for display / logging only).
	Host string

	// Token is the Bearer token sent in the Authorization header.
	// If empty, no Authorization header is added.
	Token string

	// Headers are additional per-host headers injected on every request
	// (e.g. Cloudflare Access headers loaded from hosts.yml).
	Headers map[string]string

	// IOStreams provides the ErrOut writer for version warnings.
	IOStreams *iostreams.IOStreams

	// AppVersion is the CLI version string sent in the KH-CLI-Version header.
	AppVersion string
}

// Client is a retryable HTTP client that injects version and auth headers
// on every outgoing request.
type Client struct {
	inner      *retryablehttp.Client
	appVersion string
	token      string
	headers    map[string]string
	ios        *iostreams.IOStreams
}

// NewClient creates a Client wrapping hashicorp/go-retryablehttp with
// 3 retries, exponential backoff between 1s and 30s, and suppressed logging.
func NewClient(opts ClientOptions) *Client {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 3
	rc.RetryWaitMin = 1 * time.Second
	rc.RetryWaitMax = 30 * time.Second
	rc.Logger = nil

	return &Client{
		inner:      rc,
		appVersion: opts.AppVersion,
		token:      opts.Token,
		headers:    opts.Headers,
		ios:        opts.IOStreams,
	}
}

// Do executes a retryablehttp.Request, injecting version and auth headers
// before the first attempt and checking the version compatibility header
// after each successful response.
func (c *Client) Do(req *retryablehttp.Request) (*http.Response, error) {
	req.Header.Set("KH-CLI-Version", c.appVersion)

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.inner.Do(req)
	if err != nil {
		return nil, err
	}

	if c.ios != nil {
		checkVersion(c.appVersion, resp, c.ios.ErrOut)
	}

	return resp, nil
}

// NewRequest is a convenience wrapper around retryablehttp.NewRequest.
func (c *Client) NewRequest(method, url string, body interface{}) (*retryablehttp.Request, error) {
	return retryablehttp.NewRequest(method, url, body)
}

// StandardClient returns the underlying *http.Client for compatibility with
// libraries that require a standard net/http client.
func (c *Client) StandardClient() *http.Client {
	return c.inner.StandardClient()
}
