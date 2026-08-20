package khhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError wraps an HTTP error response with status code and body.
type APIError struct {
	StatusCode int
	Body       []byte
	Message    string
	RequestID  string
	Hint       string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	msg := fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
	if e.Hint != "" {
		msg += fmt.Sprintf(" (hint: %s)", e.Hint)
	}
	if e.RequestID != "" {
		msg += fmt.Sprintf(" (request_id: %s)", e.RequestID)
	}
	return msg
}

// NewAPIError reads the response body and constructs an APIError.
// It attempts to extract JSON "error"/"message" and "detail" fields for the
// message, plus "hint" and "request_id" fields for support and remediation
// purposes.
func NewAPIError(resp *http.Response) *APIError {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       nil,
			Message:    http.StatusText(resp.StatusCode),
		}
	}

	message, hint, requestID := extractJSONMessage(body)
	if message == "" {
		message = string(body)
	}
	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Body:       body,
		Message:    message,
		RequestID:  requestID,
		Hint:       hint,
	}
}

// extractJSONMessage returns a human-readable message built from the
// "error"/"message" and "detail" fields of a JSON error body, along with the
// "hint" and "request_id" fields if present. The "detail" field carries the
// actionable, human-readable explanation the API sends; without it, callers
// only ever see an opaque error code like "invalid_input". "hint" carries a
// suggested remediation (e.g. "POST /api/integrations/wallet to provision").
func extractJSONMessage(body []byte) (message, hint, requestID string) {
	var payload struct {
		Error     string `json:"error"`
		Message   string `json:"message"`
		Detail    string `json:"detail"`
		Hint      string `json:"hint"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", ""
	}

	message = payload.Error
	if message == "" {
		message = payload.Message
	}
	if payload.Detail != "" {
		if message != "" {
			message += ": " + payload.Detail
		} else {
			message = payload.Detail
		}
	}
	return message, payload.Hint, payload.RequestID
}
