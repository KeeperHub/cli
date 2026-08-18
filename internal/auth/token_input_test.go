package auth

import (
	"bytes"
	"strings"
	"testing"

	"github.com/keeperhub/cli/pkg/iostreams"
)

// Piped input must keep working exactly as before. CI pipes a key in, and changing that would
// break every existing automation.
func TestReadTokenFromStdin_PipedInputUnchanged(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"plain token", "kh_example_token", "kh_example_token"},
		{"trailing newline from echo", "kh_example_token\n", "kh_example_token"},
		{"surrounding whitespace", "  kh_example_token \n", "kh_example_token"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			ios.In = io_NopReader(tc.input)

			got, err := ReadTokenFromStdin(ios)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReadTokenFromStdin_EmptyPipedInputIsAnError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.In = io_NopReader("   \n")

	if _, err := ReadTokenFromStdin(ios); err == nil {
		t.Fatal("expected an error for empty stdin")
	}
}

// A buffer-backed stream is not a terminal, so the reader must take the piped path and never
// attempt to prompt. This is the guard that keeps non-interactive use working.
func TestReadTokenFromStdin_NonTerminalTakesPipedPath(t *testing.T) {
	ios, _, errOut, _ := iostreams.Test()
	ios.In = io_NopReader("kh_piped")

	got, err := ReadTokenFromStdin(ios)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "kh_piped" {
		t.Errorf("got %q, want %q", got, "kh_piped")
	}
	if prompt := errOut.String(); strings.Contains(prompt, "API key") {
		t.Errorf("must not prompt when stdin is not a terminal, got: %q", prompt)
	}
}

// The token must never be written to any stream the caller controls.
func TestReadTokenFromStdin_TokenIsNotEchoed(t *testing.T) {
	ios, out, errOut, _ := iostreams.Test()
	const secret = "kh_must_not_appear_anywhere"
	ios.In = io_NopReader(secret)

	if _, err := ReadTokenFromStdin(ios); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for name, buf := range map[string]*bytes.Buffer{"stdout": out, "stderr": errOut} {
		if strings.Contains(buf.String(), secret) {
			t.Errorf("token appeared on %s: %q", name, buf.String())
		}
		if strings.Contains(buf.String(), "must_not_appear") {
			t.Errorf("part of the token appeared on %s: %q", name, buf.String())
		}
	}
}

func io_NopReader(s string) *bytes.Buffer { return bytes.NewBufferString(s) }
