package doctor_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keeperhub/cli/cmd/doctor"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newDoctorFactory builds a Factory backed by a test HTTP server.
func newDoctorFactory(ios *iostreams.IOStreams, svr *httptest.Server) *cmdutil.Factory {
	return &cmdutil.Factory{
		AppVersion: "1.2.3",
		IOStreams:   ios,
		Config: func() (config.Config, error) {
			return config.Config{DefaultHost: svr.URL}, nil
		},
		HTTPClient: func() (*khhttp.Client, error) {
			return khhttp.NewClient(khhttp.ClientOptions{
				AppVersion: "1.2.3",
				IOStreams:   ios,
			}), nil
		},
	}
}

// TestDoctorCmd_AllPass verifies that when all endpoints succeed,
// all 6 check names appear in output and the command exits 0.
func TestDoctorCmd_AllPass(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/user/wallet/balances":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"address":"0xABCD1234","balances":[]}`))
		case "/api/billing/subscription":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"limits":{"spendCap":100}}`))
		case "/api/chains":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":1,"name":"Ethereum"},{"id":137,"name":"Polygon"}]`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newDoctorFactory(ios, svr)

	tc := doctor.NewTestableCmd(f)
	err := tc.Execute([]string{})

	require.NoError(t, err)
	out := outBuf.String()
	assert.Contains(t, out, "Auth")
	assert.Contains(t, out, "API")
	assert.Contains(t, out, "Wallet")
	assert.Contains(t, out, "Spend Cap")
	assert.Contains(t, out, "Chains")
	assert.Contains(t, out, "CLI Version")
}

// TestDoctorCmd_OneFail verifies that a failing check causes a non-zero exit.
func TestDoctorCmd_OneFail(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/health":
			// 500 -> API check reports [fail]
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"service unavailable"}`))
		case "/api/user/wallet/balances":
			w.WriteHeader(http.StatusUnauthorized)
		case "/api/billing/subscription":
			w.WriteHeader(http.StatusNotFound)
		case "/api/chains":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":1}]`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newDoctorFactory(ios, svr)

	tc := doctor.NewTestableCmd(f)
	err := tc.Execute([]string{})

	require.Error(t, err, "should return error when any check fails")
	out := outBuf.String()
	assert.Contains(t, out, "[fail]")
}

// TestDoctorCmd_WarnOnly verifies that warnings alone yield exit 0.
func TestDoctorCmd_WarnOnly(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/user/wallet/balances":
			// 401 -> wallet warns, not fails
			w.WriteHeader(http.StatusUnauthorized)
		case "/api/billing/subscription":
			// 404 -> billing warns (billing not enabled)
			w.WriteHeader(http.StatusNotFound)
		case "/api/chains":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":1}]`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newDoctorFactory(ios, svr)

	tc := doctor.NewTestableCmd(f)
	err := tc.Execute([]string{})

	// Warnings only -> exit 0
	require.NoError(t, err)
	out := outBuf.String()
	assert.NotContains(t, out, "[fail]")
	assert.Contains(t, out, "[warn]")
}

// TestDoctorCmd_JSON verifies --json outputs a JSON array with exactly 6 objects.
func TestDoctorCmd_JSON(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/user/wallet/balances":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"address":"0xDEAD","balances":[]}`))
		case "/api/billing/subscription":
			w.WriteHeader(http.StatusNotFound)
		case "/api/chains":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":1},{"id":2},{"id":3}]`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newDoctorFactory(ios, svr)

	tc := doctor.NewTestableCmd(f)
	err := tc.Execute([]string{"--json"})
	require.NoError(t, err)

	var results []map[string]interface{}
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &results), "output must be valid JSON array")
	assert.Len(t, results, 6, "should have exactly 6 check results")
	for _, r := range results {
		assert.Contains(t, r, "name", "each result must have 'name'")
		assert.Contains(t, r, "status", "each result must have 'status'")
		assert.Contains(t, r, "message", "each result must have 'message'")
	}
}

// TestDoctorCmd_Timeout verifies that a slow endpoint is bounded by 5s timeout.
func TestDoctorCmd_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/user/wallet/balances":
			w.WriteHeader(http.StatusUnauthorized)
		case "/api/billing/subscription":
			w.WriteHeader(http.StatusNotFound)
		case "/api/chains":
			// Hang longer than the 5s per-check timeout.
			time.Sleep(7 * time.Second)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newDoctorFactory(ios, svr)

	start := time.Now()
	tc := doctor.NewTestableCmd(f)
	_ = tc.Execute([]string{})
	elapsed := time.Since(start)

	out := outBuf.String()
	assert.Contains(t, out, "[fail]", "timed-out chain check must report [fail]")
	assert.Less(t, elapsed, 6*time.Second, "must finish before the 7s sleep completes")
}

// TestDoctorCmd_Output verifies all 6 check names appear in non-JSON output.
func TestDoctorCmd_Output(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer svr.Close()

	ios, outBuf, _, _ := iostreams.Test()
	f := newDoctorFactory(ios, svr)

	tc := doctor.NewTestableCmd(f)
	_ = tc.Execute([]string{})

	out := outBuf.String()
	for _, name := range []string{"Auth", "API", "Wallet", "Spend Cap", "Chains", "CLI Version"} {
		assert.Contains(t, out, name, "output must include check name: %s", name)
	}
}

// TestDoctorCmd_NoLocalJSONFlag verifies the local --json flag was removed.
// The flag must be inherited from root, not defined locally on the command.
func TestDoctorCmd_NoLocalJSONFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{AppVersion: "1.0.0", IOStreams: ios}
	cmd := doctor.NewDoctorCmd(f)
	localFlag := cmd.Flags().Lookup("json")
	assert.Nil(t, localFlag, "doctor must not define --json locally; use root persistent flag")
}

// TestDoctorCmd_ExitCodeOnFail verifies SilentError is returned on [fail] checks.
func TestDoctorCmd_ExitCodeOnFail(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer svr.Close()

	ios, _, _, _ := iostreams.Test()
	f := newDoctorFactory(ios, svr)

	tc := doctor.NewTestableCmd(f)
	err := tc.Execute([]string{})
	require.Error(t, err)

	var silentErr cmdutil.SilentError
	require.ErrorAs(t, err, &silentErr, "error must be SilentError so root does not double-print")
	assert.Equal(t, 1, cmdutil.ExitCodeForError(err))
}
