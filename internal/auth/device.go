package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/iostreams"
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

	codeResp, err := requestDeviceCode(ctx, host)
	if err != nil {
		return "", err
	}

	fmt.Fprintf(ios.Out, "Open this URL to authenticate:\n  %s\n\nEnter code: %s\n\nWaiting for authentication...\n",
		codeResp.VerificationURIComplete, codeResp.UserCode)

	interval := time.Duration(codeResp.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	token, err := pollDeviceToken(ctx, host, codeResp.DeviceCode, interval)
	if err != nil {
		return "", err
	}

	if storeErr := SetToken(host, token); storeErr != nil {
		return "", fmt.Errorf("storing token: %w", storeErr)
	}

	return token, nil
}

func requestDeviceCode(ctx context.Context, host string) (deviceCodeResponse, error) {
	body, err := json.Marshal(map[string]string{"client_id": "kh-cli"})
	if err != nil {
		return deviceCodeResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		khhttp.BuildBaseURL(host)+"/api/auth/device/code", bytes.NewReader(body))
	if err != nil {
		return deviceCodeResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

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

func pollDeviceToken(ctx context.Context, host, deviceCode string, interval time.Duration) (string, error) {
	for {
		select {
		case <-ctx.Done():
			return "", errors.New("timed out waiting for device authorisation")
		case <-time.After(interval):
		}

		tokenResp, err := checkDeviceToken(ctx, host, deviceCode)
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
			return "", errors.New("Device code expired. Run 'kh auth login --no-browser' again.")
		case "access_denied":
			return "", errors.New("Authentication denied.")
		default:
			return "", fmt.Errorf("unexpected device token error: %s", tokenResp.Error)
		}
	}
}

func checkDeviceToken(ctx context.Context, host, deviceCode string) (deviceTokenResponse, error) {
	body, err := json.Marshal(map[string]string{
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
		"device_code": deviceCode,
		"client_id":   "kh-cli",
	})
	if err != nil {
		return deviceTokenResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		khhttp.BuildBaseURL(host)+"/api/auth/device/token", bytes.NewReader(body))
	if err != nil {
		return deviceTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

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
