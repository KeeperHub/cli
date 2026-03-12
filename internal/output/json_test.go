package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/keeperhub/cli/internal/output"
)

func TestWriteJSON_Struct(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"name": "foo", "id": 1}
	if err := output.WriteJSON(&buf, data); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, `"name"`) {
		t.Errorf("expected JSON to contain name field, got: %s", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Error("expected trailing newline")
	}
}

func TestWriteJSON_Nil(t *testing.T) {
	var buf bytes.Buffer
	if err := output.WriteJSON(&buf, nil); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "null" {
		t.Errorf("expected null, got: %s", got)
	}
}

func TestWriteJSONError(t *testing.T) {
	var buf bytes.Buffer
	if err := output.WriteJSONError(&buf, "something failed", 42); err != nil {
		t.Fatalf("WriteJSONError error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, `"error"`) {
		t.Errorf("expected error field, got: %s", got)
	}
	if !strings.Contains(got, "something failed") {
		t.Errorf("expected message in output, got: %s", got)
	}
	if !strings.Contains(got, "42") {
		t.Errorf("expected code 42 in output, got: %s", got)
	}
}
