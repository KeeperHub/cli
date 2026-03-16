package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultHost = "app.keeperhub.com"

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
// flagHost > envHost > h.DefaultHost > config.yml DefaultHost > "app.keeperhub.com"
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
// It tries the raw value first, then falls back to a bare hostname (scheme stripped)
// so that --host https://app-staging.keeperhub.com matches a hosts.yml key of
// app-staging.keeperhub.com.
func (h *HostsConfig) HostEntry(hostname string) (HostConfig, bool) {
	if entry, ok := h.Hosts[hostname]; ok {
		return entry, true
	}
	bare := stripScheme(hostname)
	if bare != hostname {
		if entry, ok := h.Hosts[bare]; ok {
			return entry, true
		}
	}
	return HostConfig{}, false
}

// stripScheme removes the URL scheme (http:// or https://) and any trailing
// slash from a host string, returning just the hostname[:port].
func stripScheme(host string) string {
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		if u, err := url.Parse(host); err == nil {
			return u.Host
		}
	}
	return strings.TrimRight(host, "/")
}

// resolveHostKey returns the canonical key for a host in the hosts map.
// If a bare-hostname entry already exists for a full-URL host, it returns the
// bare hostname so the token is stored alongside existing headers (e.g. CF-Access).
func resolveHostKey(hosts map[string]HostConfig, host string) string {
	if _, ok := hosts[host]; ok {
		return host
	}
	bare := stripScheme(host)
	if bare != host {
		if _, ok := hosts[bare]; ok {
			return bare
		}
	}
	// No existing entry -- prefer bare hostname for cleanliness.
	return bare
}

// SetHostToken updates the token for a specific host in the hosts config.
// If a bare-hostname entry already exists, the token is merged into it.
func SetHostToken(host, token string) error {
	hosts, err := ReadHosts()
	if err != nil {
		return err
	}
	if hosts.Hosts == nil {
		hosts.Hosts = make(map[string]HostConfig)
	}
	key := resolveHostKey(hosts.Hosts, host)
	entry := hosts.Hosts[key]
	entry.Token = token
	hosts.Hosts[key] = entry
	return WriteHosts(hosts)
}

// ClearHostToken removes the token for a specific host.
// Returns nil if the host is not in the config.
func ClearHostToken(host string) error {
	hosts, err := ReadHosts()
	if err != nil {
		return err
	}
	key := resolveHostKey(hosts.Hosts, host)
	if entry, ok := hosts.Hosts[key]; ok {
		entry.Token = ""
		hosts.Hosts[key] = entry
		return WriteHosts(hosts)
	}
	return nil
}
