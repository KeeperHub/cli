package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

// CheckResult holds the outcome of a single health check.
type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type checkFunc func(ctx context.Context, f *cmdutil.Factory) CheckResult

type namedCheck struct {
	name string
	fn   checkFunc
}

// NewDoctorCmd creates the doctor subcommand. The --json flag is NOT defined
// locally here -- it is inherited from root's persistent flags.
func NewDoctorCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "doctor",
		Short:   "Check CLI health",
		Aliases: []string{"doc"},
		Args:    cobra.NoArgs,
		Long: `Run diagnostic checks against your KeeperHub configuration and API
connectivity. Checks auth validity, API reachability, wallet status,
spend cap, chain availability, and CLI version.

See also: kh auth status, kh version`,
		Example: `  # Run all health checks
  kh doctor

  # Output results as JSON
  kh doctor --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd, f)
		},
	}

	// NOTE: --json is intentionally NOT defined here. It is a persistent flag
	// on the root command and is inherited by all subcommands automatically.
	// Adding it here would cause a cobra flag redefinition panic.

	return cmd
}

func runDoctor(cmd *cobra.Command, f *cmdutil.Factory) error {
	jsonMode, _ := cmd.Flags().GetBool("json")

	checks := []namedCheck{
		{name: "Auth", fn: checkAuth},
		{name: "API", fn: checkAPI},
		{name: "Wallet", fn: checkWallet},
		{name: "Spend Cap", fn: checkSpendCap},
		{name: "Chains", fn: checkChains},
		{name: "CLI Version", fn: checkCLIVersion},
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	results := make([]CheckResult, len(checks))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, c := range checks {
		wg.Add(1)
		go func(idx int, nc namedCheck) {
			defer wg.Done()
			checkCtx, checkCancel := context.WithTimeout(ctx, 5*time.Second)
			defer checkCancel()

			result := nc.fn(checkCtx, f)
			result.Name = nc.name

			mu.Lock()
			results[idx] = result
			if !jsonMode {
				fmt.Fprintf(f.IOStreams.Out, "[%s] %s: %s\n", result.Status, result.Name, result.Message)
			}
			mu.Unlock()
		}(i, c)
	}

	wg.Wait()

	if jsonMode {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IOStreams.Out, string(data))
	}

	for _, r := range results {
		if r.Status == "fail" {
			return cmdutil.SilentError{Err: errors.New("one or more health checks failed")}
		}
	}

	return nil
}

// getHTTPClient returns a plain *http.Client for diagnostic health checks.
// Doctor needs strict context cancellation semantics; we use a plain client
// with no retry so timeouts are respected immediately.
func getHTTPClient(_ *cmdutil.Factory) (*http.Client, error) {
	return &http.Client{}, nil
}

// getHost returns the configured host, falling back to the default.
// NOTE: cmd is not available in check functions, so we pass nil to ResolveHost.
// The --host flag is still respected via the KH_HOST env var and config fallback.
func getHost(f *cmdutil.Factory) (string, error) {
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	return cmdutil.ResolveHost(nil, cfg), nil
}

// doGet performs a GET request against url with the given context, returning
// the response. The caller is responsible for closing resp.Body.
func doGet(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

// isContextTimeout reports whether err resulted from a context timeout.
func isContextTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return strings.Contains(err.Error(), "context deadline exceeded")
}

// checkAuth checks whether a valid auth token is available for the configured host.
func checkAuth(ctx context.Context, f *cmdutil.Factory) CheckResult {
	host, err := getHost(f)
	if err != nil {
		return CheckResult{Status: "warn", Message: "could not read config"}
	}

	client, err := getHTTPClient(f)
	if err != nil {
		return CheckResult{Status: "warn", Message: "could not create HTTP client"}
	}

	url := khhttp.BuildBaseURL(host) + "/api/auth/get-session"
	resp, err := doGet(ctx, client, url)
	if err != nil {
		if isContextTimeout(err) {
			return CheckResult{Status: "warn", Message: "auth check timed out"}
		}
		return CheckResult{Status: "warn", Message: "not authenticated. Run: kh auth login"}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return CheckResult{Status: "warn", Message: "not authenticated. Run: kh auth login"}
	case http.StatusOK:
		return CheckResult{Status: "pass", Message: "authenticated"}
	default:
		return CheckResult{Status: "warn", Message: "could not verify auth status"}
	}
}

// checkAPI checks whether the KeeperHub API host is reachable, measuring latency.
func checkAPI(ctx context.Context, f *cmdutil.Factory) CheckResult {
	host, err := getHost(f)
	if err != nil {
		return CheckResult{Status: "fail", Message: "could not read config"}
	}

	client, err := getHTTPClient(f)
	if err != nil {
		return CheckResult{Status: "fail", Message: "could not create HTTP client"}
	}

	url := khhttp.BuildBaseURL(host) + "/api/health"
	start := time.Now()
	resp, err := doGet(ctx, client, url)
	latency := time.Since(start)

	if err != nil {
		if isContextTimeout(err) {
			return CheckResult{Status: "fail", Message: "unreachable (timeout after 5s)"}
		}
		return CheckResult{Status: "fail", Message: fmt.Sprintf("unreachable (%s)", stripURL(err.Error(), host))}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 500 {
		return CheckResult{Status: "fail", Message: fmt.Sprintf("server error (HTTP %d)", resp.StatusCode)}
	}

	return CheckResult{Status: "pass", Message: fmt.Sprintf("reachable (%dms)", latency.Milliseconds())}
}

// checkWallet checks whether the user's wallet is connected.
func checkWallet(ctx context.Context, f *cmdutil.Factory) CheckResult {
	host, err := getHost(f)
	if err != nil {
		return CheckResult{Status: "warn", Message: "could not check wallet"}
	}

	client, err := getHTTPClient(f)
	if err != nil {
		return CheckResult{Status: "warn", Message: "could not check wallet"}
	}

	url := khhttp.BuildBaseURL(host) + "/api/user/wallet/balances"
	resp, err := doGet(ctx, client, url)
	if err != nil {
		if isContextTimeout(err) {
			return CheckResult{Status: "warn", Message: "wallet check timed out"}
		}
		return CheckResult{Status: "warn", Message: "could not check wallet"}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return CheckResult{Status: "warn", Message: "requires authentication"}
	case http.StatusOK:
		var payload struct {
			Address string `json:"address"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return CheckResult{Status: "warn", Message: "could not parse wallet response"}
		}
		if payload.Address == "" {
			return CheckResult{Status: "warn", Message: "no wallet connected"}
		}
		short := abbreviateAddr(payload.Address)
		return CheckResult{Status: "pass", Message: fmt.Sprintf("connected (%s)", short)}
	default:
		_, _ = io.Copy(io.Discard, resp.Body)
		return CheckResult{Status: "warn", Message: "could not check wallet"}
	}
}

