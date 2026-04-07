package operator

import (
	"strings"
)

// prefixBuilder defines a builder for constructing S3 object keys.
type prefixBuilder struct {
	builder  strings.Builder
	repoHash string
}

// newPrefixBuilder creates a new prefixBuilder with the given repository hash.
func newPrefixBuilder(repoHash string) *prefixBuilder {
	return &prefixBuilder{
		repoHash: repoHash,
	}
}

// build constructs the S3 object key using the given prefix and suffix.
// The internal buffer is first reset and reused to construct new keys.
func (o *prefixBuilder) build(name string) string {
	o.builder.Reset()
	o.builder.Grow(len(o.repoHash) + 1 + len(name))
	o.builder.WriteString(o.repoHash)
	o.builder.WriteString("/")
	o.builder.WriteString(name)
	return o.builder.String()
}
