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
	if cfg.DefaultHost != "app.keeperhub.io" {
		t.Errorf("DefaultHost: got %q, want %q", cfg.DefaultHost, "app.keeperhub.io")
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
