package errors

import (
	"strings"
)

// FormatError formats an error string with the given package, kind and error.
func FormatError(pkg, kind string, err error) string {
	var b strings.Builder
	l := len(pkg) + 2 + len(kind)
	var s string
	if err != nil {
		s = err.Error()
		l += 2 + len(s)
	}
	b.Grow(l)
	b.WriteString(pkg)
	b.WriteString(": ")
	b.WriteString(kind)
	if s != "" {
		b.WriteString(": ")
		b.WriteString(s)
	}
	return b.String()
}
