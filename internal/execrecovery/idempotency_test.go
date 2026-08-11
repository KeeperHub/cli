package execrecovery_test

import (
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
