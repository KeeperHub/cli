package cmdutil

import (
	"testing"

	"github.com/keeperhub/cli/internal/config"
	"github.com/spf13/cobra"
)

func makeRootCmd(hostFlag string) *cobra.Command {
	root := &cobra.Command{Use: "kh"}
	root.PersistentFlags().StringP("host", "H", "", "KeeperHub host")
	if hostFlag != "" {
		_ = root.PersistentFlags().Set("host", hostFlag)
	}
	child := &cobra.Command{Use: "test"}
	root.AddCommand(child)
	return child
}

func TestResolveHost_FlagTakesPriority(t *testing.T) {
	cmd := makeRootCmd("http://localhost:3000")
	cfg := config.Config{DefaultHost: "staging.keeperhub.com"}
	t.Setenv("KH_HOST", "env.keeperhub.com")

	got := ResolveHost(cmd, cfg)
	if got != "http://localhost:3000" {
		t.Errorf("expected flag host, got %q", got)
	}
}

func TestResolveHost_EnvFallback(t *testing.T) {
	cmd := makeRootCmd("")
	cfg := config.Config{DefaultHost: "staging.keeperhub.com"}
	t.Setenv("KH_HOST", "env.keeperhub.com")

	got := ResolveHost(cmd, cfg)
	if got != "env.keeperhub.com" {
		t.Errorf("expected env host, got %q", got)
	}
}

func TestResolveHost_ConfigFallback(t *testing.T) {
	cmd := makeRootCmd("")
	cfg := config.Config{DefaultHost: "staging.keeperhub.com"}
	t.Setenv("KH_HOST", "")

	got := ResolveHost(cmd, cfg)
	if got != "staging.keeperhub.com" {
		t.Errorf("expected config host, got %q", got)
	}
}

func TestResolveHost_BuiltInDefault(t *testing.T) {
	cmd := makeRootCmd("")
	cfg := config.Config{}
	t.Setenv("KH_HOST", "")

	got := ResolveHost(cmd, cfg)
	if got != "app.keeperhub.com" {
		t.Errorf("expected built-in default, got %q", got)
	}
}

func TestResolveHost_NilCmd(t *testing.T) {
	cfg := config.Config{DefaultHost: "staging.keeperhub.com"}
	t.Setenv("KH_HOST", "")

	got := ResolveHost(nil, cfg)
	if got != "staging.keeperhub.com" {
		t.Errorf("expected config host with nil cmd, got %q", got)
	}
}
