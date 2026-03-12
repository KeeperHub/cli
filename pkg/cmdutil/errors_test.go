package cmdutil_test

import (
	"errors"
	"testing"
	"time"

	"github.com/keeperhub/cli/pkg/cmdutil"
)

func TestNotFoundError(t *testing.T) {
	inner := errors.New("workflow not found")
	err := cmdutil.NotFoundError{Err: inner}

	if err.Error() != "workflow not found" {
		t.Errorf("expected %q, got %q", "workflow not found", err.Error())
	}
	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to find inner error via Unwrap")
	}
}

func TestRateLimitError_WithRetryAfter(t *testing.T) {
	inner := errors.New("rate limit exceeded")
	err := cmdutil.RateLimitError{Err: inner, RetryAfter: 30 * time.Second}

	expected := "rate limit exceeded (retry after 30s)"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to find inner error via Unwrap")
	}
}

func TestRateLimitError_WithoutRetryAfter(t *testing.T) {
	inner := errors.New("rate limit exceeded")
	err := cmdutil.RateLimitError{Err: inner}

	if err.Error() != "rate limit exceeded" {
		t.Errorf("expected %q, got %q", "rate limit exceeded", err.Error())
	}
}

func TestExitCodeForError_Nil(t *testing.T) {
	if code := cmdutil.ExitCodeForError(nil); code != 0 {
		t.Errorf("expected 0 for nil, got %d", code)
	}
}

func TestExitCodeForError_CancelError(t *testing.T) {
	err := cmdutil.CancelError{Err: errors.New("cancelled")}
	if code := cmdutil.ExitCodeForError(err); code != 0 {
		t.Errorf("expected 0 for CancelError, got %d", code)
	}
}

func TestExitCodeForError_Generic(t *testing.T) {
	err := errors.New("something went wrong")
	if code := cmdutil.ExitCodeForError(err); code != 1 {
		t.Errorf("expected 1 for generic error, got %d", code)
	}
}

func TestExitCodeForError_NotFoundError(t *testing.T) {
	err := cmdutil.NotFoundError{Err: errors.New("not found")}
	if code := cmdutil.ExitCodeForError(err); code != 2 {
		t.Errorf("expected 2 for NotFoundError, got %d", code)
	}
}

func TestExitCodeForError_FlagError(t *testing.T) {
	err := cmdutil.FlagError{Err: errors.New("invalid flag")}
	if code := cmdutil.ExitCodeForError(err); code != 2 {
		t.Errorf("expected 2 for FlagError, got %d", code)
	}
}

func TestExitCodeForError_RateLimitError(t *testing.T) {
	err := cmdutil.RateLimitError{Err: errors.New("rate limit"), RetryAfter: 60 * time.Second}
	if code := cmdutil.ExitCodeForError(err); code != 5 {
		t.Errorf("expected 5 for RateLimitError, got %d", code)
	}
}
