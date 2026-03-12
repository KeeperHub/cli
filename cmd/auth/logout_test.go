package auth_test

import (
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/auth"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func TestLogoutCmd_ClearsCredentials(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()

	deleteHost := ""
	auth.DeleteTokenFunc = func(host string) error {
		deleteHost = host
		return nil
	}
	auth.ClearHostTokenFunc = func(host string) error {
		return nil
	}

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewLogoutCmd(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleteHost == "" {
		t.Error("expected DeleteToken to be called with a host")
	}

	out := buf.String()
	if !strings.Contains(out, "Logged out") {
		t.Errorf("expected 'Logged out' in output, got: %q", out)
	}
}

func TestLogoutCmd_PrintsConfirmation(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()

	auth.DeleteTokenFunc = func(host string) error { return nil }
	auth.ClearHostTokenFunc = func(host string) error { return nil }

	f := &cmdutil.Factory{IOStreams: ios}
	cmd := auth.NewLogoutCmd(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Error("expected some output after logout")
	}
}
