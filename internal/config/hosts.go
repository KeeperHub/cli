package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultHost = "app.keeperhub.io"

// HostConfig holds per-host authentication and connection details.
type HostConfig struct {
	User    string            `yaml:"user,omitempty"`
	Token   string            `yaml:"token,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

// HostsConfig is the top-level structure for hosts.yml.
type HostsConfig struct {
	Hosts         map[string]HostConfig `yaml:"hosts,omitempty"`
	DefaultHost   string                `yaml:"default_host,omitempty"`
	ConfigVersion string                `yaml:"config_version,omitempty"`
}

// ReadHosts reads HostsFile() and returns the parsed HostsConfig.
// If the file does not exist, it returns an empty HostsConfig without error.
func ReadHosts() (HostsConfig, error) {
	path := HostsFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return HostsConfig{}, nil
		}
		return HostsConfig{}, fmt.Errorf("reading hosts file: %w", err)
	}

	var cfg HostsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return HostsConfig{}, fmt.Errorf("hosts file is invalid: %w", err)
	}

	return cfg, nil
}

// WriteHosts serialises cfg to HostsFile(), creating the directory if needed.
func WriteHosts(cfg HostsConfig) error {
	path := HostsFile()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshalling hosts: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing hosts file: %w", err)
	}

	return nil
}

// ActiveHost resolves the active host using the priority chain:
// flagHost > envHost > h.DefaultHost > config.yml DefaultHost > "app.keeperhub.io"
func (h *HostsConfig) ActiveHost(flagHost, envHost string) string {
	if flagHost != "" {
		return flagHost
	}
	if envHost != "" {
		return envHost
	}
	if h.DefaultHost != "" {
		return h.DefaultHost
	}
	if cfg, err := ReadConfig(); err == nil && cfg.DefaultHost != "" && cfg.DefaultHost != defaultHost {
		return cfg.DefaultHost
	}
	return defaultHost
}

// HostEntry looks up the HostConfig for the given hostname.
// Returns the entry and true if found, or an empty HostConfig and false otherwise.
func (h *HostsConfig) HostEntry(hostname string) (HostConfig, bool) {
	entry, ok := h.Hosts[hostname]
	return entry, ok
}

// SetHostToken updates the token for a specific host in the hosts config.
// If the host entry doesn't exist, it creates one.
func SetHostToken(host, token string) error {
	hosts, err := ReadHosts()
	if err != nil {
		return err
	}
	if hosts.Hosts == nil {
		hosts.Hosts = make(map[string]HostConfig)
	}
	entry := hosts.Hosts[host]
	entry.Token = token
	hosts.Hosts[host] = entry
	return WriteHosts(hosts)
}

// ClearHostToken removes the token for a specific host.
// Returns nil if the host is not in the config.
func ClearHostToken(host string) error {
	hosts, err := ReadHosts()
	if err != nil {
		return err
	}
	if entry, ok := hosts.Hosts[host]; ok {
		entry.Token = ""
		hosts.Hosts[host] = entry
		return WriteHosts(hosts)
	}
	return nil
}
