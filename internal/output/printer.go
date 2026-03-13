package output

import (
	"fmt"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

// Printer routes command output to JSON, jq-filtered JSON, or human table
// based on the global --json and --jq flags.
type Printer struct {
	out      io.Writer
	errOut   io.Writer
	jsonMode bool
	jqExpr   string
	isTTY    bool
}

// NewPrinter creates a Printer by reading --json and --jq flags from the cobra command.
// If --jq is set, JSON mode is implied.
func NewPrinter(ios *iostreams.IOStreams, cmd *cobra.Command) *Printer {
	jsonMode, _ := cmd.Flags().GetBool("json")
	jqExpr, _ := cmd.Flags().GetString("jq")
	if jqExpr != "" {
		jsonMode = true
	}
	return &Printer{
		out:      ios.Out,
		errOut:   ios.ErrOut,
		jsonMode: jsonMode,
		jqExpr:   jqExpr,
		isTTY:    ios.IsTerminal(),
	}
}

// IsJSON reports whether the printer is in JSON output mode.
func (p *Printer) IsJSON() bool {
	return p.jsonMode
}

// PrintJSON writes data as indented JSON to stdout.
// If a jq expression is configured, the filter is applied first.
func (p *Printer) PrintJSON(data any) error {
	if p.jqExpr != "" {
		filtered, err := ApplyJQFilter(p.jqExpr, data)
		if err != nil {
			return err
		}
		return WriteJSON(p.out, filtered)
	}
	return WriteJSON(p.out, data)
}

// PrintTable creates a go-pretty table, calls populateFn to add headers/rows, then renders.
func (p *Printer) PrintTable(populateFn func(table.Writer)) {
	tw := NewTable(p.out, p.isTTY)
	populateFn(tw)
}

// PrintData writes data to stdout. In JSON/jq mode it calls PrintJSON; otherwise it calls
// populateFn with a table.Writer. Pass nil for populateFn if only JSON output is needed.
func (p *Printer) PrintData(data any, populateFn func(table.Writer)) error {
	if p.jsonMode {
		return p.PrintJSON(data)
	}
	if populateFn != nil {
		p.PrintTable(populateFn)
	}
	return nil
}

// PrintDryRun writes a dry-run message to stdout.
// In JSON mode it writes {"dry_run": true, "message": "..."}.
// In table mode it writes "[dry-run] message\n".
func (p *Printer) PrintDryRun(message string) {
	if p.jsonMode {
		_ = WriteJSON(p.out, map[string]any{"dry_run": true, "message": message})
		return
	}
	fmt.Fprintf(p.out, "[dry-run] %s\n", message)
}

// PrintKeyValue prints key-value pairs in a left-aligned format.
// In TTY mode: keys are padded to the longest key width with a colon and tab.
// In non-TTY mode: same tab-separated format.
// Example:
//
//	ID:       abc123
//	Name:     My Workflow
//	Status:   active
func (p *Printer) PrintKeyValue(pairs [][2]string) {
	if len(pairs) == 0 {
		return
	}
	maxKeyLen := 0
	for _, pair := range pairs {
		if len(pair[0]) > maxKeyLen {
			maxKeyLen = len(pair[0])
		}
	}
	for _, pair := range pairs {
		// pad key with trailing spaces so values align
		padded := pair[0] + ":"
		for len(padded) < maxKeyLen+2 {
			padded += " "
		}
		fmt.Fprintf(p.out, "%s\t%s\n", padded, pair[1])
	}
}

// PrintError writes an error message to stderr.
// In JSON mode it writes {"error": "...", "code": N}.
// In table mode it writes plain text.
func (p *Printer) PrintError(message string, code int) error {
	if p.jsonMode {
		return WriteJSONError(p.errOut, message, code)
	}
	fmt.Fprintln(p.errOut, message)
	return nil
}
