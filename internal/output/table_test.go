package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/keeperhub/cli/internal/output"
)

func TestNewTable_ReturnWriter(t *testing.T) {
	var buf bytes.Buffer
	tw := output.NewTable(&buf, false)
	if tw == nil {
		t.Fatal("expected non-nil table.Writer")
	}
}

func TestNewTable_NonTTY_TSVNoHeaders(t *testing.T) {
	var buf bytes.Buffer
	tw := output.NewTable(&buf, false)
	tw.AppendHeader(table.Row{"Name", "Status"})
	tw.AppendRow(table.Row{"my-workflow", "active"})
	tw.Render()

	got := buf.String()
	// Non-TTY should suppress headers
	if strings.Contains(got, "NAME") || strings.Contains(got, "Name") {
		t.Errorf("non-TTY output should not contain headers, got: %s", got)
	}
	// Should contain tab-separated row data
	if !strings.Contains(got, "my-workflow\tactive") {
		t.Errorf("expected tab-separated row data, got: %s", got)
	}
}

func TestNewTable_TTY_HasHeaders(t *testing.T) {
	var buf bytes.Buffer
	tw := output.NewTable(&buf, true)
	tw.AppendHeader(table.Row{"Name", "Status"})
	tw.AppendRow(table.Row{"my-workflow", "active"})
	tw.Render()

	got := buf.String()
	if !strings.Contains(got, "NAME") && !strings.Contains(got, "Name") {
		t.Errorf("TTY output should contain headers, got: %s", got)
	}
	if !strings.Contains(got, "my-workflow") {
		t.Errorf("expected row data in output, got: %s", got)
	}
}

func TestNewTable_NonTTY_MultipleRows(t *testing.T) {
	var buf bytes.Buffer
	tw := output.NewTable(&buf, false)
	tw.AppendHeader(table.Row{"ID", "Name"})
	tw.AppendRow(table.Row{"1", "alpha"})
	tw.AppendRow(table.Row{"2", "beta"})
	tw.Render()

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}
	if lines[0] != "1\talpha" {
		t.Errorf("expected first line '1\\talpha', got %q", lines[0])
	}
	if lines[1] != "2\tbeta" {
		t.Errorf("expected second line '2\\tbeta', got %q", lines[1])
	}
}

func TestNewTable_NonTTY_CutCompatible(t *testing.T) {
	var buf bytes.Buffer
	tw := output.NewTable(&buf, false)
	tw.AppendHeader(table.Row{"ID", "Name", "Status"})
	tw.AppendRow(table.Row{"abc", "my-wf", "active"})
	tw.Render()

	got := buf.String()
	// Simulates `cut -f2` by splitting on tab
	fields := strings.Split(strings.TrimSpace(got), "\t")
	if len(fields) != 3 {
		t.Fatalf("expected 3 tab-separated fields, got %d: %q", len(fields), got)
	}
	if fields[1] != "my-wf" {
		t.Errorf("cut -f2 should yield 'my-wf', got %q", fields[1])
	}
}
