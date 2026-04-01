package ignoresync

import (
	"strings"
)

// baseError represents an basic error.
type baseError struct {
	pkg  string
	kind string
	err  error
}

// Error returns the string representation of the error.
func (e *baseError) Error() string {
	var b strings.Builder
	l := len(e.pkg) + 2 + len(e.kind)
	var s string
	if e.err != nil {
		s = e.err.Error()
		l += 2 + len(s)
	}
	b.Grow(l)
	b.WriteString(e.pkg)
	b.WriteString(": ")
	b.WriteString(e.kind)
	if s != "" {
		b.WriteString(": ")
		b.WriteString(s)
	}
	return b.String()
}

// Unwrap returns the underlying error of the error.
func (e *baseError) Unwrap() error {
	return e.err
}

// NewError constructs an error message with the given package, kind and error.
// This is a common function for creating various errors in a similar way in child
// packages, and is not intended for direct use.
func NewError(pkg, kind string, err error) error {
	return &baseError{
		pkg:  pkg,
		kind: kind,
		err:  err,
	}
}
