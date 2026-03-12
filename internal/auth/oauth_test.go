package auth

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

// installBrowserCapture overrides browserOpener so it sends the auth URL on a channel.
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

func extractCallbackPort(authURL string) string {
	// authURL looks like: https://host/api/auth/sign-in/social?provider=github&callbackURL=http://127.0.0.1:PORT/callback
	idx := strings.Index(authURL, "callbackURL=")
	if idx < 0 {
		return ""
	}
	cb := authURL[idx+len("callbackURL="):]
	portIdx := strings.LastIndex(cb, ":")
	slashIdx := strings.Index(cb[portIdx:], "/")
	return cb[portIdx+1 : portIdx+slashIdx]
}

func TestBrowserLogin_CapturesToken(t *testing.T) {
	overrideKeyring(t)
	urlCh := installBrowserCapture(t) // set before goroutine starts

	ios, _, _, _ := iostreams.Test()

	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		tok, err := BrowserLogin("app.keeperhub.io", ios)
		if err != nil {
			errCh <- err
			return
		}
		tokenCh <- tok
	}()

	// Wait for browser opener to be called.
	var authURL string
	select {
	case authURL = <-urlCh:
	case <-time.After(5 * time.Second):
		t.Fatal("browser opener was never called")
	}

	port := extractCallbackPort(authURL)
	require.NotEmpty(t, port)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/callback?token=test_token_123", port))
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

	stored, err := GetToken("app.keeperhub.io")
	require.NoError(t, err)
	require.Equal(t, "test_token_123", stored)
}

func TestBrowserLogin_NoTokenInCallback(t *testing.T) {
	overrideKeyring(t)
	urlCh := installBrowserCapture(t)

	ios, _, _, _ := iostreams.Test()

	errCh := make(chan error, 1)
	go func() {
		_, err := BrowserLogin("app.keeperhub.io", ios)
		errCh <- err
	}()

	var authURL string
	select {
	case authURL = <-urlCh:
	case <-time.After(5 * time.Second):
		t.Fatal("browser opener was never called")
	}

	port := extractCallbackPort(authURL)
	require.NotEmpty(t, port)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/callback", port))
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
