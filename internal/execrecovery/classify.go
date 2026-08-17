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
	OutcomePending Outcome = "pending"
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
	// OutcomeUnconfirmed is broadcast-but-unreadable. It is neither success nor
	// failure: the transaction may already be on chain. Callers stop waiting and
	// report it; they must not poll it through and must not resubmit.
	OutcomeUnconfirmed  Outcome = "unconfirmed"
	OutcomeMalformed    Outcome = "malformed"
	OutcomeRateLimited  Outcome = "rate_limited"
	OutcomeUnrecognized Outcome = "unrecognized"
)

// Options controls classification strictness.
type Options struct {
	// RequireChainEvidence is a classifier-only option used by fixtures.
	// The shipped CLI wait paths do not set it. When true, completed
	// without a verified successful receipt is Failure.
	RequireChainEvidence bool
}

// Receipt is GET /api/execute/{id}/status receipts[]
// (DirectExecutionReceiptEntry in KeeperHub/keeperhub).
//
// receiptStatus values from lib/web3/verify-receipt.ts:
// success | reverted | not_found | timeout | safe_inner_failure.
// verifiedAt is required on the server type. chainId is optional.
type Receipt struct {
	Hash          string  `json:"hash"`
	ChainID       *int    `json:"chainId,omitempty"`
	Network       string  `json:"network,omitempty"`
	Verified      bool    `json:"verified"`
	ReceiptStatus string  `json:"receiptStatus"`
	BlockNumber   *int64  `json:"blockNumber,omitempty"`
	GasUsed       *string `json:"gasUsed,omitempty"`
	VerifiedAt    string  `json:"verifiedAt"`
}

// DirectStatus is the flat wire shape of GET /api/execute/{id}/status.
// Canonical type: cmd/execute aliases this as ExecStatusResponse.
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
// Direct-execution statuses are pending|running|unconfirmed|completed|failed
// (app/api/execute/_lib/types.ts). Workflow run statuses (success|error|cancelled)
// belong to a different API and must not be fed here — see Vocabulary().
//
// `unconfirmed` maps to OutcomeUnconfirmed, not OutcomePending: the server
// keeps reconciling that row, but a client must stop waiting on it rather than
// poll to a failure the chain never reported.
//
// An unknown future status is OutcomeUnrecognized (never success, never
// malformed) so a server addition does not look like a corrupt body.
func Classify(sample Sample, opts Options) (Outcome, string) {
	if sample.HTTPStatus == http.StatusTooManyRequests {
		return OutcomeRateLimited, "HTTP 429"
	}

	// Cold-start / missing: callers may poll again (R6). Terminal failure is a
	// poll-budget decision, not Classify's. The status endpoint answers 404
	// with {"error":"Execution not found"} and no status field.
	if sample.HTTPStatus == http.StatusNotFound {
		return OutcomePending, "http 404"
	}

	if sample.HTTPStatus != 0 && sample.HTTPStatus != http.StatusOK && sample.HTTPStatus != http.StatusAccepted {
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
		return OutcomeMalformed, "missing status field"
	}

	switch status {
	case "pending", "running":
		return OutcomePending, status
	case "unconfirmed":
		return OutcomeUnconfirmed, status
	case "failed":
		return OutcomeFailure, status
	case "completed":
		return classifyCompleted(st, opts)
	default:
		return OutcomeUnrecognized, "unrecognized status: " + status
	}
}

func classifyCompleted(st DirectStatus, opts Options) (Outcome, string) {
	if r := FirstBlockingReceipt(st.Receipts); r != nil {
		return OutcomeFailure, "receiptStatus=" + r.ReceiptStatus
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

	// Compatible default: completed without receipts is still Success.
	// The shipped CLI does not set RequireChainEvidence.
	return OutcomeSuccess, "completed"
}

// ConclusiveFailedReceipt reports a chain-answered failure
// (lib/web3/verify-receipt.ts CONCLUSIVE_STATUSES minus success).
func ConclusiveFailedReceipt(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "reverted", "safe_inner_failure":
		return true
	default:
		return false
	}
}

// NonSuccessReceipt is true when a receipt exists and is not explicitly successful.
func NonSuccessReceipt(r Receipt) bool {
	s := strings.ToLower(strings.TrimSpace(r.ReceiptStatus))
	return s != "" && s != "success"
}

// FirstBlockingReceipt returns the first receipt that must not be treated as success.
func FirstBlockingReceipt(receipts []Receipt) *Receipt {
	for i := range receipts {
		if NonSuccessReceipt(receipts[i]) {
			return &receipts[i]
		}
	}
	return nil
}

func hasVerifiedSuccess(receipts []Receipt) bool {
	for _, r := range receipts {
		if r.Verified && strings.EqualFold(r.ReceiptStatus, "success") {
			return true
		}
	}
	return false
}
