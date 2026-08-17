package execrecovery_test

import (
	"testing"

	"github.com/keeperhub/cli/internal/execrecovery"
)

func TestClassify_ReceiptStates(t *testing.T) {
	completed := func(receiptStatus string, verified bool) []byte {
		return []byte(`{"executionId":"x","status":"completed","transactionHash":"0xabc","receipts":[{"hash":"0xabc","verified":` + boolJSON(verified) + `,"receiptStatus":"` + receiptStatus + `","verifiedAt":"2026-08-11T00:00:00Z"}]}`)
	}
	unconfirmed := func(receiptStatus string) []byte {
		return []byte(`{"executionId":"x","status":"unconfirmed","transactionHash":"0xabc","receipts":[{"hash":"0xabc","verified":false,"receiptStatus":"` + receiptStatus + `","verifiedAt":"2026-08-11T00:00:00Z"}]}`)
	}

	t.Run("completed success", func(t *testing.T) {
		got, _ := execrecovery.Classify(execrecovery.Sample{HTTPStatus: 200, Body: completed("success", true)}, execrecovery.Options{})
		if got != execrecovery.OutcomeSuccess {
			t.Fatalf("got %s, want success", got)
		}
	})
	for _, st := range []string{"reverted", "safe_inner_failure", "not_found", "timeout"} {
		st := st
		t.Run("completed "+st+" is failure", func(t *testing.T) {
			got, reason := execrecovery.Classify(execrecovery.Sample{HTTPStatus: 200, Body: completed(st, false)}, execrecovery.Options{})
			if got != execrecovery.OutcomeFailure {
				t.Fatalf("got %s (%s), want failure", got, reason)
			}
		})
	}
	for _, st := range []string{"not_found", "timeout"} {
		st := st
		t.Run("unconfirmed "+st+" is pending", func(t *testing.T) {
			got, reason := execrecovery.Classify(execrecovery.Sample{HTTPStatus: 200, Body: unconfirmed(st)}, execrecovery.Options{})
			if got != execrecovery.OutcomePending {
				t.Fatalf("got %s (%s), want pending", got, reason)
			}
		})
	}
}

func boolJSON(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
