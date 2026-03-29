<p align="center">
  <h2 align="center">IGNORESYNC</h2>
  <p align="center">Your shadow repository for ignored files</p>
  <p align="center">
    <a href="https://github.com/nekrassov01/ignoresync/actions/workflows/ci.yaml"><img src="https://github.com/nekrassov01/ignoresync/actions/workflows/ci.yaml/badge.svg?branch=main" alt="CI" /></a>
    <a href="https://pkg.go.dev/github.com/nekrassov01/ignoresync"><img src="https://pkg.go.dev/badge/github.com/nekrassov01/ignoresync.svg" alt="Go Reference" /></a>
    <a href="https://goreportcard.com/report/github.com/nekrassov01/ignoresync"><img src="https://goreportcard.com/badge/github.com/nekrassov01/ignoresync" alt="Go Report Card" /></a>
    <img src="https://img.shields.io/github/license/nekrassov01/ignoresync" alt="LICENSE" />
    <a href="https://deepwiki.com/nekrassov01/ignoresync"><img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki" /></a>
  </p>
</p>

## Overview

`ignoresync` is a CLI tool for securely syncing files ignored by Git across multiple machines. It maintains data confidentiality using a dual-trust encryption model that combines locally derived keys with AWS KMS. It uses the `.gitignore` file in the repository as a baseline and only manages files that match the specified patterns. In effect, it adds a "shadow repository" for confidential files that live inside the working tree but are intentionally excluded from normal Git tracking.

> [!IMPORTANT]
> `ignoresync` is still experimental. Please use it at your own risk. The author assumes no responsibility for any loss, damage, or other harm arising from the use of this tool.

## Design Principles

- End-to-End Encryption: Encrypt files locally before upload and decrypt them only after download, ensuring data is never exposed in transit or at rest.
- Dual-Trust Encryption: Protect repository-scoped keys by combining a locally generated key with AWS KMS, requiring both local device authentication and AWS IAM authorization for decryption.
- Minimal Shared Secret Surface: Limit shared secrets to the MasterKey. Other key material is either derived locally or protected by AWS-managed mechanisms.
- Concealment: Keep ignoresync-specific configuration out of the repository by managing target patterns remotely, so that routine repository contents do not clearly advertise the tool's usage.
- Constant-Memory-Oriented Streaming: Process large file payloads in a streaming, constant-memory-oriented manner, while handling small control data such as keys and patterns in memory for simplicity.

## Architecture

`ignoresync` uses a hub-and-spoke model with a single central storage layer backed by AWS resources such as S3 and KMS, and each client connects to this environment for data synchronization. This remote environment acts as the Single Source of Truth (SSOT) for encrypted objects and metadata.

Each client is initialized using a shared master key. This key is stored in the OS keystore along with information used to uniquely identify the AWS resource. By maintaining this information as local state, there is no need to pass bucket names or key information as arguments for each operation.

During push and pull operations, the client merges local and AWS KMS-backed key material to restore the cryptographic context for secure processing. Data confidentiality remains absolute provided that neither the local key material nor the AWS operational permissions are compromised.

## Key Concepts

`ignoresync` has several key concepts and terms:

| Term        | Purpose                                                      |
| ----------- | ------------------------------------------------------------ |
| MasterKey   | Root shared secret                                           |
| State       | A set of credentials and resource identifiers stored locally |
| Environment | A set of deployed cloud resources                            |
| LocalKey    | Key derived from the MasterKey                               |
| CloudKey    | Key material obtained from AWS KMS                           |
| WrapKey     | Key derived from the LocalKey and CloudKey                   |
| RepoKey     | Repository-scoped key encryption key                         |
| DataKey     | Per-object key for body encryption                           |

### MasterKey

The master key is a random shared secret key generated during the bootstrap process. It serves as the root secret key for device activation and is used solely for deriving local keys. It is the only secret key that must be shared among devices participating in the same environment. When a rotation is performed, a new master key is registered on the local machine, and the latest key becomes active.

### State

