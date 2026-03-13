package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// BrowserLogin starts a localhost OAuth callback server, opens the browser to
// the server's /cli-auth page which initiates GitHub OAuth from the same origin,
// captures the session token from the callback, stores it in the keyring,
// and returns the token.
//
// A one-time nonce is generated and threaded through the entire flow to prevent
// an attacker from constructing a valid relay URL without knowledge of the nonce.
func BrowserLogin(host string, ios *iostreams.IOStreams) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("starting callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	nonce, err := generateNonce()
	if err != nil {
		_ = listener.Close()
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("nonce") != nonce {
			http.Error(w, "Invalid nonce", http.StatusForbidden)
			errCh <- errors.New("callback nonce mismatch")
			return
		}

		token := r.URL.Query().Get("token")
		if token == "" {
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

	// Open the server-side /cli-auth page. It POSTs to /api/auth/sign-in/social
	// from the same origin (no CORS), using /api/cli-auth/relay as the OAuth
	// callbackURL so the server can read the HttpOnly session cookie and forward
	// the token to our local callback server.
	baseURL := khhttp.BuildBaseURL(host)
	authPageURL := fmt.Sprintf("%s/cli-auth?provider=github&port=%d&nonce=%s", baseURL, port, nonce)

	fmt.Fprintf(ios.Out, "Opening browser to authenticate...\n")

	if err := browserOpener(authPageURL); err != nil {
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
