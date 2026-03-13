//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/keeperhub/cli/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stagingHost = "app.keeperhub.io"

// TestAuthLoginFlow verifies that a token can be resolved for the staging host.
// Skips if KH_TEST_EMAIL and KH_TEST_PASSWORD are not set.
func TestAuthLoginFlow(t *testing.T) {
	email := os.Getenv("KH_TEST_EMAIL")
	password := os.Getenv("KH_TEST_PASSWORD")
	if email == "" || password == "" {
		t.Skip("KH_TEST_EMAIL and KH_TEST_PASSWORD are not set; skipping auth integration test")
	}

	host := testHost()

	resolved, err := auth.ResolveToken(host)
	require.NoError(t, err, "ResolveToken should not error")

	if resolved.Method == auth.AuthMethodNone {
		t.Skip("no token configured for host; skipping (run kh auth login or set KH_API_KEY)")
	}

	assert.NotEmpty(t, resolved.Token, "resolved token should not be empty")
	assert.Equal(t, host, resolved.Host, "resolved host should match requested host")
}

// TestAuthStatus verifies that a valid token can fetch session info from the staging API.
// Skips if KH_API_KEY is not set.
func TestAuthStatus(t *testing.T) {
	apiKey := os.Getenv("KH_API_KEY")
	if apiKey == "" {
		t.Skip("KH_API_KEY is not set; skipping auth status integration test")
	}

	host := testHost()

	info, err := auth.FetchTokenInfo(host, apiKey)
	require.NoError(t, err, "FetchTokenInfo should not error with a valid API key")

	assert.NotEmpty(t, info.UserID, "session should have a user ID")
	assert.NotEmpty(t, info.Email, "session should have an email address")
}

func testHost() string {
	if h := os.Getenv("KH_TEST_HOST"); h != "" {
		return h
	}
	return stagingHost
}
