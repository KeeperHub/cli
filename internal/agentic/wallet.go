// Package agentic reads the local agentic-wallet config and signs requests as
// that wallet.
//
// Requests authenticate with the wallet's own HMAC secret rather than the
// user's bearer token. The signature proves possession of one specific
// wallet's secret, so a caller can only ever read the wallet it actually
// holds, and no endpoint has to accept "tell me about wallet X" from an
// account that merely has access to the platform.
package agentic

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// ConfigName is the file `kh wallet add` writes, mode 0600.
const ConfigName = "wallet.json"

// Config is the on-disk agentic wallet. HMACSecret must never be printed or
// logged; it is the wallet's only credential.
type Config struct {
	SubOrgID      string `json:"subOrgId"`
	WalletAddress string `json:"walletAddress"`
	HMACSecret    string `json:"hmacSecret"`
}

// ConfigPath returns the location of the agentic wallet config.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".keeperhub", ConfigName), nil
}

// ErrNotConfigured reports that no agentic wallet exists on this machine. It
// is a normal state, not a failure: most installs never provision one.
var ErrNotConfigured = errors.New("no agentic wallet configured")

// ErrIncompleteConfig reports a config file that is present but missing a
// field needed to use it. Distinct from ErrNotConfigured so a caller can say
// "the file is there and broken" rather than sending someone to re-run the
// command that already wrote it.
var ErrIncompleteConfig = errors.New("agentic wallet config is incomplete")

// Load reads the agentic wallet config, returning ErrNotConfigured when the
// file is absent.
func Load() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Config{}, ErrNotConfigured
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading %s: %w", ConfigName, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decoding %s: %w", ConfigName, err)
	}
	if cfg.SubOrgID == "" || cfg.HMACSecret == "" {
		return Config{}, ErrIncompleteConfig
	}
	return cfg, nil
}

// Header names carrying the signature. The platform also accepts an optional
// X-KH-Key-Version to pin secret selection during rotation; it is deliberately
// not sent, so an unpinned request is verified against every active secret and
// keeps working through the rotation grace window.
const (
	HeaderSubOrg    = "X-KH-Sub-Org"
	HeaderTimestamp = "X-KH-Timestamp"
	HeaderSignature = "X-KH-Signature"
)

// Sign returns the hex HMAC-SHA256 over
// `method\npath\nsubOrgId\nsha256_hex(body)\ntimestamp`, the format the
// platform verifies. path is the URL pathname only, with no host or query.
// A GET signs over an empty body.
func Sign(secret, method, path, subOrgID, body string, timestamp int64) string {
	bodyDigest := sha256.Sum256([]byte(body))
	signingString := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s",
		method,
		path,
		subOrgID,
		hex.EncodeToString(bodyDigest[:]),
		strconv.FormatInt(timestamp, 10),
	)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingString))
	return hex.EncodeToString(mac.Sum(nil))
}

// SignRequest attaches the signature headers to req.
//
// The signed path is taken from req.URL, never from a caller-supplied
// constant. The platform signs the pathname it actually received, so a host
// configured with a path prefix would otherwise produce a signature over a
// different string and come back 401, which reads as a rotated secret rather
// than a configuration problem.
//
// body must be the exact bytes sent; a GET signs over an empty string.
func SignRequest(req *http.Request, cfg Config, body string, now time.Time) {
	timestamp := now.Unix()
	req.Header.Set(HeaderSubOrg, cfg.SubOrgID)
	req.Header.Set(HeaderTimestamp, strconv.FormatInt(timestamp, 10))
	req.Header.Set(HeaderSignature, Sign(
		cfg.HMACSecret, req.Method, req.URL.Path, cfg.SubOrgID, body, timestamp,
	))
}

// String renders the config without its secret, so a Config reaching a log
// line or an error through %v or %s cannot leak the credential. Load is the
// only thing that reads the secret out; nothing else should print a Config.
func (c Config) String() string {
	secret := "unset"
	if c.HMACSecret != "" {
		secret = "redacted"
	}
	return fmt.Sprintf(
		"agentic.Config{SubOrgID: %q, WalletAddress: %q, HMACSecret: %s}",
		c.SubOrgID, c.WalletAddress, secret,
	)
}
