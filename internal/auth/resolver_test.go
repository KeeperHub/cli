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

	origHome := os.Getenv("XDG_CONFIG_HOME")
	require.NoError(t, os.Setenv("XDG_CONFIG_HOME", dir))
	t.Cleanup(func() {
		if origHome == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origHome)
		}
	})
}

func TestResolveToken_EnvVar(t *testing.T) {
	setupEmptyKeyring(t)
	t.Setenv("KH_API_KEY", "kh_test123")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "kh_test123", rt.Token)
	require.Equal(t, AuthMethodAPIKey, rt.Method)
	require.Equal(t, testHost, rt.Host)
}

func TestResolveToken_Keyring(t *testing.T) {
	setupKeyringWithToken(t, testHost, "keyring_token_abc")
	t.Setenv("KH_API_KEY", "")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "keyring_token_abc", rt.Token)
	require.Equal(t, AuthMethodToken, rt.Method)
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

func TestResolveToken_None(t *testing.T) {
	setupEmptyKeyring(t)
	t.Setenv("KH_API_KEY", "")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "", rt.Token)
	require.Equal(t, AuthMethodNone, rt.Method)
}

func TestResolveToken_EnvVarPriority(t *testing.T) {
	setupKeyringWithToken(t, testHost, "keyring_token")
	t.Setenv("KH_API_KEY", "env_wins")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "env_wins", rt.Token)
	require.Equal(t, AuthMethodAPIKey, rt.Method)
}

func TestResolveToken_KeyringOverHostsYML(t *testing.T) {
	setupKeyringWithToken(t, testHost, "keyring_takes_precedence")
	t.Setenv("KH_API_KEY", "")
	writeHostsFile(t, testHost, "hosts_yml_token")

	rt, err := ResolveToken(testHost)
	require.NoError(t, err)
	require.Equal(t, "keyring_takes_precedence", rt.Token)
	require.Equal(t, AuthMethodToken, rt.Method)
}
