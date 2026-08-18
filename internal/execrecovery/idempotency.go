package execrecovery

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const IdempotencyHeader = "Idempotency-Key"

// Codes from KeeperHub/keeperhub lib/idempotency.ts idempotencyEarlyResponse.
// Both map to HTTP 409; only `code` distinguishes them.
const (
	CodeIdempotencyConflict   = "idempotency_conflict"
	CodeIdempotencyInProgress = "idempotency_in_progress"
)

// IdempotencyBody is the 409 JSON from idempotencyEarlyResponse.
type IdempotencyBody struct {
	Error               string  `json:"error"`
	Code                string  `json:"code"`
	Retryable           *bool   `json:"retryable"`
	OriginalExecutionID *string `json:"originalExecutionId"`
}

// ParseIdempotencyBody decodes a 409 body. ok is true when `code` is one of
// the two idempotency codes, or when retryable is present as a fallback.
func ParseIdempotencyBody(body []byte) (IdempotencyBody, bool) {
	var b IdempotencyBody
	if err := json.Unmarshal(body, &b); err != nil {
		return IdempotencyBody{}, false
	}
	code := strings.ToLower(strings.TrimSpace(b.Code))
	b.Code = code
	switch code {
	case CodeIdempotencyConflict, CodeIdempotencyInProgress:
		return b, true
	}
	if b.Retryable != nil {
		return b, true
	}
	return b, false
}

// IsInProgress is true when the same key is already processing.
// The client must retry that key and must not mint a new one.
func (b IdempotencyBody) IsInProgress() bool {
	if b.Code == CodeIdempotencyInProgress {
		return true
	}
	return b.Code == "" && b.Retryable != nil && *b.Retryable
}

// IsConflict is true when the key is bound to a different payload.
// The client must fail closed and must not rotate the key.
func (b IdempotencyBody) IsConflict() bool {
	if b.Code == CodeIdempotencyConflict {
		return true
	}
	return b.Code == "" && b.Retryable != nil && !*b.Retryable
}

// ConflictError is a 409 idempotency_conflict.
type ConflictError struct {
	Body IdempotencyBody
	Key  string
}

func (e ConflictError) Error() string {
	msg := e.Body.Error
	if msg == "" {
		msg = "Idempotency-Key was reused with a different request payload"
	}
	if e.Body.OriginalExecutionID != nil && *e.Body.OriginalExecutionID != "" {
		msg = fmt.Sprintf("%s (originalExecutionId=%s)", msg, *e.Body.OriginalExecutionID)
	}
	if e.Key != "" {
		msg = fmt.Sprintf("%s; do not retry with a new key for the same intent (Idempotency-Key %s)", msg, e.Key)
	} else {
		msg = msg + "; do not retry with a new Idempotency-Key for the same intent"
	}
	return msg
}

// InProgressTimeoutError is returned when 409 idempotency_in_progress
// outlives the wait budget. The same key must be reused on the next attempt.
type InProgressTimeoutError struct {
	Key string
}

func (e InProgressTimeoutError) Error() string {
	return fmt.Sprintf("Idempotency-Key %s is still in progress; retry with --idempotency-key %s (do not rotate the key)", e.Key, e.Key)
}

// NewIdempotencyKey returns a random UUID-like key for a single write intent.
// Callers that retry the same intent across process restarts must persist or
// derive a stable key instead (see docs.keeperhub.com/api/direct-execution).
func NewIdempotencyKey() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generating idempotency key: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// ResolveIdempotencyKey returns explicit if non-empty, otherwise a new key.
func ResolveIdempotencyKey(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	return NewIdempotencyKey()
}
