package config

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the directory where kh stores its configuration files.
// Defaults to ~/.config/kh; respects $XDG_CONFIG_HOME if set.
func ConfigDir() string {
	return filepath.Join(xdgConfigHome(), "kh")
}

// StateDir returns the directory where kh stores runtime state.
// Defaults to ~/.local/state/kh; respects $XDG_STATE_HOME if set.
func StateDir() string {
	return filepath.Join(xdgStateHome(), "kh")
}

// ConfigFile returns the full path to config.yml.
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.yml")
}

// HostsFile returns the full path to hosts.yml.
func HostsFile() string {
	return filepath.Join(ConfigDir(), "hosts.yml")
}

// xdgConfigHome resolves the XDG config home directory, respecting the
// XDG_CONFIG_HOME environment variable.
func xdgConfigHome() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".config")
	}
	return filepath.Join(home, ".config")
}

// xdgStateHome resolves the XDG state home directory, respecting the
// XDG_STATE_HOME environment variable.
func xdgStateHome() string {
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".local", "state")
	}
	return filepath.Join(home, ".local", "state")
}