The state is a local data structure stored in the OS keystore that contains all the information a device needs to participate in encryption and decryption. It represents the device’s activation credentials and defines the AWS account, region, bucket, and encryption keys to be used.

### Environment

The Environment is the set of cloud resources (S3 buckets, KMS keys) provisioned during `bootstrap`. It represents the remote infrastructure used to store and protect encrypted objects.

### LocalKey

The LocalKey is derived from the MasterKey and acts as the local trust anchor. It is never uploaded and is required to derive the key that unwraps the repository-scoped key.

### CloudKey

The CloudKey is obtained from AWS KMS as data key material and provides a cloud-managed cryptographic protection layer. Its wrapped form is stored remotely, while the plaintext is used only during key derivation.

### WrapKey

The WrapKey is derived from the LocalKey and CloudKey. It is used to wrap and unwrap the RepoKey, combining local trust and AWS authorization into a unified initialization of the repository's cryptographic context.

### RepoKey

The RepoKey is a random repository-scoped key stored remotely only in wrapped form. It is bound to the repository identity and is used to wrap the per-object DataKey for files and patterns.

### DataKey

The DataKey is a random, per-object symmetric key generated for each upload. It is used exclusively for content encryption and is never stored in plaintext.

## Pipelines

`ignoresync` executes a series of processing pipelines based on the issued commands.

### Bootstrap Pipeline

The bootstrap pipeline activates machines and provisions environments. It is run only once per environment:

```text
+----------------------------------------+
| Start Bootstrap                        |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Resolve AWS Account/Region using STS   |
| GetCallerIdentity API                  |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Ask user to confirm account and region |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Generate MasterKey                     |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Generate State:                        |
| Derive bucket and key identifiers      |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Deploy Environment:                    |
| Create and deploy CloudFormation stack |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Store State:                           |
| Persist state in the OS keystore       |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Display MasterKey once for secure      |
| sharing                                |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Finish Bootstrap                       |
|                                        |
+----------------------------------------+
```

### Push Pipeline

The Push Pipeline securely uploads files and patterns:

```text
+----------------------------------------+
| Start Push                             |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Load State from local keystore         |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Get WrappedCloudKey, WrappedRepoKey,   |
| and WrappedDataKey from S3 metadata    |
+----------------------------------------+
                    |
                    +--------------------------------------------+
                    |                                            |
            (metadata exists)                        (metadata does not exist)
                    |                                            |
                    v                                            v
+----------------------------------------+   +----------------------------------------+
| Decrypt WrappedCloudKey using KMS      |   | Get CloudKey and WrappedCloudKey using |
| Decrypt API                            |   | KMS GenerateDataKey API                |
+----------------------------------------+   +----------------------------------------+
                    |                                            |
                    v                                            v
+----------------------------------------+   +----------------------------------------+
| Derive LocalKey from MasterKey         |   | Derive LocalKey from MasterKey         |
|                                        |   |                                        |
+----------------------------------------+   +----------------------------------------+
                    |                                            |
                    v                                            v
+----------------------------------------+   +----------------------------------------+
| Derive WrapKey from LocalKey and       |   | Derive WrapKey from LocalKey and       |
| CloudKey                               |   | CloudKey                               |
+----------------------------------------+   +----------------------------------------+
                    |                                            |
                    v                                            v
+----------------------------------------+   +----------------------------------------+
| Decrypt WrappedRepoKey with WrapKey    |   | Generate RepoKey                       |
|                                        |   |                                        |
+----------------------------------------+   +----------------------------------------+
                    |                                            |
                    |                                            v
                    |                        +----------------------------------------+
                    |                        | Wrap RepoKey with WrapKey              |
                    |                        |                                        |
                    |                        +----------------------------------------+
                    |                                            |
                    |<-------------------------------------------+
                    |
                    v
+----------------------------------------+
| Compress files into archive            |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Generate DataKey                       |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Wrap DataKey with RepoKey              |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Encrypt archive with DataKey           |
| (AES-GCM chunked)                      |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Upload archive to S3 with metadata     |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Finish Push                            |
|                                        |
+----------------------------------------+
```

