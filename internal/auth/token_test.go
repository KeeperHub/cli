package auth

// API-key validation used to probe /api/workflows, which resolves auth with
// required:false and answers 200 with an empty list to anonymous callers. Every
// string beginning with kh_ therefore validated, including revoked and
// fabricated keys - and `kh auth login` reads this to decide whether a stored
// credential is still good, so a revoked key made the login skip unrecoverable
// without --force. These tests pin that an unaccepted key is now rejected.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchTokenInfo_RejectsAPIKeyTheServerDoesNotAccept(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := FetchTokenInfo(srv.URL, "kh_revokedOrFabricated")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or revoked")
}

func TestFetchTokenInfo_AcceptsAPIKeyTheServerAccepts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	info, err := FetchTokenInfo(srv.URL, "kh_valid")

	require.NoError(t, err)
	assert.Equal(t, AuthMethodAPIKey, info.Method)
	assert.Equal(t, "api-key", info.Role)
}

func TestFetchTokenInfo_ProbesAnAuthRequiredEndpoint(t *testing.T) {
	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	_, err := FetchTokenInfo(srv.URL, "kh_valid")
	require.NoError(t, err)

	// Guards against regressing onto an endpoint that tolerates anonymous
	// callers - the defect this replaced.
	assert.Equal(t, khhttp.CredentialProbePath, gotPath)
	assert.NotEqual(t, "/api/workflows", gotPath)
	assert.Equal(t, "Bearer kh_valid", gotAuth)
}

func TestFetchTokenInfo_TreatsOtherFailuresAsInvalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := FetchTokenInfo(srv.URL, "kh_valid")

	require.Error(t, err)
}
