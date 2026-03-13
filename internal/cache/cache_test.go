package cache_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keeperhub/cli/internal/cache"
)

func TestCacheDir_XDGEnvSet(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	got := cache.CacheDir()
	want := filepath.Join(tmp, "kh")
	if got != want {
		t.Errorf("CacheDir() = %q, want %q", got, want)
	}
}

func TestCacheDir_Fallback(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")
	got := cache.CacheDir()
	if got == "" {
		t.Error("CacheDir() returned empty string")
	}
	// Should end in .cache/kh
	if filepath.Base(got) != "kh" {
		t.Errorf("CacheDir() base = %q, want %q", filepath.Base(got), "kh")
	}
}

func TestWriteCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	data := json.RawMessage(`{"hello":"world"}`)
	err := cache.WriteCache("test.json", data)
	if err != nil {
		t.Fatalf("WriteCache() error = %v", err)
	}

	expectedPath := filepath.Join(tmp, "kh", "test.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("WriteCache() did not create file at %s", expectedPath)
	}
}

func TestReadCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	data := json.RawMessage(`{"hello":"world"}`)
	err := cache.WriteCache("test.json", data)
	if err != nil {
		t.Fatalf("WriteCache() error = %v", err)
	}

	entry, err := cache.ReadCache("test.json")
	if err != nil {
		t.Fatalf("ReadCache() error = %v", err)
	}
	if entry == nil {
		t.Fatal("ReadCache() returned nil entry")
	}
	if entry.FetchedAt.IsZero() {
		t.Error("ReadCache() entry has zero FetchedAt")
	}
	if string(entry.Data) != `{"hello":"world"}` {
		t.Errorf("ReadCache() data = %s, want %s", string(entry.Data), `{"hello":"world"}`)
	}
}

func TestReadCache_Missing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	_, err := cache.ReadCache("nonexistent.json")
	if err == nil {
		t.Error("ReadCache() on missing file: expected error, got nil")
	}
}

func TestIsStale_Fresh(t *testing.T) {
	entry := &cache.CacheEntry{
		FetchedAt: time.Now().UTC().Add(-30 * time.Minute),
		Data:      json.RawMessage(`{}`),
	}
	if cache.IsStale(entry, 1*time.Hour) {
		t.Error("IsStale() = true for 30-minute old entry with 1hr TTL, want false")
	}
}

func TestIsStale_Expired(t *testing.T) {
	entry := &cache.CacheEntry{
		FetchedAt: time.Now().UTC().Add(-2 * time.Hour),
		Data:      json.RawMessage(`{}`),
	}
	if !cache.IsStale(entry, 1*time.Hour) {
		t.Error("IsStale() = false for 2-hour old entry with 1hr TTL, want true")
	}
}
