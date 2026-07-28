package auth_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/auth"
	internalauth "github.com/keeperhub/cli/internal/auth"
	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

const testHost = "app.keeperhub.com"

// Completing the device flow mints a new organization API key server-side and
// the plaintext is returned exactly once, so a repeat login both strands the
// stored key and leaves a spare behind on the org. These tests pin that login
// is skipped when a working credential is already stored.

func TestLoginCmd_SkipsWhenStoredTokenIsValid(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.SetHostToken(testHost, "kh_already_stored"); err != nil {
		t.Fatalf("seeding token: %v", err)
	}

	ios, buf, _, _ := iostreams.Test()
	auth.DeviceLoginFunc = func(string, *iostreams.IOStreams) (string, error) {
		t.Fatal("DeviceLogin must not run when a valid token is stored")
		return "", nil
	}
	auth.SetTokenFunc = func(string, string) error { return nil }
	auth.FetchTokenInfoFunc = func(_, token string) (internalauth.TokenInfo, error) {
		if token != "kh_already_stored" {
			t.Errorf("expected the stored token to be validated, got %q", token)
		}
		return internalauth.TokenInfo{Email: "user@example.com"}, nil
	}

	cmd := auth.NewLoginCmd(&cmdutil.Factory{IOStreams: ios})
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Already logged in") {
		t.Errorf("expected an already-logged-in notice, got: %q", out)
	}
	if !strings.Contains(out, "--force") {
		t.Errorf("expected the notice to mention --force, got: %q", out)
	}
	if strings.Contains(out, "Created a new organization API key") {
		t.Errorf("must not claim a key was created when skipping, got: %q", out)
	}
}

func TestLoginCmd_ForceReplacesStoredToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.SetHostToken(testHost, "kh_already_stored"); err != nil {
		t.Fatalf("seeding token: %v", err)
	}

	ios, buf, _, _ := iostreams.Test()
	deviceCalled := false
	auth.DeviceLoginFunc = func(string, *iostreams.IOStreams) (string, error) {
		deviceCalled = true
		return "kh_fresh_key", nil
	}
	auth.SetTokenFunc = func(string, string) error { return nil }
	auth.FetchTokenInfoFunc = func(string, string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{Email: "user@example.com"}, nil
	}

	cmd := auth.NewLoginCmd(&cmdutil.Factory{IOStreams: ios})
	cmd.SetArgs([]string{"--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deviceCalled {
		t.Error("expected --force to re-run the device flow")
	}
	if !strings.Contains(buf.String(), "Created a new organization API key") {
		t.Errorf("expected key-creation notice, got: %q", buf.String())
	}
}

func TestLoginCmd_ReLoginsWhenStoredTokenIsRejected(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.SetHostToken(testHost, "kh_revoked"); err != nil {
		t.Fatalf("seeding token: %v", err)
	}

	ios, _, _, _ := iostreams.Test()
	deviceCalled := false
	auth.DeviceLoginFunc = func(string, *iostreams.IOStreams) (string, error) {
		deviceCalled = true
		return "kh_fresh_key", nil
	}
	auth.SetTokenFunc = func(string, string) error { return nil }
	// A revoked or expired stored key must not wedge the user out of logging in.
	auth.FetchTokenInfoFunc = func(_, token string) (internalauth.TokenInfo, error) {
		if token == "kh_revoked" {
			return internalauth.TokenInfo{}, errors.New("token is invalid or expired")
		}
		return internalauth.TokenInfo{Email: "user@example.com"}, nil
	}

	cmd := auth.NewLoginCmd(&cmdutil.Factory{IOStreams: ios})
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deviceCalled {
		t.Error("expected re-login when the stored token is rejected")
	}
}

func TestLoginCmd_AnnouncesApiKeyCreation(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	ios, buf, _, _ := iostreams.Test()
	auth.DeviceLoginFunc = func(string, *iostreams.IOStreams) (string, error) {
		return "kh_fresh_key", nil
	}
	auth.SetTokenFunc = func(string, string) error { return nil }
	auth.FetchTokenInfoFunc = func(string, string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{Email: "user@example.com"}, nil
	}

	cmd := auth.NewLoginCmd(&cmdutil.Factory{IOStreams: ios})
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Created a new organization API key") {
		t.Errorf("expected key-creation notice, got: %q", out)
	}
	// The user needs to know where it landed and how to get rid of it.
	if !strings.Contains(out, "hosts.yml") {
		t.Errorf("expected the storage location in output, got: %q", out)
	}
	if !strings.Contains(out, "Revoke") {
		t.Errorf("expected revocation guidance in output, got: %q", out)
	}
}

func TestLoginCmd_WithTokenDoesNotClaimKeyCreation(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	ios, buf, _, _ := iostreams.Test()
	ios.In = strings.NewReader("kh_supplied_by_user\n")
	auth.DeviceLoginFunc = func(string, *iostreams.IOStreams) (string, error) {
		t.Fatal("DeviceLogin must not run with --with-token")
		return "", nil
	}
	auth.SetTokenFunc = func(string, string) error { return nil }
	auth.FetchTokenInfoFunc = func(string, string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{Email: "user@example.com"}, nil
	}

	cmd := auth.NewLoginCmd(&cmdutil.Factory{IOStreams: ios})
	cmd.SetArgs([]string{"--with-token"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// --with-token supplies an existing credential; nothing was created.
	if strings.Contains(buf.String(), "Created a new organization API key") {
		t.Errorf("must not claim key creation for --with-token, got: %q", buf.String())
	}
}
