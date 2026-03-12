package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/99designs/keyring"
	"github.com/stretchr/testify/require"
)

func newFileKeyring(t *testing.T) keyring.Keyring {
	t.Helper()
	dir := t.TempDir()
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
		FileDir:          filepath.Join(dir, "keyring"),
		FilePasswordFunc: keyring.FixedStringPrompt(""),
	})
	require.NoError(t, err)
	return kr
}

func overrideKeyring(t *testing.T) {
	t.Helper()
	kr := newFileKeyring(t)
	original := openKeyringFunc
	openKeyringFunc = func() (keyring.Keyring, error) { return kr, nil }
	t.Cleanup(func() { openKeyringFunc = original })
}

func TestSetGetToken(t *testing.T) {
	overrideKeyring(t)

	err := SetToken("app.keeperhub.io", "tok_abc123")
	require.NoError(t, err)

	got, err := GetToken("app.keeperhub.io")
	require.NoError(t, err)
	require.Equal(t, "tok_abc123", got)
}

func TestGetToken_Missing(t *testing.T) {
	overrideKeyring(t)

	got, err := GetToken("nonexistent.example.com")
	require.NoError(t, err)
	require.Equal(t, "", got)
}

func TestDeleteToken(t *testing.T) {
	overrideKeyring(t)

	require.NoError(t, SetToken("app.keeperhub.io", "tok_xyz"))

	require.NoError(t, DeleteToken("app.keeperhub.io"))

	got, err := GetToken("app.keeperhub.io")
	require.NoError(t, err)
	require.Equal(t, "", got)
}

func TestDeleteToken_NonExistent(t *testing.T) {
	overrideKeyring(t)

	err := DeleteToken("ghost.keeperhub.io")
	require.NoError(t, err)
}

func TestSetToken_MultipleHosts(t *testing.T) {
	overrideKeyring(t)

	require.NoError(t, SetToken("host1.example.com", "token1"))
	require.NoError(t, SetToken("host2.example.com", "token2"))

	got1, err := GetToken("host1.example.com")
	require.NoError(t, err)
	require.Equal(t, "token1", got1)

	got2, err := GetToken("host2.example.com")
	require.NoError(t, err)
	require.Equal(t, "token2", got2)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
