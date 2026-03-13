package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/iostreams"
)

// browserOpener opens a URL in the default browser. Tests override this.
var browserOpener = openBrowser

// socialSignInURLFunc fetches the OAuth redirect URL from the server. Tests override this.
var socialSignInURLFunc = fetchSocialSignInURL

func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}

// BrowserLogin starts a localhost OAuth callback server, opens the browser to
// authenticate via GitHub OAuth, captures the session token from the callback,
// stores it in the keyring, and returns the token.
func BrowserLogin(host string, ios *iostreams.IOStreams) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("starting callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			// Try cookie fallback
			if cookie, cookieErr := r.Cookie("better-auth.session_token"); cookieErr == nil {
				token = cookie.Value
			}
		}
		if token == "" {
			http.Error(w, "No token received", http.StatusBadRequest)
			errCh <- errors.New("no token in callback")
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>Authentication successful!</h1><p>You can close this tab.</p></body></html>")
		tokenCh <- token
	})

	go func() {
		if serveErr := srv.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
	}()

	// Better Auth's /sign-in/social is a POST endpoint that returns the OAuth redirect URL.
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	baseURL := khhttp.BuildBaseURL(host)

	authURL, postErr := socialSignInURLFunc(baseURL, "github", callbackURL)
	if postErr != nil {
		_ = srv.Close()
		return "", fmt.Errorf("initiating social sign-in: %w", postErr)
	}

	fmt.Fprintf(ios.Out, "Opening browser to authenticate...\n")

	if err := browserOpener(authURL); err != nil {
		_ = srv.Close()
		return "", fmt.Errorf("opening browser: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var token string
	select {
	case token = <-tokenCh:
	case err = <-errCh:
		_ = srv.Close()
		return "", err
	case <-ctx.Done():
		_ = srv.Close()
		return "", errors.New("timed out waiting for browser authentication")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)

	if storeErr := SetToken(host, token); storeErr != nil {
		return "", fmt.Errorf("storing token: %w", storeErr)
	}

	return token, nil
}

// fetchSocialSignInURL POSTs to Better Auth's /sign-in/social endpoint and
// returns the OAuth provider redirect URL from the JSON response.
func fetchSocialSignInURL(baseURL, provider, callbackURL string) (string, error) {
	body, err := json.Marshal(map[string]string{
		"provider":    provider,
		"callbackURL": callbackURL,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(baseURL+"/api/auth/sign-in/social", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		URL      string `json:"url"`
		Redirect bool   `json:"redirect"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	if result.URL == "" {
		return "", errors.New("server returned empty redirect URL")
	}
	return result.URL, nil
}

// ReadTokenFromStdin reads a token from ios.In, trims whitespace, and returns it.
// Returns an error if the input is empty after trimming.
func ReadTokenFromStdin(ios *iostreams.IOStreams) (string, error) {
	data, err := io.ReadAll(ios.In)
	if err != nil {
		return "", fmt.Errorf("reading token from stdin: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", errors.New("no token provided on stdin")
	}
	return token, nil
}
