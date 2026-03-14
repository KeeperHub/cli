package update_test

import (
	"strings"
	"testing"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/keeperhub/cli/cmd/update"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func TestIsHomebrewInstall_CellarOptHomebrew(t *testing.T) {
	result := update.TestableIsHomebrewInstall("/opt/homebrew/Cellar/kh/0.1.0/bin/kh")
	if !result {
		t.Error("expected true for /opt/homebrew/Cellar path")
	}
}

func TestIsHomebrewInstall_CellarUsrLocal(t *testing.T) {
	result := update.TestableIsHomebrewInstall("/usr/local/Cellar/kh/0.1.0/bin/kh")
	if !result {
		t.Error("expected true for /usr/local/Cellar path")
	}
}

func TestIsHomebrewInstall_SymlinkNotInCellar(t *testing.T) {
	result := update.TestableIsHomebrewInstall("/usr/local/bin/kh")
	if result {
		t.Error("expected false for /usr/local/bin path (not in Cellar)")
	}
}

func TestIsHomebrewInstall_GoInstall(t *testing.T) {
	result := update.TestableIsHomebrewInstall("/home/user/go/bin/kh")
	if result {
		t.Error("expected false for go install path")
	}
}

func TestNewUpdateCmd_Structure(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}
	cmd := update.NewUpdateCmd(f)

	if cmd.Use != "update" {
		t.Errorf("expected Use 'update', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}
	if cmd.Example == "" {
		t.Error("expected non-empty Example text")
	}
}

func TestUpdateCmd_HomebrewDelegate(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}

	update.HomebrewCheckerFunc = func() bool { return true }
	t.Cleanup(func() { update.HomebrewCheckerFunc = update.DefaultHomebrewChecker })

	cmd := update.NewUpdateCmd(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Homebrew") {
		t.Errorf("expected Homebrew message in output, got: %q", out)
	}
	if !strings.Contains(out, "brew upgrade kh") {
		t.Errorf("expected 'brew upgrade kh' in output, got: %q", out)
	}
}

func TestUpdateCmd_AlreadyOnLatest(t *testing.T) {
	ios, buf, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}

	update.HomebrewCheckerFunc = func() bool { return false }
	update.DetectLatestFunc = func(slug string) (*selfupdate.Release, bool, error) {
		return nil, false, nil
	}
	t.Cleanup(func() {
		update.HomebrewCheckerFunc = update.DefaultHomebrewChecker
		update.DetectLatestFunc = update.DefaultDetectLatest
	})

	cmd := update.NewUpdateCmd(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No release found") {
		t.Errorf("expected 'No release found' in output, got: %q", out)
	}
}
