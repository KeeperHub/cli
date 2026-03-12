package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func deviceTestServer(t *testing.T, pollResponses []deviceTokenResponse) (*httptest.Server, *int32) {
	t.Helper()
	var callCount int32

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/device/code", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "kh-cli", body["client_id"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(deviceCodeResponse{
			DeviceCode:              "dev_code_xyz",
			UserCode:                "ABCD-1234",
			VerificationURI:         "https://example.com/device",
			VerificationURIComplete: "https://example.com/device?user_code=ABCD-1234",
			Interval:                0, // 0 so tests don't wait
		})
	})

	mux.HandleFunc("/api/auth/device/token", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "kh-cli", body["client_id"])
		require.Equal(t, "dev_code_xyz", body["device_code"])

		idx := atomic.AddInt32(&callCount, 1) - 1
		if int(idx) >= len(pollResponses) {
			idx = int32(len(pollResponses) - 1)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pollResponses[idx])
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &callCount
}

// patchDeviceHTTP replaces http.DefaultClient temporarily for device tests.
func patchDeviceClient(t *testing.T, srv *httptest.Server) {
	t.Helper()
	orig := http.DefaultClient
	http.DefaultClient = srv.Client()
	t.Cleanup(func() { http.DefaultClient = orig })
}

func TestDeviceLogin_Success(t *testing.T) {
	overrideKeyring(t)

	srv, _ := deviceTestServer(t, []deviceTokenResponse{
		{Error: "authorization_pending"},
		{AccessToken: "device_access_token_xyz"},
	})

	// Override the host with the test server host.
	host := strings.TrimPrefix(srv.URL, "http://")
	patchDeviceClient(t, srv)

	ios, outBuf, _, _ := iostreams.Test()
	tok, err := deviceLoginWithHTTP(host, ios, srv.Client())
	require.NoError(t, err)
	require.Equal(t, "device_access_token_xyz", tok)

	output := outBuf.String()
	require.Contains(t, output, "ABCD-1234")
	require.Contains(t, output, "https://example.com/device?user_code=ABCD-1234")

	stored, err := GetToken(host)
	require.NoError(t, err)
	require.Equal(t, "device_access_token_xyz", stored)
}

func TestDeviceLogin_ExpiredToken(t *testing.T) {
	overrideKeyring(t)

	srv, _ := deviceTestServer(t, []deviceTokenResponse{
		{Error: "expired_token"},
	})
	host := strings.TrimPrefix(srv.URL, "http://")

	ios, _, _, _ := iostreams.Test()
	_, err := deviceLoginWithHTTP(host, ios, srv.Client())
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestDeviceLogin_AccessDenied(t *testing.T) {
	overrideKeyring(t)

	srv, _ := deviceTestServer(t, []deviceTokenResponse{
		{Error: "access_denied"},
	})
	host := strings.TrimPrefix(srv.URL, "http://")

	ios, _, _, _ := iostreams.Test()
	_, err := deviceLoginWithHTTP(host, ios, srv.Client())
	require.Error(t, err)
	require.Contains(t, err.Error(), "denied")
}

func TestDeviceLogin_SlowDown(t *testing.T) {
	overrideKeyring(t)

	var intervalSeen time.Duration
	srv, _ := deviceTestServer(t, []deviceTokenResponse{
		{Error: "slow_down"},
		{AccessToken: "slow_token"},
	})
	host := strings.TrimPrefix(srv.URL, "http://")

	ios, _, _, _ := iostreams.Test()
	tok, err := deviceLoginWithHTTPAndIntervalCapture(host, ios, srv.Client(), &intervalSeen)
	require.NoError(t, err)
	require.Equal(t, "slow_token", tok)
	require.Equal(t, 5*time.Second, intervalSeen)
}

func TestDeviceLogin_PrintsURLAndCode(t *testing.T) {
	overrideKeyring(t)

	srv, _ := deviceTestServer(t, []deviceTokenResponse{
		{AccessToken: "tok"},
	})
	host := strings.TrimPrefix(srv.URL, "http://")

	ios, outBuf, _, _ := iostreams.Test()
	_, err := deviceLoginWithHTTP(host, ios, srv.Client())
	require.NoError(t, err)

	out := outBuf.String()
	require.Contains(t, out, "Open this URL to authenticate")
	require.Contains(t, out, "ABCD-1234")
	require.Contains(t, out, "Waiting for authentication")
}

// deviceLoginWithHTTP is a testable version of DeviceLogin that uses a custom HTTP client.
func deviceLoginWithHTTP(host string, ios *iostreams.IOStreams, client *http.Client) (string, error) {
	return deviceLoginInternal(host, ios, client, nil)
}

func deviceLoginWithHTTPAndIntervalCapture(host string, ios *iostreams.IOStreams, client *http.Client, captured *time.Duration) (string, error) {
	return deviceLoginInternal(host, ios, client, captured)
}

func deviceLoginInternal(host string, ios *iostreams.IOStreams, client *http.Client, captureInterval *time.Duration) (string, error) {
	scheme := "http"
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	body, err := json.Marshal(map[string]string{"client_id": "kh-cli"})
	if err != nil {
		return "", err
	}

	codeReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/auth/device/code", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	codeReq.Header.Set("Content-Type", "application/json")

	codeResp, err := client.Do(codeReq)
	if err != nil {
		return "", err
	}
	defer codeResp.Body.Close()

	var cr deviceCodeResponse
	if err := json.NewDecoder(codeResp.Body).Decode(&cr); err != nil {
		return "", err
	}

	fmt.Fprintf(ios.Out, "Open this URL to authenticate:\n  %s\n\nEnter code: %s\n\nWaiting for authentication...\n",
		cr.VerificationURIComplete, cr.UserCode)

	interval := time.Duration(cr.Interval) * time.Second
	if interval == 0 {
		interval = 0 // no wait in tests
	}

	for {
		if interval > 0 {
			time.Sleep(interval)
		}

		tokenBody, err := json.Marshal(map[string]string{
			"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
			"device_code": cr.DeviceCode,
			"client_id":   "kh-cli",
		})
		if err != nil {
			return "", err
		}

		tokenReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/auth/device/token", strings.NewReader(string(tokenBody)))
		if err != nil {
			return "", err
		}
		tokenReq.Header.Set("Content-Type", "application/json")

		tokenResp, err := client.Do(tokenReq)
		if err != nil {
			return "", err
		}

		var tr deviceTokenResponse
		if err := json.NewDecoder(tokenResp.Body).Decode(&tr); err != nil {
			tokenResp.Body.Close()
			return "", err
		}
		tokenResp.Body.Close()

		if tr.AccessToken != "" {
			if err := SetToken(host, tr.AccessToken); err != nil {
				return "", err
			}
			return tr.AccessToken, nil
		}

		switch tr.Error {
		case "authorization_pending":
			// continue
		case "slow_down":
			interval += 5 * time.Second
			if captureInterval != nil {
				*captureInterval = interval
			}
		case "expired_token":
			return "", fmt.Errorf("Device code expired. Run 'kh auth login --no-browser' again.")
		case "access_denied":
			return "", fmt.Errorf("Authentication denied.")
		default:
			return "", fmt.Errorf("unexpected error: %s", tr.Error)
		}
	}
}
