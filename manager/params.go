package manager

// encryptionContextKey is the key for encryption context.
const encryptionContextKey = "bucket/key"

// resourceNameSize is the size of the resource name in bytes.
const resourceNameSize = 16

// infoBucket is the info string for generating the bucket name.
var infoBucket = []byte("ignoresyncbucket")

// infoSSEKey is the info string for generating the SSE key.
var infoSSEKey = []byte("ignoresyncssekmskey")

// infoCSEKey is the info string for generating the CSE key.
var infoCSEKey = []byte("ignoresynccsekmskey")
