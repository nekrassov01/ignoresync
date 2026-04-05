package config

import "github.com/nekrassov01/ignoresync/errors"

// ErrorKind represents the kind of error.
type ErrorKind int

const (
	// ErrorKindNone indicates no error.
	ErrorKindNone ErrorKind = iota

	// ErrorKindConfig indicates error in AWS configuration.
	ErrorKindConfig
)

// String returns the string representation of the ErrorKind.
func (k ErrorKind) String() string {
	switch k {
	case ErrorKindNone:
		return "no error"
	case ErrorKindConfig:
		return "config error"
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
	return errors.FormatError("config", e.Kind.String(), e.Err)
}

// Unwrap returns the underlying error of the error.
func (e *Error) Unwrap() error {
	return e.Err
}

// NewConfigError constructs a config error wrapping optional underlying error.
func NewConfigError(err error) error {
	return &Error{Kind: ErrorKindConfig, Err: err}
}
