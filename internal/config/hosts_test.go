package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keeperhub/cli/internal/config"
	"gopkg.in/yaml.v3"
)

func TestActiveHostFallback(t *testing.T) {
	h := config.HostsConfig{}
	got := h.ActiveHost("", "")
	if got != "app.keeperhub.io" {
		t.Errorf("ActiveHost fallback: got %q, want %q", got, "app.keeperhub.io")
	}
}

func TestActiveHostPriority(t *testing.T) {
	h := config.HostsConfig{DefaultHost: "default.example.com"}

	// flag > all
	if got := h.ActiveHost("flag.example.com", "env.example.com"); got != "flag.example.com" {
		t.Errorf("flag priority: got %q, want %q", got, "flag.example.com")
	}

	// env > default
	if got := h.ActiveHost("", "env.example.com"); got != "env.example.com" {
		t.Errorf("env priority: got %q, want %q", got, "env.example.com")
	}

	// default > hardcoded
	if got := h.ActiveHost("", ""); got != "default.example.com" {
		t.Errorf("default priority: got %q, want %q", got, "default.example.com")
	}
}

func TestHostsYAMLRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := config.HostsConfig{
		Hosts: map[string]config.HostConfig{
			"app-staging.keeperhub.com": {
				User:  "test@example.com",
				Token: "tok_abc123",
				Headers: map[string]string{
					"CF-Access-Client-Id":     "abc123",
					"CF-Access-Client-Secret": "secret456",
				},
			},
		},
		DefaultHost:   "app-staging.keeperhub.com",
		ConfigVersion: "1",
	}

	if err := config.WriteHosts(want); err != nil {
		t.Fatalf("WriteHosts: %v", err)
	}

	got, err := config.ReadHosts()
	if err != nil {
		t.Fatalf("ReadHosts: %v", err)
	}

	host, ok := got.Hosts["app-staging.keeperhub.com"]
	if !ok {
		t.Fatal("host entry not found after round-trip")
	}
	if host.User != "test@example.com" {
		t.Errorf("User: got %q, want %q", host.User, "test@example.com")
	}
	if host.Headers["CF-Access-Client-Id"] != "abc123" {
		t.Errorf("CF-Access-Client-Id: got %q, want %q", host.Headers["CF-Access-Client-Id"], "abc123")
	}
	if host.Headers["CF-Access-Client-Secret"] != "secret456" {
		t.Errorf("CF-Access-Client-Secret: got %q, want %q", host.Headers["CF-Access-Client-Secret"], "secret456")
	}
	if got.DefaultHost != "app-staging.keeperhub.com" {
		t.Errorf("DefaultHost: got %q, want %q", got.DefaultHost, "app-staging.keeperhub.com")
	}
}

func TestReadHostsMissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	got, err := config.ReadHosts()
	if err != nil {
		t.Fatalf("ReadHosts on missing file: %v", err)
	}
	if got.Hosts != nil {
		t.Errorf("expected nil Hosts map, got %v", got.Hosts)
	}
}

func TestHostEntryLookup(t *testing.T) {
	h := config.HostsConfig{
		Hosts: map[string]config.HostConfig{
			"app.keeperhub.io": {User: "user@example.com"},
		},
	}

	entry, ok := h.HostEntry("app.keeperhub.io")
	if !ok {
		t.Fatal("expected to find host entry")
	}
	if entry.User != "user@example.com" {
		t.Errorf("User: got %q, want %q", entry.User, "user@example.com")
	}

	_, ok = h.HostEntry("missing.example.com")
	if ok {
		t.Error("expected false for missing host")
	}
}

func TestHostsYAMLCFAccessHeaders(t *testing.T) {
	rawYAML := `hosts:
  app-staging.keeperhub.com:
    user: "test@example.com"
    headers:
      CF-Access-Client-Id: "abc123"
      CF-Access-Client-Secret: "secret456"
`
	dir := t.TempDir()
	hostsPath := filepath.Join(dir, "kh", "hosts.yml")
	if err := os.MkdirAll(filepath.Dir(hostsPath), 0o700); err != nil {
		t.Fatalf("creating dir: %v", err)
	}
	if err := os.WriteFile(hostsPath, []byte(rawYAML), 0o600); err != nil {
		t.Fatalf("writing hosts: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)

	got, err := config.ReadHosts()
	if err != nil {
		t.Fatalf("ReadHosts: %v", err)
	}

	host, ok := got.Hosts["app-staging.keeperhub.com"]
	if !ok {
		t.Fatal("host entry not found")
	}
	if host.Headers["CF-Access-Client-Id"] != "abc123" {
		t.Errorf("CF-Access-Client-Id: got %q, want %q", host.Headers["CF-Access-Client-Id"], "abc123")
	}
}

// Ensure HostsConfig marshals correctly for round-trip verification via raw yaml
func TestHostsConfigMarshal(t *testing.T) {
	h := config.HostsConfig{
		Hosts: map[string]config.HostConfig{
			"test.example.com": {
				User:    "alice",
				Headers: map[string]string{"X-Custom": "value"},
			},
		},
	}

	data, err := yaml.Marshal(&h)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}

	var roundTripped config.HostsConfig
	if err := yaml.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}

	host, ok := roundTripped.Hosts["test.example.com"]
	if !ok {
		t.Fatal("host entry missing after marshal/unmarshal")
	}
	if host.User != "alice" {
		t.Errorf("User: got %q, want %q", host.User, "alice")
	}
	if host.Headers["X-Custom"] != "value" {
		t.Errorf("X-Custom header: got %q, want %q", host.Headers["X-Custom"], "value")
	}
}
