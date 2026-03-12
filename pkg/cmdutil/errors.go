package cmdutil

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
