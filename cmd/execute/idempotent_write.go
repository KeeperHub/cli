package execute

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/keeperhub/cli/internal/execrecovery"
	khhttp "github.com/keeperhub/cli/internal/http"
)

const inProgressBackoff = 200 * time.Millisecond
const inProgressBackoffMax = 2 * time.Second

// postIdempotentJSON POSTs body with a stable Idempotency-Key.
//
// HTTP 5xx retries are handled by the retryable client (same key).
// HTTP 409 is classified by body code from lib/idempotency.ts:
//
//	idempotency_in_progress -> retry the same key until deadline
//	idempotency_conflict    -> fail; never mint a new key
func postIdempotentJSON(client *khhttp.Client, url string, body []byte, idemKey string, deadline time.Time) (*http.Response, error) {
	backoff := inProgressBackoff
	for {
		req, err := client.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(execrecovery.IdempotencyHeader, idemKey)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusConflict {
			return resp, nil
		}

		raw, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading 409 body: %w", readErr)
		}

		info, ok := execrecovery.ParseIdempotencyBody(raw)
		if ok && info.IsInProgress() {
			if !deadline.IsZero() && time.Now().After(deadline) {
				return nil, execrecovery.InProgressTimeoutError{Key: idemKey}
			}
			time.Sleep(backoff)
			if backoff < inProgressBackoffMax {
				backoff *= 2
				if backoff > inProgressBackoffMax {
					backoff = inProgressBackoffMax
				}
			}
			continue
		}
		if ok && info.IsConflict() {
			return nil, execrecovery.ConflictError{Body: info, Key: idemKey}
		}

		msg := string(raw)
		if info.Error != "" {
			msg = info.Error
		}
		if msg == "" {
			msg = http.StatusText(http.StatusConflict)
		}
		return nil, &khhttp.APIError{StatusCode: http.StatusConflict, Body: raw, Message: msg}
	}
}
