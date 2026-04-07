package manager

import "github.com/nekrassov01/ignoresync/internal/errors"

// ErrorKind represents the kind of error.
type ErrorKind int

const (
	// ErrorKindNone indicates no error.
	ErrorKindNone ErrorKind = iota

	// ErrorKindAuth indicates error in authentication.
	ErrorKindAuth

	// ErrorKindCredential indicates error in credential management.
	ErrorKindCredential

	// ErrorKindState indicates error in state management.
	ErrorKindState
)

// String returns the string representation of the ErrorKind.
func (k ErrorKind) String() string {
	switch k {
	case ErrorKindNone:
		return "no error"
	case ErrorKindAuth:
		return "authentication error"
	case ErrorKindCredential:
		return "credential error"
	case ErrorKindState:
		return "state error"
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
	return errors.FormatError("manager", e.Kind.String(), e.Err)
}

// Unwrap returns the underlying error of the error.
func (e *Error) Unwrap() error {
	return e.Err
}

// NewAuthError constructs a auth error wrapping optional underlying error.
func NewAuthError(err error) error {
	return &Error{Kind: ErrorKindAuth, Err: err}
}

// NewCredentialError constructs a credential error wrapping optional underlying error.
func NewCredentialError(err error) error {
	return &Error{Kind: ErrorKindCredential, Err: err}
}

// NewStateError constructs a state error wrapping optional underlying error.
func NewStateError(err error) error {
	return &Error{Kind: ErrorKindState, Err: err}
}
