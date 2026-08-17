package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/iostreams"
	"golang.org/x/term"
)

type deviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	Interval                int    `json:"interval"`
}

type deviceTokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

// DeviceLogin implements the Device Authorization Grant (RFC 8628).
// It requests a device code, prints the verification URL and user code,
// then polls until the user authorises or the code expires.
func DeviceLogin(host string, ios *iostreams.IOStreams) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// Load per-host headers (e.g. CF-Access) from hosts.yml.
	var hostHeaders map[string]string
	if hostsCfg, hostsErr := config.ReadHosts(); hostsErr == nil {
		if entry, ok := hostsCfg.HostEntry(host); ok {
			hostHeaders = entry.Headers
		}
	}

	baseURL := khhttp.BuildBaseURL(host)
	codeResp, err := requestDeviceCode(ctx, baseURL, hostHeaders)
	if err != nil {
		return "", err
	}

	fmt.Fprintf(ios.Out, "Open this URL to authenticate:\n  %s\n\nEnter code: %s\n\nWaiting for authentication...\n",
		codeResp.VerificationURIComplete, codeResp.UserCode)

	interval := time.Duration(codeResp.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	token, err := pollDeviceToken(ctx, baseURL, codeResp.DeviceCode, interval, hostHeaders)
	if err != nil {
		return "", err
	}

	if storeErr := config.SetHostToken(host, token); storeErr != nil {
		return "", fmt.Errorf("storing token: %w", storeErr)
	}

	return token, nil
}

func requestDeviceCode(ctx context.Context, baseURL string, headers map[string]string) (deviceCodeResponse, error) {
	body, err := json.Marshal(map[string]string{"client_id": "kh-cli"})
	if err != nil {
		return deviceCodeResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		baseURL+"/api/auth/device/code", bytes.NewReader(body))
	if err != nil {
		return deviceCodeResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", baseURL)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return deviceCodeResponse{}, fmt.Errorf("requesting device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return deviceCodeResponse{}, fmt.Errorf("device code request failed: status %d", resp.StatusCode)
	}

	var codeResp deviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&codeResp); err != nil {
		return deviceCodeResponse{}, fmt.Errorf("decoding device code response: %w", err)
	}

	return codeResp, nil
}

func pollDeviceToken(ctx context.Context, baseURL, deviceCode string, interval time.Duration, headers map[string]string) (string, error) {
	for {
		select {
		case <-ctx.Done():
			return "", errors.New("timed out waiting for device authorisation")
		case <-time.After(interval):
		}

		tokenResp, err := checkDeviceToken(ctx, baseURL, deviceCode, headers)
		if err != nil {
			return "", err
		}

		if tokenResp.AccessToken != "" {
			return tokenResp.AccessToken, nil
		}

		switch tokenResp.Error {
		case "authorization_pending":
			// continue polling
		case "slow_down":
			interval += 5 * time.Second
		case "expired_token":
			// --no-browser does not exist; suggesting it sent users looking for
			// a flag the CLI has never had.
			return "", errors.New("device code expired, run 'kh auth login' again")
		case "access_denied":
			return "", errors.New("authentication denied")
		default:
			return "", fmt.Errorf("unexpected device token error: %s", tokenResp.Error)
		}
	}
}

func checkDeviceToken(ctx context.Context, baseURL, deviceCode string, headers map[string]string) (deviceTokenResponse, error) {
	body, err := json.Marshal(map[string]string{
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
		"device_code": deviceCode,
		"client_id":   "kh-cli",
	})
	if err != nil {
		return deviceTokenResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		baseURL+"/api/auth/device/token", bytes.NewReader(body))
	if err != nil {
		return deviceTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", baseURL)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return deviceTokenResponse{}, fmt.Errorf("polling device token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp deviceTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return deviceTokenResponse{}, fmt.Errorf("decoding device token response: %w", err)
	}

	return tokenResp, nil
}

// ReadTokenFromStdin reads an API key without echoing it when stdin is a terminal.
//
// Piped input is unchanged, so `echo $KEY | kh auth login --with-token` and CI usage keep
// working exactly as before. What changes is the interactive case: running the command with no
// pipe used to leave the terminal echoing, so the key appeared on screen and stayed in
// scrollback. An organisation API key is a bearer credential, and a terminal is a bad place to
// leave one lying around.
//
// Echo suppression is verified rather than assumed. term.ReadPassword tells us whether it could
// put the terminal into no-echo mode, and if it cannot, this returns an error instead of falling
// back to a visible read. Failing closed costs the user one message; failing open costs them the
// key.
func ReadTokenFromStdin(ios *iostreams.IOStreams) (string, error) {
	if fd, isTTY := stdinFd(ios); isTTY {
		fmt.Fprint(ios.ErrOut, "KeeperHub organisation API key: ")
		data, err := term.ReadPassword(fd)
		fmt.Fprintln(ios.ErrOut)
		if err != nil {
			return "", fmt.Errorf("reading token from terminal: %w", err)
		}
		token := strings.TrimSpace(string(data))
		if token == "" {
			return "", errors.New("no token provided")
		}
		return token, nil
	}

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

// stdinFd reports the file descriptor for stdin when it is an interactive terminal. A pipe, a
// redirect, a heredoc and a test buffer all report false, which is the distinction that matters:
// any of those means the bytes are not being typed by a person.
func stdinFd(ios *iostreams.IOStreams) (int, bool) {
	f, ok := ios.In.(*os.File)
	if !ok {
		return 0, false
	}
	fd := int(f.Fd())
	return fd, term.IsTerminal(fd)
}
