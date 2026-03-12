package khhttp_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientSetsVersionHeader(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("KH-CLI-Version")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "1.2.3",
		IOStreams:   ios,
	})

	req, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "1.2.3", gotHeader)
}

func TestClientSetsCustomHeaders(t *testing.T) {
	var gotCFHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCFHeader = r.Header.Get("CF-Access-Client-Id")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "1.0.0",
		Headers:    map[string]string{"CF-Access-Client-Id": "my-client-id"},
		IOStreams:   ios,
	})

	req, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "my-client-id", gotCFHeader)
}

func TestClientSetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "1.0.0",
		Token:      "my-secret-token",
		IOStreams:   ios,
	})

	req, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "Bearer my-secret-token", gotAuth)
}

func TestClientNoAuthorizationHeaderWhenTokenEmpty(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "1.0.0",
		Token:      "",
		IOStreams:   ios,
	})

	req, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "", gotAuth)
}

func TestCheckVersionWritesWarningWhenOutdated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("KH-Minimum-CLI-Version", "2.0.0")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, errOut, _ := iostreams.Test()
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "1.0.0",
		IOStreams:   ios,
	})

	req, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Contains(t, errOut.String(), "outdated")
}

func TestCheckVersionNoWarningWhenHeaderAbsent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, errOut, _ := iostreams.Test()
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "1.0.0",
		IOStreams:   ios,
	})

	req, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Empty(t, errOut.String())
}

func TestCheckVersionNoWarningWhenCurrentMeetsMinimum(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("KH-Minimum-CLI-Version", "1.0.0")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, errOut, _ := iostreams.Test()
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "1.0.0",
		IOStreams:   ios,
	})

	req, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Empty(t, errOut.String())
}

func TestCheckVersionNoWarningForDevVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("KH-Minimum-CLI-Version", "99.0.0")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, errOut, _ := iostreams.Test()
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: "dev",
		IOStreams:   ios,
	})

	req, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Empty(t, errOut.String())
}

func TestSemverLessThan(t *testing.T) {
	tests := []struct {
		current  string
		minimum  string
		expected bool
	}{
		{"0.1.0", "1.0.0", true},
		{"1.0.0", "0.9.0", false},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.1.0", true},
		{"1.1.0", "1.0.0", false},
		{"dev", "1.0.0", false},
		{"dev", "99.0.0", false},
		{"unparseable", "1.0.0", false},
		{"2.0.0", "1.9.9", false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tc.current, tc.minimum), func(t *testing.T) {
			result := khhttp.SemverLessThan(tc.current, tc.minimum)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAPIErrorFormatsMessage(t *testing.T) {
	body := io.NopCloser(strings.NewReader(`{"error":"not found"}`))
	resp := &http.Response{
		StatusCode: 404,
		Body:       body,
	}

	apiErr := khhttp.NewAPIError(resp)
	assert.Equal(t, 404, apiErr.StatusCode)
	assert.Contains(t, apiErr.Error(), "404")
	assert.Contains(t, apiErr.Error(), "not found")
}

func TestAPIErrorNonJSONBody(t *testing.T) {
	body := io.NopCloser(strings.NewReader("plain text error"))
	resp := &http.Response{
		StatusCode: 500,
		Body:       body,
	}

	apiErr := khhttp.NewAPIError(resp)
	assert.Equal(t, 500, apiErr.StatusCode)
	assert.Contains(t, apiErr.Error(), "500")
}
