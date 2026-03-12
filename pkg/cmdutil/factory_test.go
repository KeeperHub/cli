package cmdutil_test

import (
	"fmt"
	"testing"

	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func TestFactoryConstruction(t *testing.T) {
	s, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{
		AppVersion: "1.2.3",
		IOStreams:   s,
		Config: func() (config.Config, error) {
			return config.DefaultConfig(), nil
		},
	}

	if f.AppVersion != "1.2.3" {
		t.Errorf("AppVersion: got %q, want %q", f.AppVersion, "1.2.3")
	}
	if f.IOStreams == nil {
		t.Error("IOStreams is nil")
	}
}

func TestFactoryConfigFunc(t *testing.T) {
	want := config.Config{ConfigVersion: "42", DefaultHost: "test.example.com"}
	f := &cmdutil.Factory{
		Config: func() (config.Config, error) {
			return want, nil
		},
	}

	got, err := f.Config()
	if err != nil {
		t.Fatalf("Config() error: %v", err)
	}
	if got.ConfigVersion != want.ConfigVersion {
		t.Errorf("ConfigVersion: got %q, want %q", got.ConfigVersion, want.ConfigVersion)
	}
	if got.DefaultHost != want.DefaultHost {
		t.Errorf("DefaultHost: got %q, want %q", got.DefaultHost, want.DefaultHost)
	}
}

func TestFactoryConfigFuncError(t *testing.T) {
	f := &cmdutil.Factory{
		Config: func() (config.Config, error) {
			return config.Config{}, fmt.Errorf("config load failed")
		},
	}

	_, err := f.Config()
	if err == nil {
		t.Error("expected error from Config(), got nil")
	}
}
