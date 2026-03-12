package iostreams_test

import (
	"testing"

	"github.com/keeperhub/cli/pkg/iostreams"
)

func TestSystem(t *testing.T) {
	s := iostreams.System()
	if s == nil {
		t.Fatal("System() returned nil")
	}
	if s.Out == nil {
		t.Error("System().Out is nil")
	}
	if s.ErrOut == nil {
		t.Error("System().ErrOut is nil")
	}
	if s.In == nil {
		t.Error("System().In is nil")
	}
}

func TestTest(t *testing.T) {
	s, outBuf, errOutBuf, _ := iostreams.Test()
	if s == nil {
		t.Fatal("Test() returned nil IOStreams")
	}

	_, err := s.Out.Write([]byte("hello out"))
	if err != nil {
		t.Fatalf("writing to Out: %v", err)
	}
	if got := outBuf.String(); got != "hello out" {
		t.Errorf("Out capture: got %q, want %q", got, "hello out")
	}

	_, err = s.ErrOut.Write([]byte("hello err"))
	if err != nil {
		t.Fatalf("writing to ErrOut: %v", err)
	}
	if got := errOutBuf.String(); got != "hello err" {
		t.Errorf("ErrOut capture: got %q, want %q", got, "hello err")
	}
}

func TestIsTerminal(t *testing.T) {
	s, _, _, _ := iostreams.Test()
	if s.IsTerminal() {
		t.Error("IsTerminal() should return false for buffer-backed streams")
	}
}
