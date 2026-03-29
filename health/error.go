package health

import (
	"github.com/nekrassov01/ignoresync"
)

// ErrorKind represents the kind of error.
type ErrorKind int

const (
	// ErrorKindNone indicates no error.
	ErrorKindNone ErrorKind = iota

	// ErrorKindState indicates error in state validation.
	ErrorKindState

	// ErrorKindStack indicates error in stack validation.
	ErrorKindStack

	// ErrorKindBucket indicates error in bucket validation.
	ErrorKindBucket

	// ErrorKindKey indicates error in key validation.
	ErrorKindKey
)

// String returns the string representation of the ErrorKind.
func (k ErrorKind) String() string {
	switch k {
	case ErrorKindNone:
		return "no error"
	case ErrorKindState:
		return "state error"
	case ErrorKindStack:
		return "stack error"
	case ErrorKindBucket:
		return "bucket error"
	case ErrorKindKey:
		return "key error"
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
	return ignoresync.NewError("healthchecker", e.Kind.String(), e.Err).Error()
}

// Unwrap returns the underlying error of the error.
func (e *Error) Unwrap() error {
	return e.Err
}

// NewStateError constructs a state validation error wrapping optional underlying error.
func NewStateError(err error) error {
	return &Error{Kind: ErrorKindState, Err: err}
}

// NewStackError constructs a stack validation error wrapping optional underlying error.
func NewStackError(err error) error {
	return &Error{Kind: ErrorKindStack, Err: err}
}

// NewBucketError constructs a bucket validation error wrapping optional underlying error.
func NewBucketError(err error) error {
	return &Error{Kind: ErrorKindBucket, Err: err}
}

// NewKeyError constructs a key validation error wrapping optional underlying error.
func NewKeyError(err error) error {
	return &Error{Kind: ErrorKindKey, Err: err}
}
