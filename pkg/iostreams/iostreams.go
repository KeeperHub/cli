package iostreams

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/term"
)

// IOStreams bundles the three standard streams for a CLI process.
type IOStreams struct {
	Out    io.Writer
	ErrOut io.Writer
	In     io.Reader
}

// System returns IOStreams wired to the real OS file descriptors.
func System() *IOStreams {
	return &IOStreams{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		In:     os.Stdin,
	}
}

// Test returns IOStreams backed by in-memory buffers, suitable for unit tests.
// The three returned buffers capture writes to Out, ErrOut, and provide input
// via In respectively.
func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	in := &bytes.Buffer{}
	return &IOStreams{
		Out:    out,
		ErrOut: errOut,
		In:     in,
	}, out, errOut, in
}

// IsTerminal reports whether Out is connected to a real terminal.
func (s *IOStreams) IsTerminal() bool {
	f, ok := s.Out.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
