package khhttp

import (
	"net/http"
	"os"

	"github.com/keeperhub/cli/internal/config"
)

// Cloudflare Access env vars. CF_ACCESS_CLIENT_ID + CF_ACCESS_CLIENT_SECRET
// is the service-token pair used by CI; CF_AUTHORIZATION carries the
// cf_authorization JWT minted by `cloudflared access login` for interactive
// devs. Both are sent when set; CF Access accepts either.
const (
	envCFAccessClientID     = "CF_ACCESS_CLIENT_ID"
	envCFAccessClientSecret = "CF_ACCESS_CLIENT_SECRET"
	envCFAuthorization      = "CF_AUTHORIZATION"
)

// MergeCloudflareAccessEnv returns base with Cloudflare Access headers added
// from the environment. Env values take precedence over base entries with the
// same key (matching the KH_API_KEY > hosts.yml precedence used elsewhere).
//
// Service-token headers are only added when both ID and secret are set, to
// avoid sending half a credential pair.
func MergeCloudflareAccessEnv(base map[string]string) map[string]string {
	id := os.Getenv(envCFAccessClientID)
	secret := os.Getenv(envCFAccessClientSecret)
	cookie := os.Getenv(envCFAuthorization)

	if id == "" && secret == "" && cookie == "" {
		return base
	}

	out := make(map[string]string, len(base)+3)
	for k, v := range base {
		out[k] = v
	}
	if id != "" && secret != "" {
		out["CF-Access-Client-Id"] = id
		out["CF-Access-Client-Secret"] = secret
	}
	if cookie != "" {
		out["Cookie"] = "CF_Authorization=" + cookie
	}
	return out
}

// HeadersForHost returns the set of headers to apply to a request targeting
// host: per-host entries from hosts.yml with Cloudflare Access env vars merged
// on top. Used by code paths that build their own *http.Client and bypass the
// shared khhttp.Client (auth token validation, doctor checks).
func HeadersForHost(host string) map[string]string {
	hosts, err := config.ReadHosts()
	if err != nil {
		return MergeCloudflareAccessEnv(nil)
	}
	entry, _ := hosts.HostEntry(host)
	return MergeCloudflareAccessEnv(entry.Headers)
}

// ApplyHostHeaders sets per-host headers (hosts.yml + CF env) on req.
func ApplyHostHeaders(req *http.Request, host string) {
	for k, v := range HeadersForHost(host) {
		req.Header.Set(k, v)
	}
}
