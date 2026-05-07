package khhttp_test

import (
	"testing"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/stretchr/testify/assert"
)

func TestMergeCloudflareAccessEnv_NoEnv(t *testing.T) {
	t.Setenv("CF_ACCESS_CLIENT_ID", "")
	t.Setenv("CF_ACCESS_CLIENT_SECRET", "")
	t.Setenv("CF_AUTHORIZATION", "")

	base := map[string]string{"X-Existing": "v"}
	out := khhttp.MergeCloudflareAccessEnv(base)

	assert.Equal(t, base, out, "with no env vars, base should pass through unchanged")
}

func TestMergeCloudflareAccessEnv_ServiceToken(t *testing.T) {
	t.Setenv("CF_ACCESS_CLIENT_ID", "id-123")
	t.Setenv("CF_ACCESS_CLIENT_SECRET", "secret-abc")
	t.Setenv("CF_AUTHORIZATION", "")

	out := khhttp.MergeCloudflareAccessEnv(nil)

	assert.Equal(t, "id-123", out["CF-Access-Client-Id"])
	assert.Equal(t, "secret-abc", out["CF-Access-Client-Secret"])
	assert.NotContains(t, out, "Cookie")
}

func TestMergeCloudflareAccessEnv_PartialServiceTokenIgnored(t *testing.T) {
	// Only ID set, no secret -> nothing added (would 403 anyway and is misleading).
	t.Setenv("CF_ACCESS_CLIENT_ID", "id-only")
	t.Setenv("CF_ACCESS_CLIENT_SECRET", "")
	t.Setenv("CF_AUTHORIZATION", "")

	out := khhttp.MergeCloudflareAccessEnv(nil)

	assert.NotContains(t, out, "CF-Access-Client-Id")
	assert.NotContains(t, out, "CF-Access-Client-Secret")
}

func TestMergeCloudflareAccessEnv_Cookie(t *testing.T) {
	t.Setenv("CF_ACCESS_CLIENT_ID", "")
	t.Setenv("CF_ACCESS_CLIENT_SECRET", "")
	t.Setenv("CF_AUTHORIZATION", "jwt-token")

	out := khhttp.MergeCloudflareAccessEnv(nil)

	assert.Equal(t, "CF_Authorization=jwt-token", out["Cookie"])
}

func TestMergeCloudflareAccessEnv_EnvWinsOverBase(t *testing.T) {
	t.Setenv("CF_ACCESS_CLIENT_ID", "env-id")
	t.Setenv("CF_ACCESS_CLIENT_SECRET", "env-secret")
	t.Setenv("CF_AUTHORIZATION", "")

	base := map[string]string{
		"CF-Access-Client-Id":     "yaml-id",
		"CF-Access-Client-Secret": "yaml-secret",
		"X-Other":                 "preserved",
	}
	out := khhttp.MergeCloudflareAccessEnv(base)

	assert.Equal(t, "env-id", out["CF-Access-Client-Id"])
	assert.Equal(t, "env-secret", out["CF-Access-Client-Secret"])
	assert.Equal(t, "preserved", out["X-Other"])
}

func TestMergeCloudflareAccessEnv_CookieAppendsToExisting(t *testing.T) {
	t.Setenv("CF_ACCESS_CLIENT_ID", "")
	t.Setenv("CF_ACCESS_CLIENT_SECRET", "")
	t.Setenv("CF_AUTHORIZATION", "jwt-token")

	base := map[string]string{"Cookie": "session=abc"}
	out := khhttp.MergeCloudflareAccessEnv(base)

	assert.Equal(t, "session=abc; CF_Authorization=jwt-token", out["Cookie"],
		"existing Cookie value should be preserved and CF_Authorization appended")
}

func TestMergeCloudflareAccessEnv_ServiceTokenAndCookieCombined(t *testing.T) {
	t.Setenv("CF_ACCESS_CLIENT_ID", "id-1")
	t.Setenv("CF_ACCESS_CLIENT_SECRET", "secret-1")
	t.Setenv("CF_AUTHORIZATION", "jwt-1")

	out := khhttp.MergeCloudflareAccessEnv(nil)

	assert.Equal(t, "id-1", out["CF-Access-Client-Id"])
	assert.Equal(t, "secret-1", out["CF-Access-Client-Secret"])
	assert.Equal(t, "CF_Authorization=jwt-1", out["Cookie"])
}

func TestMergeCloudflareAccessEnv_DoesNotMutateBase(t *testing.T) {
	t.Setenv("CF_ACCESS_CLIENT_ID", "id")
	t.Setenv("CF_ACCESS_CLIENT_SECRET", "secret")
	t.Setenv("CF_AUTHORIZATION", "")

	base := map[string]string{"K": "v"}
	_ = khhttp.MergeCloudflareAccessEnv(base)

	assert.Equal(t, map[string]string{"K": "v"}, base, "base map should not be mutated")
}
