package output_test

import (
	"testing"

	"github.com/keeperhub/cli/internal/output"
)

func TestApplyJQFilter_SingleField(t *testing.T) {
	data := map[string]any{"name": "foo", "id": 1}
	result, err := output.ApplyJQFilter(".name", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "foo" {
		t.Errorf("expected foo, got %v", result)
	}
}

func TestApplyJQFilter_ArrayIteration(t *testing.T) {
	data := map[string]any{
		"items": []any{
			map[string]any{"id": 1},
			map[string]any{"id": 2},
		},
	}
	result, err := output.ApplyJQFilter(".items[]|.id", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slice, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(slice) != 2 {
		t.Errorf("expected 2 items, got %d", len(slice))
	}
}

func TestApplyJQFilter_Identity(t *testing.T) {
	data := map[string]any{"x": 1}
	result, err := output.ApplyJQFilter(".", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["x"] != 1 {
		t.Errorf("expected x=1, got %v", m["x"])
	}
}

func TestApplyJQFilter_InvalidExpression(t *testing.T) {
	data := map[string]any{"x": 1}
	_, err := output.ApplyJQFilter("|||invalid", data)
	if err == nil {
		t.Error("expected error for invalid jq expression")
	}
}