// checkSpendCap checks whether a spending cap is configured.
func checkSpendCap(ctx context.Context, f *cmdutil.Factory) CheckResult {
	host, err := getHost(f)
	if err != nil {
		return CheckResult{Status: "warn", Message: "could not check spend cap"}
	}

	client, err := getHTTPClient(f)
	if err != nil {
		return CheckResult{Status: "warn", Message: "could not check spend cap"}
	}

	url := khhttp.BuildBaseURL(host) + "/api/billing/subscription"
	resp, err := doGet(ctx, client, url)
	if err != nil {
		if isContextTimeout(err) {
			return CheckResult{Status: "warn", Message: "spend cap check timed out"}
		}
		return CheckResult{Status: "warn", Message: "could not check spend cap"}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return CheckResult{Status: "pass", Message: "billing not enabled (no cap needed)"}
	case http.StatusUnauthorized, http.StatusForbidden:
		_, _ = io.Copy(io.Discard, resp.Body)
		return CheckResult{Status: "warn", Message: "requires authentication"}
	case http.StatusOK:
		var payload struct {
			Limits struct {
				SpendCap *float64 `json:"spendCap"`
			} `json:"limits"`
			SpendCap *float64 `json:"spendCap"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return CheckResult{Status: "warn", Message: "could not parse billing response"}
		}
		cap := payload.SpendCap
		if cap == nil {
			cap = payload.Limits.SpendCap
		}
		if cap == nil || *cap == 0 {
			return CheckResult{Status: "warn", Message: "no spending cap set"}
		}
		return CheckResult{Status: "pass", Message: fmt.Sprintf("set ($%.0f)", *cap)}
	default:
		_, _ = io.Copy(io.Discard, resp.Body)
		return CheckResult{Status: "warn", Message: "could not check spend cap"}
	}
}

// checkChains checks whether the chain service is reachable.
func checkChains(ctx context.Context, f *cmdutil.Factory) CheckResult {
	host, err := getHost(f)
	if err != nil {
		return CheckResult{Status: "fail", Message: "could not check chain service"}
	}

	client, err := getHTTPClient(f)
	if err != nil {
		return CheckResult{Status: "fail", Message: "could not check chain service"}
	}

	url := khhttp.BuildBaseURL(host) + "/api/chains"
	resp, err := doGet(ctx, client, url)
	if err != nil {
		if isContextTimeout(err) {
			return CheckResult{Status: "fail", Message: "could not reach chain service (timeout after 5s)"}
		}
		return CheckResult{Status: "fail", Message: fmt.Sprintf("could not reach chain service (%s)", stripURL(err.Error(), host))}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return CheckResult{Status: "fail", Message: fmt.Sprintf("could not reach chain service (HTTP %d)", resp.StatusCode)}
	}

	var chains []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&chains); err != nil {
		return CheckResult{Status: "pass", Message: "reachable"}
	}

	return CheckResult{Status: "pass", Message: fmt.Sprintf("%d chains available", len(chains))}
}

// checkCLIVersion reports the current CLI version. Network-based latest-version
// checking is deferred to Phase 24; this check is always instantaneous.
func checkCLIVersion(_ context.Context, f *cmdutil.Factory) CheckResult {
	v := f.AppVersion
	if v == "" || v == "dev" || strings.HasPrefix(v, "v0.0.0") {
		return CheckResult{Status: "warn", Message: "development build"}
	}
	return CheckResult{Status: "pass", Message: fmt.Sprintf("v%s", strings.TrimPrefix(v, "v"))}
}

// abbreviateAddr shortens a long address like 0xABCD...1234.
func abbreviateAddr(addr string) string {
	if len(addr) <= 10 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-4:]
}

// stripURL removes the raw URL from an error message to avoid leaking test server addresses.
func stripURL(msg, host string) string {
	return strings.ReplaceAll(msg, khhttp.BuildBaseURL(host), "<host>")
}

// TestableCmd wraps the doctor command with a parent that carries root persistent
// flags (--json, --jq, --yes, --no-color, --host) so tests can exercise --json
// without spinning up a full root command.
type TestableCmd struct {
	f   *cmdutil.Factory
	ios *iostreams.IOStreams
}

// NewTestableCmd creates a TestableCmd for use in tests.
func NewTestableCmd(f *cmdutil.Factory) *TestableCmd {
	return &TestableCmd{f: f, ios: f.IOStreams}
}

// Execute runs the doctor command with the given args via a minimal parent
// that exposes root persistent flags.
func (tc *TestableCmd) Execute(args []string) error {
	parent := &cobra.Command{Use: "kh", SilenceErrors: true, SilenceUsage: true}
	parent.PersistentFlags().Bool("json", false, "Output as JSON")
	parent.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")
	parent.PersistentFlags().BoolP("yes", "y", false, "Skip confirmation prompts")
	parent.PersistentFlags().Bool("no-color", false, "Disable color output")
	parent.PersistentFlags().StringP("host", "H", "", "KeeperHub host")

	doctorCmd := NewDoctorCmd(tc.f)
	parent.AddCommand(doctorCmd)
	parent.SetArgs(append([]string{"doctor"}, args...))
	parent.SetOut(tc.ios.Out)
	parent.SetErr(tc.ios.ErrOut)
	return parent.Execute()
}
