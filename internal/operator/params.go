package operator

import (
	"time"

	"github.com/nekrassov01/ignoresync/internal/params"
)

/* General parameters */

const (
	// schemeVersion is included in AAD to allow future changes to the encryption
	// scheme while maintaining backward compatibility.
	schemeVersion = 1
)

/* Encryption parameters */

const (
	// chunkSize is the size of each chunk for AES-GCM encryption.
	chunkSize = 5 * 1024 * 1024

	// keySize is the size of encryption keys in bytes.
	keySize = 32

	// baseNonceSize is the size of base nonce in bytes.
	baseNonceSize = 12

	// chunkNonceSize is the nonce size for AES-GCM in bytes.
	chunkNonceSize = 12

	// tagSize is the authentication tag size for AES-GCM in bytes.
	tagSize = 16
)

/* Key derivation parameters */

const (
	// infoLocalKey is the info string for deriving local key from master key.
	infoLocalKey = params.CommandName + "localkey"

	// infoWrapKey is the info string for deriving wrap key from local key and cloud key.
	infoWrapKey = params.CommandName + "wrapkey"
)

/* S3 metadata keys */

const (
	// metadataKeySchemeVersion is the S3 user-metadata key for the encryption/decryption scheme version.
	metadataKeySchemeVersion = params.CommandName + "-scheme-version"

	// metadataKeyKeyID is the S3 user-metadata key for the master key ID.
	metadataKeyKeyID = params.CommandName + "-master-key-id"

	// metadataKeyCloudKey is the S3 user-metadata key for the encrypted cloud key (base64-encoded).
	metadataKeyCloudKey = params.CommandName + "-cloud-key"

	// metadataKeyRepoKey is the S3 user-metadata key for the encrypted repo key (base64-encoded).
	metadataKeyRepoKey = params.CommandName + "-repo-key"

	// metadataKeyDataKey is the S3 user-metadata key for the encrypted data key (base64-encoded).
	metadataKeyDataKey = params.CommandName + "-data-key"

	// metadataKeyBaseNonce is the S3 user-metadata key for the base nonce used in encryption (base64-encoded).
	metadataKeyBaseNonce = params.CommandName + "-base-nonce"

	// metadataKeyChunkSize is the S3 user-metadata key for the chunk size used in encryption.
	metadataKeyChunkSize = params.CommandName + "-chunk-size"

	// metadataKeyGitUser is the S3 user-metadata key for the git user name.
	metadataKeyGitUser = params.CommandName + "-git-user"
)

/* S3 uploader parameters */

const (
	// uploadMaxWaitDur is the maximum duration to wait for S3 upload completion.
	uploadMaxWaitDur = time.Minute

	// uploadPartSize is the size of each part for S3 multipart upload.
	uploadPartSize = chunkSize

	// uploadConcurrency is the number of concurrent uploads for S3 multipart upload.
	uploadConcurrency = 4
)

/* Retry parameters */

const (
	// maxRetryAttemptsConditionalError is the maximum number of retry attempts for "PreconditionFailed" and "ConditionalRequestConflict".
	maxRetryAttemptsConditionalError = 3
)

/* Repository parameters */

var (
	// gitDir indicates the path of the .git directory.
	// It can be replaced during testing.
	gitDir = ".git"

	// defaultPatterns defines the default target patterns to include in the bundle.
	defaultPatterns = []string{}

	// maxFileNum defines the maximum number of files to walk in the repository.
	maxFileNum = 1024 * 8

	// maxFileSize defines the maximum size of a single file to include in the bundle.
	maxFileSize int64 = 1024 * 1024 * 10

	// maxRepoSize defines the maximum size of the repository to process.
	maxRepoSize int64 = 1024 * 1024 * 500

	// maxDiffSize defines the maximum size to load into memory for diff display.
	maxDiffSize int64 = 1024 * 1024 * 1
)
