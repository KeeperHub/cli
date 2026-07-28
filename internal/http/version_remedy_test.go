package khhttp_test

// Updates are manual and unprompted, so the KH-Minimum-CLI-Version warning is
// the only signal an out-of-date CLI ever gets. It has to name the remedy, not
// just state the mismatch.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doWithMinimumVersion(t *testing.T, appVersion, minimum string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if minimum != "" {
			w.Header().Set("KH-Minimum-CLI-Version", minimum)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ios, _, errBuf, _ := iostreams.Test()
	client := khhttp.NewClient(khhttp.ClientOptions{
		Host:       srv.URL,
		AppVersion: appVersion,
		IOStreams:  ios,
	})

	req, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	return errBuf.String()
}

func TestCheckVersion_OutdatedWarningNamesTheRemedy(t *testing.T) {
	out := doWithMinimumVersion(t, "0.3.0", "0.11.1")

	assert.Contains(t, out, "outdated")
	assert.Contains(t, out, "0.11.1")
	// Stating the mismatch without the remedy leaves the user to guess.
	assert.Contains(t, out, "kh update")
}

func TestCheckVersion_SilentWhenCurrent(t *testing.T) {
	assert.Empty(t, doWithMinimumVersion(t, "0.12.0", "0.11.1"))
}

func TestCheckVersion_SilentWhenServerAdvertisesNothing(t *testing.T) {
	assert.Empty(t, doWithMinimumVersion(t, "0.3.0", ""))
}
