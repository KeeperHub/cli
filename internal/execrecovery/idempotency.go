package execrecovery

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const IdempotencyHeader = "Idempotency-Key"

// NewIdempotencyKey returns a random UUID-like key for a single write intent.
// Callers that retry the same intent across process restarts must persist or
// derive a stable key instead (see docs.keeperhub.com/api/direct-execution).
func NewIdempotencyKey() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generating idempotency key: %w", err)
	}
	// UUID v4 layout bits are not required by the API; hex is sufficient.
	return hex.EncodeToString(b[:]), nil
}

// ResolveIdempotencyKey returns explicit if non-empty, otherwise a new key.
func ResolveIdempotencyKey(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	return NewIdempotencyKey()
}
