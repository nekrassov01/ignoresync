package operator

import (
	"bytes"
)

// prefixBuilder defines a builder for constructing S3 object keys.
type prefixBuilder struct {
	buf      *bytes.Buffer
	repoHash string
}

// newPrefixBuilder creates a new prefixBuilder with the given repository hash.
func newPrefixBuilder(repoHash string) *prefixBuilder {
	return &prefixBuilder{
		buf:      &bytes.Buffer{},
		repoHash: repoHash,
	}
}

// build constructs the S3 object key using the given prefix and suffix.
// The internal buffer is first reset and reused to construct new keys.
func (o *prefixBuilder) build(name string) string {
	o.buf.Reset()
	o.buf.Grow(len(o.repoHash) + 1 + len(name))
	o.buf.WriteString(o.repoHash)
	o.buf.WriteString("/")
	o.buf.WriteString(name)
	return o.buf.String()
}
