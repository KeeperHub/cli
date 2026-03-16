package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/99designs/keyring"
	"github.com/stretchr/testify/require"
)

const testHost = "app.keeperhub.com"

func setupEmptyKeyring(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
		FileDir:          filepath.Join(dir, "keyring"),
		FilePasswordFunc: keyring.FixedStringPrompt(""),
	})
	require.NoError(t, err)
	original := openKeyringFunc
	openKeyringFunc = func() (keyring.Keyring, error) { return kr, nil }
	t.Cleanup(func() { openKeyringFunc = original })
}

func setupKeyringWithToken(t *testing.T, host, token string) {
	t.Helper()
	dir := t.TempDir()
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
		FileDir:          filepath.Join(dir, "keyring"),
		FilePasswordFunc: keyring.FixedStringPrompt(""),
	})
	require.NoError(t, err)
	require.NoError(t, kr.Set(keyring.Item{Key: host, Data: []byte(token)}))
	original := openKeyringFunc
	openKeyringFunc = func() (keyring.Keyring, error) { return kr, nil }
	t.Cleanup(func() { openKeyringFunc = original })
}

func writeHostsFile(t *testing.T, host, token string) {
	t.Helper()
	dir := t.TempDir()
	khDir := filepath.Join(dir, "kh")
	require.NoError(t, os.MkdirAll(khDir, 0o700))
	hostsFile := filepath.Join(khDir, "hosts.yml")
	content := "hosts:\n  " + host + ":\n    token: " + token + "\n"
	require.NoError(t, os.WriteFile(hostsFile, []byte(content), 0o600))

	t.Setenv("XDG_CONFIG_HOME", dir)
}

// isolateHostsFile sets XDG_CONFIG_HOME to an empty temp dir so
// ResolveToken does not read the user's real hosts.yml.
func isolateHostsFile(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

func TestResolveToken_EnvVar(t *testing.T) {
	isolateHostsFile(t)
	setupEmptyKeyring(t)
	t.Setenv("KH_API_KEY", "kh_test123")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "kh_test123", rt.Token)
	require.Equal(t, AuthMethodAPIKey, rt.Method)
	require.Equal(t, testHost, rt.Host)
}

func TestResolveToken_HostsYML(t *testing.T) {
	setupEmptyKeyring(t)
	t.Setenv("KH_API_KEY", "")
	writeHostsFile(t, testHost, "hosts_yml_token")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "hosts_yml_token", rt.Token)
	require.Equal(t, AuthMethodToken, rt.Method)
}

func TestResolveToken_KeyringLegacyFallback(t *testing.T) {
	// hosts.yml has no token for this host, so the keyring fallback is used.
	isolateHostsFile(t)
	setupKeyringWithToken(t, testHost, "keyring_token_abc")
	t.Setenv("KH_API_KEY", "")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "keyring_token_abc", rt.Token)
	require.Equal(t, AuthMethodToken, rt.Method)
}

func TestResolveToken_None(t *testing.T) {
	isolateHostsFile(t)
	setupEmptyKeyring(t)
	t.Setenv("KH_API_KEY", "")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "", rt.Token)
	require.Equal(t, AuthMethodNone, rt.Method)
}

func TestResolveToken_EnvVarPriority(t *testing.T) {
	isolateHostsFile(t)
	setupKeyringWithToken(t, testHost, "keyring_token")
	t.Setenv("KH_API_KEY", "env_wins")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "env_wins", rt.Token)
	require.Equal(t, AuthMethodAPIKey, rt.Method)
}

func TestResolveToken_HostsYMLOverKeyring(t *testing.T) {
	// hosts.yml now takes priority over the legacy keyring.
	setupKeyringWithToken(t, testHost, "keyring_token")
	t.Setenv("KH_API_KEY", "")
	writeHostsFile(t, testHost, "hosts_yml_wins")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "hosts_yml_wins", rt.Token)
	require.Equal(t, AuthMethodToken, rt.Method)
}
