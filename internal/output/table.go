package output

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
)

// NewTable returns a go-pretty table.Writer configured to write to w.
// When isTTY is true, table.StyleLight is used (box-drawing characters).
// When false, table.StyleDefault is used (ASCII, suitable for piped output).
func NewTable(w io.Writer, isTTY bool) table.Writer {
	tw := table.NewWriter()
	tw.SetOutputMirror(w)
	if isTTY {
		tw.SetStyle(table.StyleLight)
	} else {
		tw.SetStyle(table.StyleDefault)
	}
	return tw
}
