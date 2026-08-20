package khhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newErrorResponse(t *testing.T, status int, body string) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	rec.WriteHeader(status)
	_, err := rec.WriteString(body)
	require.NoError(t, err)
	return rec.Result()
}

func TestNewAPIError_IncludesDetail(t *testing.T) {
	resp := newErrorResponse(t, http.StatusBadRequest, `{"error":"invalid_input","detail":"limit must be <= 200","request_id":"8d848585-2dc0-4f83-9bb6-ab6d12fe1e70"}`)

	apiErr := khhttp.NewAPIError(resp)

	assert.Equal(t, "invalid_input: limit must be <= 200", apiErr.Message)
	assert.Equal(t, "8d848585-2dc0-4f83-9bb6-ab6d12fe1e70", apiErr.RequestID)
	assert.Contains(t, apiErr.Error(), "invalid_input: limit must be <= 200")
	assert.Contains(t, apiErr.Error(), "8d848585-2dc0-4f83-9bb6-ab6d12fe1e70")
}

func TestNewAPIError_NoDetailFallsBackToErrorField(t *testing.T) {
	resp := newErrorResponse(t, http.StatusBadRequest, `{"error":"invalid_input"}`)

	apiErr := khhttp.NewAPIError(resp)

	assert.Equal(t, "invalid_input", apiErr.Message)
	assert.Empty(t, apiErr.RequestID)
	assert.Equal(t, "HTTP 400: invalid_input", apiErr.Error())
}

func TestNewAPIError_IncludesHint(t *testing.T) {
	resp := newErrorResponse(t, http.StatusPreconditionFailed, `{"error":"wallet_not_configured","detail":"No wallet provisioned for chain 8453 in org X","hint":"POST /api/integrations/wallet to provision","request_id":"req_abc"}`)

	apiErr := khhttp.NewAPIError(resp)

	assert.Equal(t, "POST /api/integrations/wallet to provision", apiErr.Hint)
	assert.Contains(t, apiErr.Error(), "hint: POST /api/integrations/wallet to provision")
}

func TestNewAPIError_DetailOnlyBody(t *testing.T) {
	resp := newErrorResponse(t, http.StatusBadRequest, `{"detail":"something went wrong"}`)

	apiErr := khhttp.NewAPIError(resp)

	assert.Equal(t, "something went wrong", apiErr.Message)
}

func TestNewAPIError_NonJSONBody(t *testing.T) {
	resp := newErrorResponse(t, http.StatusInternalServerError, "internal server error")

	apiErr := khhttp.NewAPIError(resp)

	assert.Equal(t, "internal server error", apiErr.Message)
}

func TestNewAPIError_EmptyBodyUsesStatusText(t *testing.T) {
	resp := newErrorResponse(t, http.StatusInternalServerError, "")

	apiErr := khhttp.NewAPIError(resp)

	assert.Equal(t, http.StatusText(http.StatusInternalServerError), apiErr.Message)
}
