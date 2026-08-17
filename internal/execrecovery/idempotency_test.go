package execrecovery_test

import (
	"strings"
	"testing"

	"github.com/keeperhub/cli/internal/execrecovery"
)

func TestResolveIdempotencyKey_ExplicitStable(t *testing.T) {
	a, err := execrecovery.ResolveIdempotencyKey("stable-key-1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := execrecovery.ResolveIdempotencyKey("stable-key-1")
	if err != nil {
		t.Fatal(err)
	}
	if a != b || a != "stable-key-1" {
		t.Fatalf("explicit key must be preserved: %q vs %q", a, b)
	}
}

func TestResolveIdempotencyKey_GeneratedUnique(t *testing.T) {
	a, err := execrecovery.ResolveIdempotencyKey("")
	if err != nil {
		t.Fatal(err)
	}
	b, err := execrecovery.ResolveIdempotencyKey("")
	if err != nil {
		t.Fatal(err)
	}
	if a == "" || b == "" {
		t.Fatal("generated keys must be non-empty")
	}
	if a == b {
		t.Fatal("different intents must get different generated keys")
	}
}

func TestParseIdempotencyBody_InProgress(t *testing.T) {
	b, ok := execrecovery.ParseIdempotencyBody([]byte(`{"error":"A request with this Idempotency-Key is already being processed. Retry the same key shortly; do not rotate it.","code":"idempotency_in_progress","retryable":true}`))
	if !ok {
		t.Fatal("expected parse ok")
	}
	if !b.IsInProgress() || b.IsConflict() {
		t.Fatalf("in_progress misclassified: %+v", b)
	}
}

func TestParseIdempotencyBody_Conflict(t *testing.T) {
	orig := "exec-original"
	b, ok := execrecovery.ParseIdempotencyBody([]byte(`{"error":"Idempotency-Key was reused with a different request payload. Use a new key for a different request.","code":"idempotency_conflict","originalExecutionId":"exec-original","retryable":false}`))
	if !ok {
		t.Fatal("expected parse ok")
	}
	if !b.IsConflict() || b.IsInProgress() {
		t.Fatalf("conflict misclassified: %+v", b)
	}
	if b.OriginalExecutionID == nil || *b.OriginalExecutionID != orig {
		t.Fatalf("originalExecutionId=%v", b.OriginalExecutionID)
	}
	err := execrecovery.ConflictError{Body: b, Key: "k1"}
	if !strings.Contains(err.Error(), "exec-original") || !strings.Contains(err.Error(), "do not retry with a new key") {
		t.Fatalf("conflict error=%s", err.Error())
	}
}

func TestParseIdempotencyBody_Unknown409(t *testing.T) {
	_, ok := execrecovery.ParseIdempotencyBody([]byte(`{"error":"something else"}`))
	if ok {
		t.Fatal("generic 409 must not look like an idempotency code")
	}
}
