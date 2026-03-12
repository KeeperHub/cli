package auth_test

import (
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/auth"
	internalauth "github.com/keeperhub/cli/internal/auth"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func TestLoginCmd_BrowserFlow(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()

	browserCalled := false
	auth.BrowserLoginFunc = func(host string, ios2 *iostreams.IOStreams) (string, error) {
		browserCalled = true
		return "test-token", nil
	}
	auth.DeviceLoginFunc = func(host string, ios2 *iostreams.IOStreams) (string, error) {
		t.Fatal("DeviceLogin should not be called in browser flow")
		return "", nil
	}
	auth.SetTokenFunc = func(host, token string) error { return nil }
	auth.FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{Email: "user@example.com"}, nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewLoginCmd(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !browserCalled {
		t.Error("expected BrowserLogin to be called")
	}
	out := buf.String()
	if !strings.Contains(out, "app.keeperhub.io") {
		t.Errorf("expected host in output, got: %q", out)
	}
}

func TestLoginCmd_NoBrowserFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	deviceCalled := false
	auth.BrowserLoginFunc = func(host string, ios2 *iostreams.IOStreams) (string, error) {
		t.Fatal("BrowserLogin should not be called with --no-browser")
		return "", nil
	}
	auth.DeviceLoginFunc = func(host string, ios2 *iostreams.IOStreams) (string, error) {
		deviceCalled = true
		return "test-token", nil
	}
	auth.SetTokenFunc = func(host, token string) error { return nil }
	auth.FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{Email: "user@example.com"}, nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewLoginCmd(f)
	cmd.SetArgs([]string{"--no-browser"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deviceCalled {
		t.Error("expected DeviceLogin to be called")
	}
}

func TestLoginCmd_WithTokenFlag(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()
	ios.In = strings.NewReader("my-token-from-stdin\n")

	storeHost := ""
	storeToken := ""
	auth.BrowserLoginFunc = func(host string, ios2 *iostreams.IOStreams) (string, error) {
		t.Fatal("BrowserLogin should not be called with --with-token")
		return "", nil
	}
	auth.DeviceLoginFunc = func(host string, ios2 *iostreams.IOStreams) (string, error) {
		t.Fatal("DeviceLogin should not be called with --with-token")
		return "", nil
	}
	auth.SetTokenFunc = func(host, token string) error {
		storeHost = host
		storeToken = token
		return nil
	}
	auth.FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{Email: "user@example.com"}, nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewLoginCmd(f)
	cmd.SetArgs([]string{"--with-token"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storeToken != "my-token-from-stdin" {
		t.Errorf("expected stored token 'my-token-from-stdin', got %q", storeToken)
	}
	if storeHost == "" {
		t.Error("expected host to be set")
	}
	out := buf.String()
	if !strings.Contains(out, storeHost) {
		t.Errorf("expected host in output, got: %q", out)
	}
}

func TestLoginCmd_WithTokenFlag_EmptyStdin(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.In = strings.NewReader("")

	auth.BrowserLoginFunc = func(host string, ios2 *iostreams.IOStreams) (string, error) { return "", nil }
	auth.DeviceLoginFunc = func(host string, ios2 *iostreams.IOStreams) (string, error) { return "", nil }
	auth.SetTokenFunc = func(host, token string) error { return nil }
	auth.FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{}, nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewLoginCmd(f)
	cmd.SetArgs([]string{"--with-token"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty stdin, got nil")
	}
}
