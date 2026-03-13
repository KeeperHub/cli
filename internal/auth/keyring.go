package auth

import (
	"errors"
	"os"

	"github.com/99designs/keyring"
	"github.com/keeperhub/cli/internal/config"
)

const keyringService = "kh"

// openKeyringFunc is the function used to open the keyring. Tests may override this.
var openKeyringFunc = openKeyring

func openKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName:     keyringService,
		FileDir:         config.ConfigDir(),
		FilePasswordFunc: keyring.FixedStringPrompt(""),
		AllowedBackends: []keyring.BackendType{
			keyring.FileBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
		},
	})
}

// SetToken stores a token for the given host in the OS keychain.
func SetToken(host, token string) error {
	kr, err := openKeyringFunc()
	if err != nil {
		return err
	}
	return kr.Set(keyring.Item{
		Key:  host,
		Data: []byte(token),
	})
}

// GetToken retrieves the token for the given host from the OS keychain.
// Returns an empty string (and nil error) if no token is stored for the host.
func GetToken(host string) (string, error) {
	kr, err := openKeyringFunc()
	if err != nil {
		return "", err
	}
	item, err := kr.Get(host)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", nil
		}
		return "", err
	}
	return string(item.Data), nil
}

// DeleteToken removes the stored token for the given host from the OS keychain.
// Returns nil if no token exists for that host.
func DeleteToken(host string) error {
	kr, err := openKeyringFunc()
	if err != nil {
		return err
	}
	err = kr.Remove(host)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) || errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}
