package auth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/keeperhub/cli/cmd/auth"
	internalauth "github.com/keeperhub/cli/internal/auth"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func TestStatusCmd_ShowsUserDetails(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()

	auth.ResolveTokenFunc = func(host string) (internalauth.ResolvedToken, error) {
		return internalauth.ResolvedToken{Token: "tok", Method: internalauth.AuthMethodToken, Host: host}, nil
	}
	auth.FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{
			Email:     "user@example.com",
			Name:      "Test User",
			OrgName:   "My Org",
			OrgID:     "org-123",
			Role:      "owner",
			ExpiresAt: time.Date(2026, 3, 14, 15, 30, 0, 0, time.UTC),
			Method:    internalauth.AuthMethodToken,
		}, nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewStatusCmd(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("expected email in output, got: %q", out)
	}
	if !strings.Contains(out, "My Org") {
		t.Errorf("expected org name in output, got: %q", out)
	}
}

func TestStatusCmd_NotAuthenticated(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	auth.ResolveTokenFunc = func(host string) (internalauth.ResolvedToken, error) {
		return internalauth.ResolvedToken{Method: internalauth.AuthMethodNone, Host: host}, nil
	}
	auth.FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{}, nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewStatusCmd(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when not authenticated, got nil")
	}
}

func TestStatusCmd_APIKeyMethod(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()

	auth.ResolveTokenFunc = func(host string) (internalauth.ResolvedToken, error) {
		return internalauth.ResolvedToken{Token: "kh_apikey", Method: internalauth.AuthMethodAPIKey, Host: host}, nil
	}
	auth.FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{
			Email:   "user@example.com",
			OrgName: "My Org",
			Role:    "member",
			Method:  internalauth.AuthMethodAPIKey,
		}, nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewStatusCmd(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "api-key") {
		t.Errorf("expected 'api-key' method in output, got: %q", out)
	}
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()

	auth.ResolveTokenFunc = func(host string) (internalauth.ResolvedToken, error) {
		return internalauth.ResolvedToken{Token: "tok", Method: internalauth.AuthMethodToken, Host: host}, nil
	}
	auth.FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{
			Email:   "user@example.com",
			OrgName: "My Org",
			Role:    "owner",
			Method:  internalauth.AuthMethodToken,
		}, nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewStatusCmd(f)
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.SetArgs([]string{"--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"email"`) {
		t.Errorf("expected JSON with 'email' field, got: %q", out)
	}
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("expected email value in JSON, got: %q", out)
	}
}
