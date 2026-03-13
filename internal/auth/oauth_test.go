package auth

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

// installBrowserCapture overrides browserOpener so it sends the URL on a channel.
// Must be called BEFORE the BrowserLogin goroutine is started.
func installBrowserCapture(t *testing.T) <-chan string {
	t.Helper()
	ch := make(chan string, 1)
	browserOpener = func(url string) error {
		ch <- url
		return nil
	}
	t.Cleanup(func() { browserOpener = openBrowser })
	return ch
}

var portPattern = regexp.MustCompile(`port=(\d+)`)
var noncePattern = regexp.MustCompile(`nonce=([a-f0-9]{32})`)

func extractPort(cliAuthURL string) string {
	m := portPattern.FindStringSubmatch(cliAuthURL)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func extractNonce(cliAuthURL string) string {
	m := noncePattern.FindStringSubmatch(cliAuthURL)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func TestBrowserLogin_CapturesToken(t *testing.T) {
	overrideKeyring(t)
	urlCh := installBrowserCapture(t)

	ios, _, _, _ := iostreams.Test()

	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		tok, err := BrowserLogin("app.keeperhub.com", ios)
		if err != nil {
			errCh <- err
			return
		}
		tokenCh <- tok
	}()

	var cliAuthURL string
	select {
	case cliAuthURL = <-urlCh:
	case <-time.After(5 * time.Second):
		t.Fatal("browser opener was never called")
	}

	require.Contains(t, cliAuthURL, "/cli-auth")
	require.Contains(t, cliAuthURL, "provider=github")

	port := extractPort(cliAuthURL)
	require.NotEmpty(t, port)

	nonce := extractNonce(cliAuthURL)
	require.NotEmpty(t, nonce, "nonce must be present in auth URL")

	// Simulate the relay redirecting to the CLI's local callback server.
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/callback?token=test_token_123&nonce=%s", port, nonce))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	select {
	case tok := <-tokenCh:
		require.Equal(t, "test_token_123", tok)
	case err := <-errCh:
		t.Fatalf("BrowserLogin returned error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for BrowserLogin to return")
	}

	stored, err := GetToken("app.keeperhub.com")
	require.NoError(t, err)
	require.Equal(t, "test_token_123", stored)
}

func TestBrowserLogin_BadNonce(t *testing.T) {
	overrideKeyring(t)
	urlCh := installBrowserCapture(t)

	ios, _, _, _ := iostreams.Test()

	errCh := make(chan error, 1)
	go func() {
		_, err := BrowserLogin("app.keeperhub.com", ios)
		errCh <- err
	}()

	var cliAuthURL string
	select {
	case cliAuthURL = <-urlCh:
	case <-time.After(5 * time.Second):
		t.Fatal("browser opener was never called")
	}

	port := extractPort(cliAuthURL)
	require.NotEmpty(t, port)

	// Send a request with a wrong nonce.
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/callback?token=stolen&nonce=bad", port))
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	select {
	case err := <-errCh:
		require.Error(t, err)
		require.Contains(t, err.Error(), "nonce")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for BrowserLogin to return error")
	}
}

func TestBrowserLogin_NoTokenInCallback(t *testing.T) {
	overrideKeyring(t)
	urlCh := installBrowserCapture(t)

	ios, _, _, _ := iostreams.Test()

	errCh := make(chan error, 1)
	go func() {
		_, err := BrowserLogin("app.keeperhub.com", ios)
		errCh <- err
	}()

	var cliAuthURL string
	select {
	case cliAuthURL = <-urlCh:
	case <-time.After(5 * time.Second):
		t.Fatal("browser opener was never called")
	}

	port := extractPort(cliAuthURL)
	nonce := extractNonce(cliAuthURL)
	require.NotEmpty(t, port)
	require.NotEmpty(t, nonce)

	// Valid nonce but no token.
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/callback?nonce=%s", port, nonce))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	select {
	case err := <-errCh:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for BrowserLogin to return error")
	}
}

func TestReadTokenFromStdin(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.In = strings.NewReader("  my_token_value\n  ")

	tok, err := ReadTokenFromStdin(ios)
	require.NoError(t, err)
	require.Equal(t, "my_token_value", tok)
}

func TestReadTokenFromStdin_Empty(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.In = strings.NewReader("   \n   ")

	_, err := ReadTokenFromStdin(ios)
	require.Error(t, err)
}
