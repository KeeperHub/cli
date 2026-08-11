// Package execrecovery implements the execution-recovery contract (R1–R6)
// used by fixture conformance tests and by direct-execution wait paths.
package execrecovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Outcome is the classified result of one status observation.
type Outcome string

const (
	OutcomePending     Outcome = "pending"
	OutcomeSuccess     Outcome = "success"
	OutcomeFailure     Outcome = "failure"
	OutcomeMalformed   Outcome = "malformed"
	OutcomeRateLimited Outcome = "rate_limited"
)

// Options controls classification strictness.
type Options struct {
	// RequireChainEvidence enables R2 strict mode: completed without a
	// verified successful receipt is Failure, not Success.
	RequireChainEvidence bool
}

// Receipt is a chain-re-fetched proof entry (direct-execution status API).
type Receipt struct {
	Hash          string `json:"hash"`
	ChainID       int64  `json:"chainId"`
	Verified      bool   `json:"verified"`
	ReceiptStatus string `json:"receiptStatus"`
}

// DirectStatus is the flat wire shape of GET /api/execute/{id}/status.
type DirectStatus struct {
	ExecutionID     string    `json:"executionId"`
	Status          string    `json:"status"`
	Type            string    `json:"type"`
	TransactionHash *string   `json:"transactionHash"`
	TransactionLink *string   `json:"transactionLink"`
	Result          any       `json:"result"`
	Error           *string   `json:"error"`
	CreatedAt       string    `json:"createdAt"`
	CompletedAt     *string   `json:"completedAt"`
	Receipts        []Receipt `json:"receipts,omitempty"`
}

// Sample is one HTTP observation of an execution status endpoint.
type Sample struct {
	HTTPStatus int
	Body       []byte
}

// Classify maps one status observation to an Outcome.
//
// Vocabulary note: direct-execution statuses are pending|running|completed|failed
// (and transport-level not_found). Workflow run statuses (success|error|cancelled)
// belong to a different API and must not be fed here — see Vocabulary().
func Classify(sample Sample, opts Options) (Outcome, string) {
	if sample.HTTPStatus == http.StatusTooManyRequests {
		return OutcomeRateLimited, "HTTP 429"
	}

	// Cold-start / missing: callers may poll again (R6). Terminal failure is a
	// poll-budget decision, not Classify's.
	if sample.HTTPStatus == http.StatusNotFound {
		return OutcomePending, "not_found"
	}

	if sample.HTTPStatus != 0 && sample.HTTPStatus != http.StatusOK && sample.HTTPStatus != http.StatusAccepted {
		// Non-404 errors are failures for a status read.
		if sample.HTTPStatus >= 400 {
			return OutcomeFailure, fmt.Sprintf("HTTP %d", sample.HTTPStatus)
		}
	}

	if len(sample.Body) == 0 {
		return OutcomeMalformed, "empty body"
	}

	trimmed := strings.TrimSpace(string(sample.Body))
	if !json.Valid([]byte(trimmed)) {
		return OutcomeMalformed, "unparseable body"
	}

	var st DirectStatus
	if err := json.Unmarshal([]byte(trimmed), &st); err != nil {
		return OutcomeMalformed, "json decode failed"
	}

	status := strings.ToLower(strings.TrimSpace(st.Status))
	if status == "" {
		// Unrecognised schema: valid JSON but no status field.
		return OutcomeMalformed, "missing status field"
	}

	switch status {
	case "pending", "running", "queued", "unconfirmed":
		return OutcomePending, status
	case "not_found":
		return OutcomePending, status
	case "failed", "error", "cancelled":
		return OutcomeFailure, status
	case "completed", "success":
		return classifyCompleted(st, opts)
	default:
		return OutcomeMalformed, "unrecognised status: " + status
	}
}

func classifyCompleted(st DirectStatus, opts Options) (Outcome, string) {
	if hasRevertedReceipt(st.Receipts) {
		return OutcomeFailure, "receiptStatus=reverted"
	}

	if hasVerifiedSuccess(st.Receipts) {
		return OutcomeSuccess, "verified successful receipt"
	}

	if opts.RequireChainEvidence {
		if st.TransactionHash == nil || strings.TrimSpace(*st.TransactionHash) == "" {
			return OutcomeFailure, "completed without transaction hash"
		}
		if len(st.Receipts) == 0 {
			return OutcomeFailure, "completed without verified successful receipt"
		}
		return OutcomeFailure, "no verified successful receipt"
	}

	// Compatible default: completed without receipts is still Success for
	// callers that have not opted into R2 strict mode (see --require-verified).
	return OutcomeSuccess, "completed"
}

func hasRevertedReceipt(receipts []Receipt) bool {
	for _, r := range receipts {
		if strings.EqualFold(r.ReceiptStatus, "reverted") {
			return true
		}
	}
	return false
}

func hasVerifiedSuccess(receipts []Receipt) bool {
	for _, r := range receipts {
		if r.Verified && strings.EqualFold(r.ReceiptStatus, "success") {
			return true
		}
	}
	return false
}

// HasRevertedReceipt reports whether any receipt is an onchain revert.
func HasRevertedReceipt(receipts []Receipt) bool {
	return hasRevertedReceipt(receipts)
}
