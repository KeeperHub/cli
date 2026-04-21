package wallet_test

import (
	"testing"

	"github.com/keeperhub/cli/cmd/wallet"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAgenticFactory builds a Factory suitable for the npx-wrapper subcommands.
// HTTPClient is unused by these commands but stubbed for interface completeness.
func newAgenticFactory(ios *iostreams.IOStreams) *cmdutil.Factory {
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:  ios,
		HTTPClient: func() (*khhttp.Client, error) {
			return khhttp.NewClient(khhttp.ClientOptions{Host: "https://app.keeperhub.com", AppVersion: "1.0.0"}), nil
		},
		Config: func() (config.Config, error) {
			return config.Config{DefaultHost: "app.keeperhub.com"}, nil
		},
	}
}

func TestNewAddCmd_Help(t *testing.T) {
	ios, outBuf, _, _ := iostreams.Test()
	f := newAgenticFactory(ios)
	root := wallet.NewWalletCmd(f)
	// Route Cobra's help output into our captured buffer.
	root.SetOut(outBuf)
	root.SetErr(outBuf)
	root.SetArgs([]string{"add", "--help"})
	err := root.Execute()
	require.NoError(t, err)
	out := outBuf.String()
	assert.Contains(t, out, "agentic wallet", "help should describe agentic wallet, not creator wallet")
	assert.Contains(t, out, "npx @keeperhub/wallet", "help should reference the underlying npm package")
}

func TestNewInfoCmd_Help(t *testing.T) {
	ios, outBuf, _, _ := iostreams.Test()
	f := newAgenticFactory(ios)
	root := wallet.NewWalletCmd(f)
	root.SetOut(outBuf)
	root.SetErr(outBuf)
	root.SetArgs([]string{"info", "--help"})
	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "subOrgId")
}

func TestNewFundCmd_Help(t *testing.T) {
	ios, outBuf, _, _ := iostreams.Test()
	f := newAgenticFactory(ios)
	root := wallet.NewWalletCmd(f)
	root.SetOut(outBuf)
	root.SetErr(outBuf)
	root.SetArgs([]string{"fund", "--help"})
	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "Coinbase Onramp")
}

func TestNewLinkCmd_RequiresSessionCookie(t *testing.T) {
	ios, outBuf, errBuf, _ := iostreams.Test()
	f := newAgenticFactory(ios)
	root := wallet.NewWalletCmd(f)
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	root.SetArgs([]string{"link"})
	// Ensure KH_SESSION_COOKIE is unset for this test.
	t.Setenv("KH_SESSION_COOKIE", "")
	// Silence Cobra's auto-usage printing so the assertion is against the RunE error.
	root.SilenceUsage = true
	root.SilenceErrors = true
	err := root.Execute()
	require.Error(t, err)
	combined := err.Error() + errBuf.String() + outBuf.String()
	assert.Contains(t, combined, "KH_SESSION_COOKIE")
}

func TestWalletCmd_PreservesCreatorWalletSubcommands(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := newAgenticFactory(ios)
	root := wallet.NewWalletCmd(f)
	// Walk the direct subcommand list; assert balance + tokens still present.
	subs := map[string]bool{}
	for _, c := range root.Commands() {
		subs[c.Name()] = true
	}
	for _, expected := range []string{"balance", "tokens", "add", "info", "fund", "link"} {
		assert.True(t, subs[expected], "expected subcommand %q on kh wallet", expected)
	}
}
