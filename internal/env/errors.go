package env

import "github.com/nekrassov01/ignoresync/internal/errors"

// ErrorKind represents the kind of error.
type ErrorKind int

const (
	// ErrorKindNone indicates no error.
	ErrorKindNone ErrorKind = iota

	// ErrorKindEnv indicates error in environment.
	ErrorKindEnv
)

// String returns the string representation of the ErrorKind.
func (k ErrorKind) String() string {
	switch k {
	case ErrorKindNone:
		return "no error"
	case ErrorKindEnv:
		return "environment error"
	default:
		return "unknown error"
	}
}

// Error represents an error with type and message.
type Error struct {
	Kind ErrorKind
	Err  error
}

// Error returns the string representation of the error.
func (e *Error) Error() string {
	return errors.FormatError("env", e.Kind.String(), e.Err)
}

// Unwrap returns the underlying error of the error.
func (e *Error) Unwrap() error {
	return e.Err
}

// NewEnvError constructs an environment error wrapping optional underlying error.
func NewEnvError(err error) error {
	return &Error{Kind: ErrorKindEnv, Err: err}
}
