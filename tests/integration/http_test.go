//go:build integration

package integration

import (
	"net/http"
	"os"
	"testing"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPClientRetry verifies that the retryable HTTP client can reach the staging
// /api/mcp/schemas endpoint and receive a 200 response.
func TestHTTPClientRetry(t *testing.T) {
	apiKey := os.Getenv("KH_API_KEY")
	if apiKey == "" {
		t.Skip("KH_API_KEY is not set; skipping HTTP integration test")
	}

	host := testHost()
	ios := iostreams.System()

	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       host,
		Token:      apiKey,
		AppVersion: "integration-test",
		IOStreams:   ios,
	})

	url := khhttp.BuildBaseURL(host) + "/api/mcp/schemas"
	req, err := client.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err, "NewRequest should not error")

	resp, err := client.Do(req)
	require.NoError(t, err, "HTTP request should not error")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 from /api/mcp/schemas")
}

// TestHTTPClientVersionHeader verifies that the KH-CLI-Version header is sent on
// every request. The server reflects request headers back via /api/mcp/schemas 200.
func TestHTTPClientVersionHeader(t *testing.T) {
	apiKey := os.Getenv("KH_API_KEY")
	if apiKey == "" {
		t.Skip("KH_API_KEY is not set; skipping HTTP version header integration test")
	}

	host := testHost()
	ios := iostreams.System()
	const testVersion = "test-1.2.3"

	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       host,
		Token:      apiKey,
		AppVersion: testVersion,
		IOStreams:   ios,
	})

	url := khhttp.BuildBaseURL(host) + "/api/mcp/schemas"
	req, err := client.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, testVersion, req.Header.Get("KH-CLI-Version"), "version header should be set on the request")
}
