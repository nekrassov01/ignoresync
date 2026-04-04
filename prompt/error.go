package prompt

import (
	"github.com/nekrassov01/ignoresync"
)

// ErrorKind represents the kind of error that occurred in the operator.
type ErrorKind int

const (
	// ErrorKindNone indicates no error.
	ErrorKindNone ErrorKind = iota

	// ErrorKindPrompt indicates error in interaction.
	ErrorKindPrompt
)

// String returns the string representation of the ErrorKind.
func (k ErrorKind) String() string {
	switch k {
	case ErrorKindNone:
		return "no error"
	case ErrorKindPrompt:
		return "interaction error"
	default:
		return "unknown error"
	}
}

// Error represents an error that occurred in the prompt.
type Error struct {
	Kind ErrorKind
	Err  error
}

// Error returns the string representation of the prompt error.
func (e *Error) Error() string {
	return ignoresync.FormatError("prompt", e.Kind.String(), e.Err)
}

// Unwrap returns the underlying error of the prompt error.
func (e *Error) Unwrap() error {
	return e.Err
}

// NewPromptError constructs a prompt error wrapping optional underlying error.
func NewPromptError(err error) error {
	return &Error{Kind: ErrorKindPrompt, Err: err}
}
