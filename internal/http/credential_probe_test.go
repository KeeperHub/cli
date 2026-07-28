package khhttp_test

// Both the doctor auth check and API-key validation resolve their endpoint from
// CredentialProbePath, so the semantic requirement - that the endpoint actually
// requires authentication - can only be pinned on the constant itself. A test at
// either call site would simply follow the constant wherever it points, which is
// how the original defect went unnoticed.

import (
	"testing"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/stretchr/testify/assert"
)

// Endpoints known to answer 200 to anonymous callers. Probing any of them
// reports every caller as authenticated.
//
//	/api/workflows        - resolves auth with required:false, returns an empty list
//	/api/auth/get-session - returns a null session rather than 401
//	/api/openapi          - public schema
func TestCredentialProbePath_IsNotAnonymousTolerant(t *testing.T) {
	anonymousTolerant := []string{
		"/api/workflows",
		"/api/auth/get-session",
		"/api/openapi",
	}

	for _, path := range anonymousTolerant {
		assert.NotEqual(
			t,
			path,
			khhttp.CredentialProbePath,
			"%s answers 200 to anonymous callers, so probing it cannot detect a bad credential",
			path,
		)
	}
}

func TestCredentialProbePath_IsAnAPIRoute(t *testing.T) {
	assert.Equal(t, "/api/projects", khhttp.CredentialProbePath)
}
