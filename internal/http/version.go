package khhttp

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// MinimumVersionHeader is the response header the server uses to advertise
// the oldest CLI version it considers supported (see keeperhub's
// lib/cli-version.ts). Shared so every call site checks the same header.
const MinimumVersionHeader = "KH-Minimum-CLI-Version"

// semverLessThan returns true if current is strictly less than minimum.
// Returns false if current is "dev" or unparseable -- dev builds never trigger warnings.
func semverLessThan(current, minimum string) bool {
	return SemverLessThan(current, minimum)
}

// SemverLessThan is the exported form used in tests.
func SemverLessThan(current, minimum string) bool {
	if current == "dev" {
		return false
	}

	cv, err := parseSemver(current)
	if err != nil {
		return false
	}

	mv, err := parseSemver(minimum)
	if err != nil {
		return false
	}

	if cv[0] != mv[0] {
		return cv[0] < mv[0]
	}
	if cv[1] != mv[1] {
		return cv[1] < mv[1]
	}
	return cv[2] < mv[2]
}

func parseSemver(v string) ([3]int, error) {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("invalid semver: %s", v)
	}

	var result [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return [3]int{}, fmt.Errorf("invalid semver segment %q in %s: %w", part, v, err)
		}
		result[i] = n
	}
	return result, nil
}

// checkVersion inspects the KH-Minimum-CLI-Version response header.
// If the current version is older than the minimum, it writes a warning to errOut.
func checkVersion(current string, resp *http.Response, errOut io.Writer) {
	if resp == nil {
		return
	}
	minimum := resp.Header.Get(MinimumVersionHeader)
	if minimum == "" {
		return
	}
	if semverLessThan(current, minimum) {
		// Updates are manual and unprompted, so this line is the only signal an
		// out-of-date CLI gets. Name the remedy rather than only the mismatch.
		fmt.Fprintf(errOut, "warning: your CLI version (%s) is outdated; minimum required is %s. Run: kh update\n", current, minimum)
	}
}