### Pull Pipeline

The Pull Pipeline securely downloads files and patterns:

```text
+----------------------------------------+
| Start Pull                             |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Load State from local keystore         |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Get WrappedCloudKey, WrappedRepoKey,   |
| and WrappedDataKey from S3 metadata    |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Decrypt WrappedCloudKey using KMS      |
| Decrypt API                            |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Derive LocalKey from MasterKey         |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Derive WrapKey from LocalKey and       |
| CloudKey                               |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Decrypt WrappedRepoKey with WrapKey    |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Download archive and metadata from S3  |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Get WrappedRepoKey from metadata and   |
| unwrap DataKey from metadata           |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Decrypt archive with DataKey           |
| (AES-GCM chunked)                      |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Restore files from archive             |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Finish Pull                            |
|                                        |
+----------------------------------------+
```

### Rewrap Pipeline

```text
+----------------------------------------+
| Start Rewrap                           |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Load State from local keystore         |
|                                        |
+----------------------------------------+
                    |
                    v
+----------------------------------------+
| Download archive and WrappedCloudKey,  |
| WrappedRepoKey, and WrappedDataKey     |
+----------------------------------------+
                    |
                    | Skip if match
                    +-----------------------------+
   Run if not match |                             |
                    v                             |
+----------------------------------------+        |
| Decrypt WrappedCloudKey using KMS      |        |
| Decrypt API                            |        |
+----------------------------------------+        |
                    |                             |
                    v                             |
+----------------------------------------+        |
| Derive current LocalKey from current   |        |
| MasterKey                              |        |
+----------------------------------------+        |
                    |                             |
                    v                             |
+----------------------------------------+        |
| Derive current WrapKey from current    |        |
| LocalKey and CloudKey                  |        |
+----------------------------------------+        |
                    |                             |
                    v                             |
+----------------------------------------+        |
| Decrypt WrappedRepoKey with WrapKey    |        |
|                                        |        |
+----------------------------------------+        |
                    |                             |
                    v                             |
+----------------------------------------+        |
| Derive new LocalKey from new MasterKey |        |
|                                        |        |
+----------------------------------------+        |
                    |                             |
                    v                             |
+----------------------------------------+        |
| Derive new WrapKey from new LocalKey   |        |
| and CloudKey                           |        |
+----------------------------------------+        |
                    |                             |
                    v                             |
+----------------------------------------+        |
| Wrap RepoKey with WrapKey              |        |
|                                        |        |
+----------------------------------------+        |
                    |                             |
                    v                             |
+----------------------------------------+        |
| Upload archive to S3 with metadata     |        |
|                                        |        |
+----------------------------------------+        |
                    |                             |
                    |<----------------------------+
                    |
                    v
+----------------------------------------+
| Finish Rewrap                          |
|                                        |
+----------------------------------------+
```

## Commands

The CLI commands and their functions are as follows:

| Command      | Description                                                                 | Scope      |
| ------------ | --------------------------------------------------------------------------- | ---------- |
| `bootstrap`  | Provisions a new Environment and generates the initial MasterKey and State. | Global     |
| `check`      | Checks the local State and the health of remote resources.                  | Global     |
| `activate`   | Add specidfied MasterKey to the local machine.                              | Local      |
| `deactivate` | Removes specified MasderKey from the local machine.                         | Local      |
| `list`       | List the key IDs stored in the State.                                       | Local      |
| `rotate`     | Generate new MasterKey and add it to the State.                             | Local      |
| `leave`      | Completely delete that state from the local machine.                        | Local      |
| `push`       | Archives, encrypts, and uploads files to S3.                                | Repository |
| `pull`       | Downloads, decrypts, and restores files from S3.                            | Repository |
| `rm`         | Deletes the files and patterns from S3.                                     | Repository |
| `set`        | Sets and uploads target patterns.                                           | Repository |
| `preview`    | Decrypts and previews files and patterns without writing to disk.           | Repository |
| `rewrap`     | Re-encrypt files and patterns using the latest MasterKey in the State.      | Repository |

