package cmdutil

import (
	"errors"
	"fmt"
	"time"
)

// NotFoundError wraps a 404 response. Exit code: 2.
type NotFoundError struct {
	Err error
}

func (e NotFoundError) Error() string { return e.Err.Error() }
func (e NotFoundError) Unwrap() error { return e.Err }

// RateLimitError wraps a 429 response. Exit code: 5.
type RateLimitError struct {
	Err        error
	RetryAfter time.Duration
}

func (e RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s (retry after %s)", e.Err.Error(), e.RetryAfter)
	}
	return e.Err.Error()
}

func (e RateLimitError) Unwrap() error { return e.Err }

// ExitCodeForError maps error types to process exit codes:
//   - nil or CancelError -> 0
//   - FlagError or NotFoundError -> 2
//   - RateLimitError -> 5
//   - everything else -> 1
func ExitCodeForError(err error) int {
	if err == nil {
		return 0
	}
	var cancelErr CancelError
	if errors.As(err, &cancelErr) {
		return 0
	}
	var flagErr FlagError
	if errors.As(err, &flagErr) {
		return 2
	}
	var notFoundErr NotFoundError
	if errors.As(err, &notFoundErr) {
		return 2
	}
	var rateLimitErr RateLimitError
	if errors.As(err, &rateLimitErr) {
		return 5
	}
	return 1
}

// FlagError indicates an invalid or missing CLI flag value.
// The error message is printed without a stack trace and the process exits 2.
type FlagError struct {
	Err error
}

func (e FlagError) Error() string {
	return e.Err.Error()
}

func (e FlagError) Unwrap() error {
	return e.Err
}

// SilentError signals that the error message has already been printed and the
// process should exit non-zero without printing anything further.
type SilentError struct {
	Err error
}

func (e SilentError) Error() string {
	return e.Err.Error()
}

func (e SilentError) Unwrap() error {
	return e.Err
}

// CancelError indicates a user-initiated cancellation (e.g., answering "no" to
// a confirmation prompt). The process exits 0 with no error output.
type CancelError struct {
	Err error
}

func (e CancelError) Error() string {
	return e.Err.Error()
}

func (e CancelError) Unwrap() error {
	return e.Err
}
