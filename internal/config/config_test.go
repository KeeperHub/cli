package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keeperhub/cli/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.ConfigVersion != "1" {
		t.Errorf("ConfigVersion: got %q, want %q", cfg.ConfigVersion, "1")
	}
	if cfg.DefaultHost != "app.keeperhub.com" {
		t.Errorf("DefaultHost: got %q, want %q", cfg.DefaultHost, "app.keeperhub.com")
	}
}

func TestReadConfigMissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := config.ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig on missing file: %v", err)
	}
	if cfg.ConfigVersion != "1" {
		t.Errorf("ConfigVersion: got %q, want %q", cfg.ConfigVersion, "1")
	}
}

func TestWriteReadConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := config.Config{
		ConfigVersion: "2",
		DefaultHost:   "staging.keeperhub.com",
	}

	if err := config.WriteConfig(want); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	got, err := config.ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}

	if got.ConfigVersion != want.ConfigVersion {
		t.Errorf("ConfigVersion: got %q, want %q", got.ConfigVersion, want.ConfigVersion)
	}
	if got.DefaultHost != want.DefaultHost {
		t.Errorf("DefaultHost: got %q, want %q", got.DefaultHost, want.DefaultHost)
	}
}

func TestReadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configPath := filepath.Join(dir, "kh", "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("creating dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(":\tinvalid: yaml: [\n"), 0o600); err != nil {
		t.Fatalf("writing invalid config: %v", err)
	}

	_, err := config.ReadConfig()
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
	if err.Error() != "config file is invalid, run 'kh auth login' to reset it" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestConfigDirUsesXDG(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	got := config.ConfigDir()
	want := filepath.Join(dir, "kh")
	if got != want {
		t.Errorf("ConfigDir: got %q, want %q", got, want)
	}
}

func TestRPCEndpoint_Found(t *testing.T) {
	cfg := config.Config{
		RPC: map[string]string{
			"1":   "https://eth-mainnet.example.com",
			"137": "https://polygon.example.com",
		},
	}
	got := cfg.RPCEndpoint("1")
	if got != "https://eth-mainnet.example.com" {
		t.Errorf("RPCEndpoint(1): got %q, want %q", got, "https://eth-mainnet.example.com")
	}
}

func TestRPCEndpoint_NotFound(t *testing.T) {
	cfg := config.Config{
		RPC: map[string]string{
			"1": "https://eth-mainnet.example.com",
		},
	}
	got := cfg.RPCEndpoint("999")
	if got != "" {
		t.Errorf("RPCEndpoint(999): got %q, want empty string", got)
	}
}

func TestRPCEndpoint_NilMap(t *testing.T) {
	cfg := config.Config{}
	got := cfg.RPCEndpoint("1")
	if got != "" {
		t.Errorf("RPCEndpoint(1) with nil map: got %q, want empty string", got)
	}
}

func TestWriteReadConfigRoundTripWithRPC(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := config.Config{
		ConfigVersion: "1",
		DefaultHost:   "app.keeperhub.com",
		DefaultOrg:    "org_test123",
		RPC: map[string]string{
			"1":   "https://eth.example.com",
			"137": "https://polygon.example.com",
		},
	}

	if err := config.WriteConfig(want); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	got, err := config.ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}

	if got.DefaultOrg != want.DefaultOrg {
		t.Errorf("DefaultOrg: got %q, want %q", got.DefaultOrg, want.DefaultOrg)
	}
	if got.RPCEndpoint("1") != "https://eth.example.com" {
		t.Errorf("RPCEndpoint(1): got %q, want %q", got.RPCEndpoint("1"), "https://eth.example.com")
	}
	if got.RPCEndpoint("137") != "https://polygon.example.com" {
		t.Errorf("RPCEndpoint(137): got %q, want %q", got.RPCEndpoint("137"), "https://polygon.example.com")
	}
}
