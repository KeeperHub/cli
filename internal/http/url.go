package khhttp

import "strings"

// BuildBaseURL normalises a host string into a base URL.
// If host already starts with http:// or https://, it is returned as-is
// (with any trailing slash stripped). Otherwise https:// is prepended.
func BuildBaseURL(host string) string {
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return strings.TrimRight(host, "/")
	}
	return "https://" + strings.TrimRight(host, "/")
}
