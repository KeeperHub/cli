package output_test

import (
	"strings"
	"testing"

	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func makeCmd(jsonFlag bool, jqExpr string) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().String("jq", "", "")
	if jsonFlag {
		_ = cmd.Flags().Set("json", "true")
	}
	if jqExpr != "" {
		_ = cmd.Flags().Set("jq", jqExpr)
	}
	return cmd
}

func TestPrinter_JSONMode(t *testing.T) {
	ios, out, _, _ := iostreams.Test()
	cmd := makeCmd(true, "")
	p := output.NewPrinter(ios, cmd)

	data := map[string]any{"name": "foo"}
	if err := p.PrintData(data, nil); err != nil {
		t.Fatalf("PrintData error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, `"name"`) {
		t.Errorf("expected JSON output, got: %s", got)
	}
}

func TestPrinter_JQMode(t *testing.T) {
	ios, out, _, _ := iostreams.Test()
	cmd := makeCmd(false, ".name")
	p := output.NewPrinter(ios, cmd)

	data := map[string]any{"name": "bar", "id": 2}
	if err := p.PrintData(data, nil); err != nil {
		t.Fatalf("PrintData error: %v", err)
	}

	got := strings.TrimSpace(out.String())
	if got != `"bar"` {
		t.Errorf("expected filtered JSON \"bar\", got: %s", got)
	}
}

func TestPrinter_JQImpliesJSON_NoTable(t *testing.T) {
	ios, out, _, _ := iostreams.Test()
	cmd := makeCmd(false, ".name")
	p := output.NewPrinter(ios, cmd)

	if !p.IsJSON() {
		t.Error("--jq should imply IsJSON() == true")
	}

	data := map[string]any{"name": "baz"}
	tableCalled := false
	err := p.PrintData(data, func(_ table.Writer) {
		tableCalled = true
	})
	if err != nil {
		t.Fatalf("PrintData error: %v", err)
	}
	if tableCalled {
		t.Error("table function should not be called in jq/json mode")
	}
	_ = out
}

func TestPrinter_TableMode(t *testing.T) {
	ios, out, _, _ := iostreams.Test()
	cmd := makeCmd(false, "")
	p := output.NewPrinter(ios, cmd)

	data := map[string]any{"name": "wf1"}
	tableCalled := false
	err := p.PrintData(data, func(tw table.Writer) {
		tableCalled = true
		tw.AppendHeader(table.Row{"Name"})
		tw.AppendRow(table.Row{"wf1"})
		tw.Render()
	})
	if err != nil {
		t.Fatalf("PrintData error: %v", err)
	}
	if !tableCalled {
		t.Error("table function should be called in table mode")
	}
	got := out.String()
	if !strings.Contains(got, "wf1") {
		t.Errorf("expected table output, got: %s", got)
	}
}

func TestPrinter_PrintDryRun(t *testing.T) {
	ios, out, _, _ := iostreams.Test()
	cmd := makeCmd(false, "")
	p := output.NewPrinter(ios, cmd)

	p.PrintDryRun("would create workflow foo")

	got := out.String()
	if !strings.Contains(got, "[dry-run]") {
		t.Errorf("expected [dry-run] prefix, got: %s", got)
	}
	if !strings.Contains(got, "would create workflow foo") {
		t.Errorf("expected message in dry-run output, got: %s", got)
	}
}

func TestPrinter_PrintError_WritesToErrOut(t *testing.T) {
	ios, out, errOut, _ := iostreams.Test()
	cmd := makeCmd(false, "")
	p := output.NewPrinter(ios, cmd)

	if err := p.PrintError("something went wrong", 1); err != nil {
		t.Fatalf("PrintError error: %v", err)
	}

	if out.String() != "" {
		t.Errorf("expected nothing on stdout, got: %s", out.String())
	}
	if !strings.Contains(errOut.String(), "something went wrong") {
		t.Errorf("expected error on stderr, got: %s", errOut.String())
	}
}

func TestNewPrinter_ReadsFlags(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cmd := makeCmd(true, "")
	p := output.NewPrinter(ios, cmd)

	if !p.IsJSON() {
		t.Error("expected IsJSON() == true when --json flag set")
	}
}
