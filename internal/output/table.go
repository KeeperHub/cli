package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// tsvWriter implements table.Writer but writes tab-separated values
// with no headers and no borders, matching gh CLI non-TTY output.
type tsvWriter struct {
	w    io.Writer
	rows [][]string
}

func (t *tsvWriter) AppendHeader(_ table.Row, _ ...table.RowConfig) {}
func (t *tsvWriter) AppendRow(row table.Row, _ ...table.RowConfig) {
	cols := make([]string, len(row))
	for i, v := range row {
		cols[i] = fmt.Sprintf("%v", v)
	}
	t.rows = append(t.rows, cols)
}
func (t *tsvWriter) AppendRows(rows []table.Row, _ ...table.RowConfig) {
	for _, row := range rows {
		t.AppendRow(row)
	}
}
func (t *tsvWriter) Render() string {
	var sb strings.Builder
	for _, cols := range t.rows {
		line := strings.Join(cols, "\t")
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	out := sb.String()
	fmt.Fprint(t.w, out)
	return out
}

func (t *tsvWriter) AppendFooter(_ table.Row, _ ...table.RowConfig) {}
func (t *tsvWriter) AppendSeparator()                               {}
func (t *tsvWriter) FilterBy(_ []table.FilterBy)                    {}
func (t *tsvWriter) ImportGrid(_ interface{}) bool                  { return false }
func (t *tsvWriter) Length() int                                    { return len(t.rows) }
func (t *tsvWriter) Pager(_ ...table.PagerOption) table.Pager      { return nil }
func (t *tsvWriter) RenderCSV() string                              { return t.Render() }
func (t *tsvWriter) RenderHTML() string                             { return t.Render() }
func (t *tsvWriter) RenderMarkdown() string                         { return t.Render() }
func (t *tsvWriter) RenderTSV() string                              { return t.Render() }
func (t *tsvWriter) ResetFooters()                                  {}
func (t *tsvWriter) ResetHeaders()                                  {}
func (t *tsvWriter) ResetRows()                                     { t.rows = nil }
func (t *tsvWriter) SetAutoIndex(_ bool)                            {}
func (t *tsvWriter) SetCaption(_ string, _ ...interface{})          {}
func (t *tsvWriter) SetColumnConfigs(_ []table.ColumnConfig)        {}
func (t *tsvWriter) SetIndexColumn(_ int)                           {}
func (t *tsvWriter) SetOutputMirror(_ io.Writer)                    {}
func (t *tsvWriter) SetRowPainter(_ interface{})                    {}
func (t *tsvWriter) SetStyle(_ table.Style)                         {}
func (t *tsvWriter) SetTitle(_ string, _ ...interface{})            {}
func (t *tsvWriter) SortBy(_ []table.SortBy)                        {}
func (t *tsvWriter) Style() *table.Style                            { return nil }
func (t *tsvWriter) SuppressEmptyColumns()                          {}
func (t *tsvWriter) SuppressTrailingSpaces()                        {}
func (t *tsvWriter) SetAllowedRowLength(_ int)                      {}
func (t *tsvWriter) SetHTMLCSSClass(_ string)                       {}
func (t *tsvWriter) SetPageSize(_ int)                              {}

// NewTable returns a table.Writer configured for the current output context.
// When isTTY is true, a go-pretty table with StyleLight (box-drawing) is returned.
// When false, a lightweight TSV writer is returned that outputs tab-separated
// values with no headers or borders, making `kh wf ls | cut -f2` work like gh.
func NewTable(w io.Writer, isTTY bool) table.Writer {
	if !isTTY {
		return &tsvWriter{w: w}
	}
	tw := table.NewWriter()
	tw.SetOutputMirror(w)
	tw.SetStyle(table.Style{
		Name: "clean",
		Box:  table.StyleBoxDefault,
		Format: table.FormatOptions{
			Header: 0,
		},
		Options: table.Options{
			DrawBorder:      false,
			SeparateColumns: true,
			SeparateFooter:  false,
			SeparateHeader:  false,
			SeparateRows:    false,
		},
	})
	return tw
}
