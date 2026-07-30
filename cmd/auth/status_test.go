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

// fetchAPIKeyInfo has no real identity for an API key, so it sets Email to a
// truncated key prefix (e.g. "kh_EU7Fc1Xi..."). This pins that the table
// stops labeling that prefix "User", which reads as though the key were an
// account, and uses "Credential" for API-key auth instead.
func TestStatusCmd_APIKeyLabeledAsCredentialNotUser(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()

	auth.ResolveTokenFunc = func(host string) (internalauth.ResolvedToken, error) {
		return internalauth.ResolvedToken{Token: "kh_EU7Fc1Xi", Method: internalauth.AuthMethodAPIKey, Host: host}, nil
	}
	auth.FetchTokenInfoFunc = func(host, token string) (internalauth.TokenInfo, error) {
		return internalauth.TokenInfo{
			Email:   "kh_EU7Fc1Xi...",
			OrgName: "My Org",
			Role:    "api-key",
			Method:  internalauth.AuthMethodAPIKey,
		}, nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewStatusCmd(f)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Credential") {
		t.Errorf("expected the row labeled Credential for API-key auth, got: %q", out)
	}
	if strings.Contains(out, "User") {
		t.Errorf("must not label the key prefix as User, got: %q", out)
	}
}

func TestStatusCmd_SessionLabeledAsUser(t *testing.T) {
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
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "User") {
		t.Errorf("expected the row labeled User for session auth, got: %q", buf.String())
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
