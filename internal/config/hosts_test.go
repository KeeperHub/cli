package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keeperhub/cli/internal/config"
	"gopkg.in/yaml.v3"
)

func TestActiveHostFallback(t *testing.T) {
	// Use an empty temp dir so no config.yml exists, ensuring the hardcoded default is used.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

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

// TestActiveHostConfigYMLFallback verifies that when hosts.yml has no default_host,
// ActiveHost falls back to the default_host in config.yml.
func TestActiveHostConfigYMLFallback(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Write config.yml with a custom default_host
	cfg := config.Config{ConfigVersion: "1", DefaultHost: "staging.example.com"}
	if err := config.WriteConfig(cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	// hosts.yml has no default_host (empty HostsConfig)
	h := config.HostsConfig{}
	got := h.ActiveHost("", "")
	if got != "staging.example.com" {
		t.Errorf("ActiveHost config.yml fallback: got %q, want %q", got, "staging.example.com")
	}
}

// TestActiveHostHostsYMLOverridesConfigYML verifies that hosts.yml default_host
// takes priority over config.yml default_host.
func TestActiveHostHostsYMLOverridesConfigYML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Write config.yml with a custom default_host
	cfg := config.Config{ConfigVersion: "1", DefaultHost: "config.example.com"}
	if err := config.WriteConfig(cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	// hosts.yml has its own default_host -- it should win
	h := config.HostsConfig{DefaultHost: "hosts.example.com"}
	got := h.ActiveHost("", "")
	if got != "hosts.example.com" {
		t.Errorf("hosts.yml should override config.yml: got %q, want %q", got, "hosts.example.com")
	}
}

// TestActiveHostFlagOverridesConfigYML verifies that the flag argument wins
// even when config.yml has a custom default_host.
func TestActiveHostFlagOverridesConfigYML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Write config.yml with a custom default_host
	cfg := config.Config{ConfigVersion: "1", DefaultHost: "config.example.com"}
	if err := config.WriteConfig(cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	h := config.HostsConfig{}
	got := h.ActiveHost("flag.example.com", "")
	if got != "flag.example.com" {
		t.Errorf("flag should override config.yml: got %q, want %q", got, "flag.example.com")
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
