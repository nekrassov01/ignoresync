package operator

import (
	"github.com/nekrassov01/ignoresync"
)

// ErrorKind represents the kind of error that occurred in the operator.
type ErrorKind int

const (
	// ErrorKindNone indicates no error.
	ErrorKindNone ErrorKind = iota

	// ErrorKindRepository indicates error in repository operations.
	ErrorKindRepository

	// ErrorKindGenerate indicates error in key generation process.
	ErrorKindGenerate

	// ErrorKindDerive indicates error in key derivation process.
	ErrorKindDerive

	// ErrorKindValidate indicates error in validation process.
	ErrorKindValidate

	// ErrorKindArchive indicates error in push pipeline archiving process.
	ErrorKindArchive

	// ErrorKindEncrypt indicates error in push pipeline encryption process.
	ErrorKindEncrypt

	// ErrorKindUpload indicates error in push pipeline upload process.
	ErrorKindUpload

	// ErrorKindDownload indicates error in pull pipeline download process.
	ErrorKindDownload

	// ErrorKindDecrypt indicates error in pull pipeline decryption process.
	ErrorKindDecrypt

	// ErrorKindRestore indicates error in pull pipeline restoration process.
	ErrorKindRestore

	// ErrorKindDelete indicates error in object deletion process.
	ErrorKindDelete

	// ErrorKindCleanup indicates error in cleanup process.
	ErrorKindCleanup

	// ErrorKindCommand indicates error in command execution process.
	ErrorKindCommand
)

// String returns the string representation of the ErrorKind.
func (k ErrorKind) String() string {
	switch k {
	case ErrorKindNone:
		return "no error"
	case ErrorKindRepository:
		return "repository error"
	case ErrorKindGenerate:
		return "generation error"
	case ErrorKindDerive:
		return "derivation error"
	case ErrorKindValidate:
		return "validation error"
	case ErrorKindArchive:
		return "archive error"
	case ErrorKindEncrypt:
		return "encrypt error"
	case ErrorKindUpload:
		return "upload error"
	case ErrorKindDownload:
		return "download error"
	case ErrorKindDecrypt:
		return "decrypt error"
	case ErrorKindRestore:
		return "restore error"
	case ErrorKindDelete:
		return "deletion error"
	case ErrorKindCleanup:
		return "cleanup error"
	case ErrorKindCommand:
		return "command error"
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
	return ignoresync.FormatError("operator", e.Kind.String(), e.Err)
}

// Unwrap returns the underlying error of the error.
func (e *Error) Unwrap() error {
	return e.Err
}

// NewRepositoryError constructs a repository error wrapping optional underlying error.
func NewRepositoryError(err error) error {
	return &Error{Kind: ErrorKindRepository, Err: err}
}

// NewGenerateError constructs a generation error wrapping optional underlying error.
func NewGenerateError(err error) error {
	return &Error{Kind: ErrorKindGenerate, Err: err}
}

// NewDeriveError constructs a derivation error wrapping optional underlying error.
func NewDeriveError(err error) error {
	return &Error{Kind: ErrorKindDerive, Err: err}
}

// NewValidateError constructs a validation error wrapping optional underlying error.
func NewValidateError(err error) error {
	return &Error{Kind: ErrorKindValidate, Err: err}
}

// NewArchiveError constructs an archive error wrapping optional underlying error.
func NewArchiveError(err error) error {
	return &Error{Kind: ErrorKindArchive, Err: err}
}

// NewEncryptError constructs an encrypt error wrapping optional underlying error.
func NewEncryptError(err error) error {
	return &Error{Kind: ErrorKindEncrypt, Err: err}
}

// NewUploadError constructs an upload error wrapping optional underlying error.
func NewUploadError(err error) error {
	return &Error{Kind: ErrorKindUpload, Err: err}
}

// NewDownloadError constructs a download error wrapping optional underlying error.
func NewDownloadError(err error) error {
	return &Error{Kind: ErrorKindDownload, Err: err}
}

// NewDecryptError constructs a decrypt error wrapping optional underlying error.
func NewDecryptError(err error) error {
	return &Error{Kind: ErrorKindDecrypt, Err: err}
}

// NewRestoreError constructs a restore error wrapping optional underlying error.
func NewRestoreError(err error) error {
	return &Error{Kind: ErrorKindRestore, Err: err}
}

// NewDeleteError constructs a delete error wrapping optional underlying error.
func NewDeleteError(err error) error {
	return &Error{Kind: ErrorKindDelete, Err: err}
}

// NewCleanupError constructs a cleanup error wrapping optional underlying error.
func NewCleanupError(err error) error {
	return &Error{Kind: ErrorKindCleanup, Err: err}
}

// NewCommandError constructs a command error wrapping optional underlying error.
func NewCommandError(err error) error {
	return &Error{Kind: ErrorKindCommand, Err: err}
}