### Bootstrap

The `bootstrap` command runs the bootstrap pipeline. Before it runs, it displays the account and region of the current profile and asks whether you want to use them for the environment. You can explicitly specify the account and region with the `--profile` and `--region` options. If they are not specified, the default profile is loaded. If no region is specified, it falls back to `us-east-1`.

```sh
ignoresync bootstrap --profile dev --region us-east-1
```

### Check

The `check` command is a non-destructive validation tool. It concurrently performs the following checks and returns either OK or NG:

- Field-level validation of the local state
- CloudFormation stack status and stability
- S3 bucket accessibility via HEAD, PUT, GET, and DELETE operations
- KMS key health via round-trip encryption and decryption

An NG result may indicate that the bootstrap process was not completed correctly, resources were modified after bootstrap, or the current AWS profile is incorrect. Note that the `check` command is intended for functional validation and is not an exhaustive test suite.

```sh
ignoresync check --profile dev
```

### Activate

The `activate` command generates a state using the specified master key and saves it to the OS keystore, thereby joining the local machine to the existing environment. If an incorrect master key is entered, an incorrect local key will be generated, causing all encryption and decryption operations to fail. If a state already exists, this command attempts to register the next new master key. The most recent key becomes active.

```sh
ignoresync activate --profile dev
```

### Deactivate

The `deactivate` command removes the specified master key from State. Attempting to remove incorrect key information will result in an error. You cannot remove a key that is currently active.

```sh
ignoresync deactivate --profile dev
```

### List

The `list` command retrieves a list of key IDs from the key information stored in the State and returns it in JSON format. It also displays the currently active key.

```sh
ignoresync list --profile dev
```

### Rotate

The `rotate` command generates a new master key, registers it with State, and activates it. Please note that this command is intended solely for rotating keys on the local machine. To continue encrypting and decrypting data, you must run the `rewrap` command. It is the user’s responsibility to securely share the latest master key and ensure that all users have the most recent and correct key.

```sh
ignoresync rotate --profile dev
```

### Leave

The `leave` command completely removes State from your local machine. This disconnects you from the environment. Please note that all master keys registered with State will be deleted.

```sh
ignoresync rotate --profile dev
```

### Push

The `push` command runs the push pipeline and securely uploads files that match the configured patterns to the repository. If the remote name of the repository is not "origin", you can specify it with the `--remote` option. If omitted, it defaults to "origin". This also applies to other commands. The repository identity includes the service domain, such as `github.com/owner/repo`. You can also enable the `--dry-run` option to check the target files without uploading them.

```sh
ignoresync push --profile dev --remote not-origin --dry-run
```

### Pull

The `pull` command runs the pull pipeline and securely downloads files from the environment that match the configured patterns. If there are differences in file content or permissions between the local and remote versions, the user is prompted to confirm whether to overwrite the file. If the `--overwrite` option is enabled, the differences are ignored and the file is overwritten.

```sh
ignoresync pull --profile dev --overwrite
```

### Rm

The `rm` command removes the files and the patterns from the environment. This command is allowed only when the current machine can successfully decrypt the RepoKey.

```sh
ignoresync rm --profile dev
```

### Set

The `set` command takes a JSON array of target patterns for your repository and uploads them to your environment. This ensures that target patterns are centrally managed in your environment and leave no traces in your repository. You can also remove patterns entirely by sending an empty JSON array. Note that `ignoresync` only targets files that match both patterns ignored in `.gitignore` and patterns set with `set`. This means that if you don't run `set`, the files won't be included in the target. You also need to be careful not to accidentally include hierarchical structures such as `node_modules`. You can always simulate this with `ignoresync push --dry-run`.

```sh
ignoresync set '["**/.env*","**/*/bin"]' --profile dev
```

### Preview

