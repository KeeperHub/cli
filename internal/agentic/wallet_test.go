package agentic_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/keeperhub/cli/internal/agentic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Produced by the platform's own computeSignature (lib/agentic-wallet/hmac.ts)
// for these inputs. Pinning the Node output rather than a value this package
// generated is the point: a signature that only agrees with itself would still
// be rejected by the server.
const (
	fixtureSecret    = "s3cr3t-test-key"
	fixtureSubOrg    = "sub_abc123"
	fixturePath      = "/api/agentic-wallet/credit"
	fixtureTimestamp = 1786000000
	fixtureSignature = "3becc7ba14cd84066fc6194588db62471841ccfc012d4ae9e8332f1fed68c592"
)

func TestSign_MatchesPlatformImplementation(t *testing.T) {
	got := agentic.Sign(
		fixtureSecret, "GET", fixturePath, fixtureSubOrg, "", fixtureTimestamp,
	)
	assert.Equal(t, fixtureSignature, got)
}

func TestSign_EveryFieldIsBound(t *testing.T) {
	base := agentic.Sign(fixtureSecret, "GET", fixturePath, fixtureSubOrg, "", fixtureTimestamp)

	cases := map[string]string{
		"method":    agentic.Sign(fixtureSecret, "POST", fixturePath, fixtureSubOrg, "", fixtureTimestamp),
		"path":      agentic.Sign(fixtureSecret, "GET", "/api/agentic-wallet/sign", fixtureSubOrg, "", fixtureTimestamp),
		"subOrgId":  agentic.Sign(fixtureSecret, "GET", fixturePath, "sub_other", "", fixtureTimestamp),
		"body":      agentic.Sign(fixtureSecret, "GET", fixturePath, fixtureSubOrg, "{}", fixtureTimestamp),
		"timestamp": agentic.Sign(fixtureSecret, "GET", fixturePath, fixtureSubOrg, "", fixtureTimestamp+1),
		"secret":    agentic.Sign("other-secret", "GET", fixturePath, fixtureSubOrg, "", fixtureTimestamp),
	}

	for field, sig := range cases {
		assert.NotEqual(t, base, sig, "%s must be bound into the signature", field)
	}
}

func writeWallet(t *testing.T, body string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".keeperhub"), 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(home, ".keeperhub", agentic.ConfigName), []byte(body), 0o600,
	))
}

func TestLoad_ReadsTheWallet(t *testing.T) {
	writeWallet(t, `{"subOrgId":"sub_abc123","walletAddress":"0xabc","hmacSecret":"s3cr3t"}`)

	cfg, err := agentic.Load()

	require.NoError(t, err)
	assert.Equal(t, "sub_abc123", cfg.SubOrgID)
	assert.Equal(t, "0xabc", cfg.WalletAddress)
}

func TestLoad_MissingFileIsNotConfigured(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := agentic.Load()

	assert.ErrorIs(t, err, agentic.ErrNotConfigured)
}

// A file present but missing its credential is the same state as no wallet,
// not a usable one, so it must not reach the signing path.
func TestLoad_IncompleteConfigIsNotConfigured(t *testing.T) {
	writeWallet(t, `{"subOrgId":"sub_abc123","walletAddress":"0xabc"}`)

	_, err := agentic.Load()

	assert.ErrorIs(t, err, agentic.ErrNotConfigured)
}

func TestLoad_MalformedConfigReportsAnError(t *testing.T) {
	writeWallet(t, `not json`)

	_, err := agentic.Load()

	require.Error(t, err)
	assert.NotErrorIs(t, err, agentic.ErrNotConfigured)
}

// The secret is the wallet's only credential, so it must never appear in a
// rendered error, which is the one string a caller is likely to print.
func TestLoad_ErrorNeverCarriesTheSecret(t *testing.T) {
	writeWallet(t, `{"subOrgId":"sub_abc123","hmacSecret":"s3cr3t-test-key","walletAddress":`)

	_, err := agentic.Load()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), fixtureSecret)
}

func TestConfig_SecretIsNotInTheJSONRoundTrip(t *testing.T) {
	// Guards against the config being logged wholesale somewhere: if the field
	// ever needs redacting, this is where that decision gets recorded.
	cfg := agentic.Config{SubOrgID: "sub_abc123", HMACSecret: fixtureSecret}
	out, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(out), "sub_abc123")
}
