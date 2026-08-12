package wallet

import (
	"os/exec"
	"testing"

	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

func TestAgenticWrapperInvokesExplicitBinary(t *testing.T) {
	origExec := execCommand
	origLook := lookPath
	t.Cleanup(func() {
		execCommand = origExec
		lookPath = origLook
	})

	lookPath = func(string) (string, error) { return "/usr/bin/npx", nil }

	var gotArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		gotArgs = args
		return exec.Command("true")
	}

	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:  ios,
		HTTPClient: func() (*khhttp.Client, error) {
			return khhttp.NewClient(khhttp.ClientOptions{Host: "https://app.keeperhub.com", AppVersion: "1.0.0"}), nil
		},
		Config: func() (config.Config, error) {
			return config.Config{DefaultHost: "app.keeperhub.com"}, nil
		},
	}

	_ = runNpxWallet(f, &cobra.Command{}, "info", nil)

	want := []string{"-p", "@keeperhub/wallet", "keeperhub-wallet"}
	for i, w := range want {
		if len(gotArgs) <= i || gotArgs[i] != w {
			t.Fatalf("argv[%d] = %q, want %q (full argv: %v)", i, func() string {
				if len(gotArgs) > i {
					return gotArgs[i]
				}
				return "<missing>"
			}(), w, gotArgs)
		}
	}
}
