package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	// ProtocolCacheName is the filename for the cached /api/mcp/schemas response.
	ProtocolCacheName = "schemas.json"

	// ProtocolCacheTTL is the TTL for protocol schema cache entries.
	ProtocolCacheTTL = 1 * time.Hour
)

// CacheEntry wraps cached data with a fetch timestamp.
type CacheEntry struct {
	FetchedAt time.Time       `json:"fetchedAt"`
	Data      json.RawMessage `json:"data"`
}

// CacheDir returns the directory where kh stores cache files.
// Defaults to ~/.cache/kh; respects $XDG_CACHE_HOME if set.
func CacheDir() string {
	return filepath.Join(xdgCacheHome(), "kh")
}

// ReadCache reads and unmarshals the named cache file from the cache directory.
// Returns an error if the file does not exist or cannot be parsed.
func ReadCache(name string) (*CacheEntry, error) {
	path := filepath.Join(CacheDir(), name)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entry CacheEntry
	if err := json.Unmarshal(b, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// WriteCache serializes data into a CacheEntry with the current UTC time and
// writes it to the cache directory. The cache directory is created if needed.
func WriteCache(name string, data json.RawMessage) error {
	dir := CacheDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	entry := CacheEntry{
		FetchedAt: time.Now().UTC(),
		Data:      data,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), b, 0o600)
}

// WriteRaw writes pre-serialized bytes directly to the named cache file.
// This is useful for tests that need to inject a cache entry with a specific
// FetchedAt timestamp.
func WriteRaw(name string, b []byte) error {
	dir := CacheDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), b, 0o600)
}

// IsStale reports whether the cache entry is older than the given TTL.
func IsStale(entry *CacheEntry, ttl time.Duration) bool {
	return time.Since(entry.FetchedAt) > ttl
}

// xdgCacheHome resolves the XDG cache home directory, respecting the
// XDG_CACHE_HOME environment variable.
func xdgCacheHome() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".cache")
	}
	return filepath.Join(home, ".cache")
}
