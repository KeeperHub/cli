package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/keeperhub/cli/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
)

func TestNewTable_ReturnWriter(t *testing.T) {
	var buf bytes.Buffer
	tw := output.NewTable(&buf, false)
	if tw == nil {
		t.Fatal("expected non-nil table.Writer")
	}
}

func TestNewTable_RendersRowsWithHeaders(t *testing.T) {
	var buf bytes.Buffer
	tw := output.NewTable(&buf, false)
	tw.AppendHeader(table.Row{"Name", "Status"})
	tw.AppendRow(table.Row{"my-workflow", "active"})
	tw.Render()

	got := buf.String()
	if !strings.Contains(got, "NAME") && !strings.Contains(got, "Name") {
		t.Errorf("expected header Name/NAME in output, got: %s", got)
	}
	if !strings.Contains(got, "my-workflow") {
		t.Errorf("expected row data in output, got: %s", got)
	}
}
