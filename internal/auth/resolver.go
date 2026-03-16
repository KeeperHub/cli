package auth

import (
	"os"

	"github.com/keeperhub/cli/internal/config"
)

// ResolveToken resolves the auth token for host using the priority chain:
// 1. KH_API_KEY environment variable
// 2. hosts.yml token field
// 3. OS keyring (legacy fallback -- tokens are migrated to hosts.yml on login)
// Returns a ResolvedToken with Method set to AuthMethodNone if no token found.
func ResolveToken(host string) (ResolvedToken, error) {
	if apiKey := os.Getenv("KH_API_KEY"); apiKey != "" {
		return ResolvedToken{Token: apiKey, Method: AuthMethodAPIKey, Host: host}, nil
	}

	hosts, err := config.ReadHosts()
	if err != nil {
		return ResolvedToken{}, err
	}
	if entry, ok := hosts.HostEntry(host); ok && entry.Token != "" {
		return ResolvedToken{Token: entry.Token, Method: AuthMethodToken, Host: host}, nil
	}

	// Legacy fallback: check OS keyring for tokens stored before the hosts.yml migration.
	token, err := GetToken(host)
	if err != nil {
		return ResolvedToken{}, err
	}
	if token != "" {
		return ResolvedToken{Token: token, Method: AuthMethodToken, Host: host}, nil
	}

	return ResolvedToken{Method: AuthMethodNone, Host: host}, nil
}
