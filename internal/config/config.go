package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level application configuration stored in config.yml.
type Config struct {
	ConfigVersion string `yaml:"config_version"`
	DefaultHost   string `yaml:"default_host,omitempty"`
}

// DefaultConfig returns a Config with built-in defaults applied.
func DefaultConfig() Config {
	return Config{
		ConfigVersion: "1",
		DefaultHost:   "app.keeperhub.com",
	}
}

// ReadConfig reads ConfigFile() and returns the parsed Config.
// If the file does not exist, it returns DefaultConfig() without error.
func ReadConfig() (Config, error) {
	path := ConfigFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return Config{}, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("config file is invalid, run 'kh auth login' to reset it")
	}

	return cfg, nil
}

// WriteConfig serialises cfg to ConfigFile(), creating the directory if needed.
func WriteConfig(cfg Config) error {
	path := ConfigFile()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
