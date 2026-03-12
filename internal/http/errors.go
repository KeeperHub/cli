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
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// NewAPIError reads the response body and constructs an APIError.
// It attempts to extract a JSON "error" or "message" field for the message.
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

	message := extractJSONMessage(body)
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
	}
}

func extractJSONMessage(body []byte) string {
	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	if payload.Error != "" {
		return payload.Error
	}
	return payload.Message
}