The `preview` command runs a pull pipeline and securely downloads files and patterns for preview purposes only. It lists the target files and patterns and discards them without writing them to disk.

```sh
ignoresync preview --profile dev
```

### Rewrap

The `rewrap` command runs the rewrap pipeline to rotate RepoKey protection under a new MasterKey; updates wrapped keys in metadata. After running the `rotate` command, be sure to run `rewrap` on the local machine where you executed the command.

```sh
ignoresync rewrap --profile dev
```

## Usage for CI

`ignoresync` can also be used in CI. Before accessing the keystore, it checks whether the environment variable `IGNORESYNC_CREDENTIAL` is set. If a value is set, it validates its format and, if successful, derives the state directly from that key instead of loading it from the local keystore. Commands that would create, update, or delete local keystore state are skipped. We strongly recommend storing `IGNORESYNC_CREDENTIAL` as a Secret with your CI provider (e.g., GitHub Actions Secrets) to avoid exposing it in build logs.

The behavior of each subcommand in CI mode is as follows:

| Command      | Behavior in CI mode                                                                                                       |
| ------------ | ------------------------------------------------------------------------------------------------------------------------- |
| `bootstrap`  | Skipped. No environment is created, and no local state is stored.                                                         |
| `check`      | Runs normally. The state is derived from `IGNORESYNC_CREDENTIAL` instead of being loaded from the keystore.               |
| `activate`   | Skipped. No local state is stored.                                                                                        |
| `deactivate` | Skipped. No local state is deleted.                                                                                       |
| `list`       | Skipped. No local state is rotated.                                                                                       |
| `rotate`     | Skipped. No local state is rotated.                                                                                       |
| `leave`      | Skipped. No local state is rotated.                                                                                       |
| `push`       | Runs normally. The state is derived from `IGNORESYNC_CREDENTIAL` instead of being loaded from the keystore.               |
| `pull`       | Runs normally. The state is derived from `IGNORESYNC_CREDENTIAL`, and `--overwrite` is effectively enabled automatically. |
| `rm`         | Runs normally. The state is derived from `IGNORESYNC_CREDENTIAL` instead of being loaded from the keystore.               |
| `set`        | Runs normally. The state is derived from `IGNORESYNC_CREDENTIAL` instead of being loaded from the keystore.               |
| `preview`    | Runs normally. The state is derived from `IGNORESYNC_CREDENTIAL` instead of being loaded from the keystore.               |
| `rewrap`     | Runs normally. The state is derived from `IGNORESYNC_CREDENTIAL` instead of being loaded from the keystore.               |

## Advanced Usage

One effective way to use `ignoresync` is to prepare a dedicated repository that acts as an inventory of personal confidential files. In this style, the repository itself contains almost nothing except a `.gitignore` file and files that are ignored by Git. You then use `ignoresync set` to define which ignored files should actually be synchronized.

This approach works well when you want a single place to manage private artifacts such as local environment files, personal certificates, SSH-related materials, or other machine-specific secrets. The repository remains minimal, while the target patterns managed in the environment serve as the catalog of what should be shared across your machines.

Because `ignoresync` only targets files matched by both `.gitignore` and the configured target patterns, this usage keeps the repository clean and makes synchronization intent explicit without introducing ignoresync-specific configuration files into the repository.

## Installation

Install with Homebrew

```sh
brew install nekrassov01/tap/ignoresync
```

Install with go

```sh
go install github.com/nekrassov01/ignoresync@latest
```

Or download the binary from [releases](https://github.com/nekrassov01/ignoresync/releases)

## Shell completion

Supported Shells are as follows:

- bash
- zsh
- fish
- pwsh

```sh
ignoresync completion bash|zsh|fish|pwsh
```

## Todo

- [x] Support CI mode
- [x] Support master key rotation
- [ ] Improve release workflows
- [ ] Add action.yaml
- [ ] Add E2E testing

## Author

[nekrassov01](https://github.com/nekrassov01)

## License

[MIT](https://github.com/nekrassov01/ignoresync/blob/main/LICENSE)
