package params

const (
	// CommandName is the name of the application.
	CommandName = "ignoresync"

	// CanonicalName is the canonical name of the application.
	CanonicalName = "IgnoreSync"

	// LogLabel is the label of the application logger.
	LogLabel = "IGNORESYNC"
)

const (
	// DefaultUserName is the default user name used when the actual user name cannot be determined.
	DefaultUserName = "unknown"

	// DefaultRemoteName is the default remote name used when the actual remote name cannot be determined.
	DefaultRemoteName = "origin"

	// DefaultTempPattern is the default pattern for temporary files and directories.
	DefaultTempPattern = "ignoresync-temp-*"

	// KeyIDSize is the size of the master key ID in bytes.
	KeyIDSize = 8

	// MasterKeySize is the size of the master key in bytes.
	MasterKeySize = 16

	// CredentialSize is the total size of the credential in bytes.
	CredentialSize = KeyIDSize + MasterKeySize
)

const (
	// EnvCredential is the name of the environment variable that contains the credential string.
	// If this environment variable is set, its value is used as the credential to derive
	// state every time, thus avoiding the use of a keystore. This is intended for use in
	// a CI environment and should not be used for normal use.
	// #nosec G101
	EnvCredential = "IGNORESYNC_CREDENTIAL"

	// EnvLogLevel is the name of the environment variable that contains the log level.
	// This environment variable can be used in place of the CLI option `--log-level`.
	EnvLogLevel = "IGNORESYNC_LOG_LEVEL"

	// EnvRemoteName is the name of the environment variable that specifies the remote repository name other than `origin`.
	// This environment variable can be used in place of the CLI option `--remote-name`.
	EnvRemoteName = "IGNORESYNC_REMOTE_NAME"

	// EnvTargetPatterns is the name of the environment variable that specifies the target patterns.
	// This environment variable can be used in place of the CLI option `--patterns`.
	EnvTargetPatterns = "IGNORESYNC_TARGET_PATTERNS"

	// EnvAWSProfile is the name of the environment variable that contains the AWS profile name.
	// This environment variable can be used in place of the CLI option `--profile`.
	// Prefix "IGNORESYNC_" is added to avoid conflict with the AWS CLI environment variable "AWS_PROFILE".
	EnvAWSProfile = "IGNORESYNC_AWS_PROFILE"

	// EnvAWSRegion is the name of the environment variable that contains the AWS region name.
	// This environment variable can be used in place of the CLI option `--region`.
	// Prefix "IGNORESYNC_" is added to avoid conflict with the AWS CLI environment variable "AWS_REGION".
	EnvAWSRegion = "IGNORESYNC_AWS_REGION"
)
