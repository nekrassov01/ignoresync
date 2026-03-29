package manager

import (
	"github.com/nekrassov01/ignoresync"
)

// encryptionContextKey is the key for encryption context.
const encryptionContextKey = "bucket/key"

// resourceNameSize is the size of the resource name in bytes.
const resourceNameSize = 16

// infoBucket is the info string for generating the bucket name.
var infoBucket = []byte(ignoresync.CommandName + "bucket")

// infoSSEKey is the info string for generating the SSE key.
var infoSSEKey = []byte(ignoresync.CommandName + "ssekmskey")

// infoCSEKey is the info string for generating the CSE key.
var infoCSEKey = []byte(ignoresync.CommandName + "csekmskey")
