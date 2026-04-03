package operator

import (
	"io"

	"github.com/dustin/go-humanize"
	"github.com/nekrassov01/ignoresync/color"
)

// printStatus prints the status of the file operation in a formatted manner.
func (o *Operator) printStatus(msg, path string, size int64, prefix string) {
	if prefix != "" {
		_, _ = io.WriteString(o.w, color.Mute(prefix))
		_, _ = io.WriteString(o.w, " ")
	}
	_, _ = io.WriteString(o.w, msg)
	_, _ = io.WriteString(o.w, color.Mute(" -> "))
	_, _ = io.WriteString(o.w, path)
	if size >= 0 {
		_, _ = io.WriteString(o.w, " (")
		_, _ = io.WriteString(o.w, sizeString(size))
		_, _ = io.WriteString(o.w, ")")
	}
	_, _ = io.WriteString(o.w, "\n")
}

// sizeString formats the file size in a human-readable format, handling negative sizes as zero.
func sizeString(size int64) string {
	if size < 0 {
		return humanize.Bytes(0)
	}
	return humanize.Bytes(uint64(size))
}
