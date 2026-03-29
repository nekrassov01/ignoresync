package operator

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/nekrassov01/ignoresync"
	"github.com/nekrassov01/ignoresync/manager"
	"golang.org/x/sync/errgroup"
)

// keySet is data transfer object that contains the required keys and ETag.
type keySet struct {
	rk   []byte  // RepoKey
	wrk  []byte  // WrappedRepoKey
	wck  []byte  // WrappedCloudKey
	etag *string // S3 ETag
}

// PushFiles archives files, encrypts them, and uploads them to S3.
func (o *Operator) PushFiles(ctx context.Context, state *manager.State) error {
	for i := range MaxRetryAttemptsConditionalError {
		attempt := func() error {
			// Step 1: Create RepoKey and wrapping keys
			set, err := o.loadOrCreateKeySet(ctx, state, o.prefixFiles)
			if err != nil {
				return err
			}
			defer clear(set.rk)

			// Step 2: Archive files to tar.gz
			body := o.bundleFiles()
			if o.dryrun {
				defer func() {
					_ = body.Close()
				}()
				if _, err := io.Copy(io.Discard, body); err != nil {
					return NewArchiveError(err)
				}
				return nil
			}

			// Step 3: Push archived files
			return o.push(ctx, state, body, o.prefixFiles, set)
		}()
		if attempt == nil {
			return nil
		}

		if isConditionalError(attempt) && i < MaxRetryAttemptsConditionalError-1 {
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			continue
		}
		return attempt
	}
	return nil
}

// PushPatterns compresses patterns, encrypts them, and uploads them to S3.
func (o *Operator) PushPatterns(ctx context.Context, state *manager.State, patterns []string) error {
	for i := range MaxRetryAttemptsConditionalError {
		attempt := func() error {
			// Step 1: Create RepoKey and wrapping keys
			set, err := o.loadOrCreateKeySet(ctx, state, o.prefixPatterns)
			if err != nil {
				return err
			}
			defer clear(set.rk)

			// Step 2: Archive patterns to tar.gz
			body := o.bundlePatterns(patterns)

			// Step 3: Push archived patterns
			return o.push(ctx, state, body, o.prefixPatterns, set)
		}()
		if attempt == nil {
			return nil
		}

		if isConditionalError(attempt) && i < MaxRetryAttemptsConditionalError-1 {
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			continue
		}
		return attempt
	}
	return nil
}

// push is a helper that encrypts the given body and uploads it to S3 with the given key.
func (o *Operator) push(ctx context.Context, state *manager.State, body io.ReadCloser, prefix string, set *keySet) (err error) {
	// Step 1: Generate DataKey
	dk, err := generateKey()
	if err != nil {
		_ = body.Close()
		return NewGenerateError(err)
	}
	defer clear(dk)

	// Step 2: Generate base nonce
	baseNonce, err := generateNonce()
	if err != nil {
		_ = body.Close()
		return NewGenerateError(err)
	}

	// Step 3: Encrypt DataKey with RepoKey
	aad, err := buildAAD(prefix, "data", baseNonce)
	if err != nil {
		_ = body.Close()
		return NewEncryptError(err)
	}
	wdk, err := encryptKey(set.rk, dk, aad)
	if err != nil {
		_ = body.Close()
		return NewEncryptError(err)
	}

	// Step 4: Encrypt body with DataKey
	body, err = encryptBody(body, dk, baseNonce, aad, chunkSize)
	if err != nil {
		return NewEncryptError(err)
	}
	defer func() {
		closeErr := body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// Step 5: Create metadata
	gitUser := ignoresync.DefaultUserName
	if o.repo != nil && o.repo.user != "" {
		gitUser = o.repo.user
	}
	metadata := map[string]string{
		metadataKeySchemeVersion: strconv.Itoa(schemeVersion),
		metadataKeyKeyID:         state.KeyID,
		metadataKeyCloudKey:      base64.StdEncoding.EncodeToString(set.wck),
		metadataKeyRepoKey:       base64.StdEncoding.EncodeToString(set.wrk),
		metadataKeyDataKey:       base64.StdEncoding.EncodeToString(wdk),
		metadataKeyBaseNonce:     base64.StdEncoding.EncodeToString(baseNonce),
		metadataKeyChunkSize:     strconv.Itoa(chunkSize),
		metadataKeyGitUser:       gitUser,
	}

	// Step 6: Upload encrypted body and metadata
	if err := o.upload(ctx, state, body, prefix, metadata, set.etag); err != nil {
		return NewUploadError(err)
	}

	return nil
}

// PullFiles downloads files from S3, decrypts them, and unpacks them.
func (o *Operator) PullFiles(ctx context.Context, state *manager.State) (err error) {
	// Step 1: Load RepoKey from S3
	set, err := o.loadKeySet(ctx, state, o.prefixFiles)
	if err != nil {
		if isObjectNotFound(err) {
			return nil
		}
		return err
	}
	defer clear(set.rk)

	// Step 2: Pull encrypted files from S3
	body, err := o.pull(ctx, state, o.prefixFiles, set.rk)
	if err != nil {
		if isObjectNotFound(err) {
			return nil
		}
		return err
	}
	defer func() {
		closeErr := body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// Step 3: Restore files from decrypted body
	if err := o.restoreFiles(body); err != nil {
		return NewRestoreError(err)
	}

	return nil
}

// PullPatterns downloads patterns from S3, decrypts them, and unpacks them.
func (o *Operator) PullPatterns(ctx context.Context, state *manager.State) ([]string, error) {
	// Step 1: Load RepoKey from S3
	set, err := o.loadKeySet(ctx, state, o.prefixPatterns)
	if err != nil {
		if isObjectNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	defer clear(set.rk)

	// Step 2: Pull encrypted patterns from S3
	body, err := o.pull(ctx, state, o.prefixPatterns, set.rk)
	if err != nil {
		if isObjectNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		closeErr := body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// Step 3: Restore patterns from decrypted body
	patterns, err := o.restorePatterns(body)
	if err != nil {
		return nil, NewRestoreError(err)
	}

	return patterns, nil
}

// pull is a helper that downloads an object from S3 with the given key, decrypts it, and returns its body.
func (o *Operator) pull(ctx context.Context, state *manager.State, prefix string, rk []byte) (io.ReadCloser, error) {
	// Step 1: Download encrypted object and metadata from S3
	body, m, _, err := o.download(ctx, state, prefix)
	if err != nil {
		if isObjectNotFound(err) {
			return nil, err
		}
		return nil, NewDownloadError(err)
	}

	// Step 2: Decrypt DataKey with RepoKey
	aad, err := buildAAD(prefix, "data", m.BaseNonce)
	if err != nil {
		_ = body.Close()
		return nil, NewDecryptError(err)
	}
	dk, err := decryptKey(rk, m.WrappedDataKey, aad)
	if err != nil {
		_ = body.Close()
		return nil, NewDecryptError(err)
	}
	defer clear(dk)

	// Step 3: Decrypt body with DataKey
	body, err = decryptBody(body, dk, m.BaseNonce, aad, m.ChunkSize)
	if err != nil {
		return nil, NewDecryptError(err)
	}

	// Step 4: Return decrypted body
	return body, nil
}

// Delete deletes files and patterns only after confirming that the RepoKey can be decrypted.
func (o *Operator) Delete(ctx context.Context, state *manager.State) error {
	// Step 1: Load RepoKey from S3
	_, err := o.loadKeySet(ctx, state, o.prefixFiles)
	if err != nil {
		return err
	}

	// Step 2: Delete objects from S3
	// If the repoKey cannot be retrieved, do not delete it
	eg, ctx := errgroup.WithContext(ctx)
	for _, prefix := range []string{o.prefixFiles, o.prefixPatterns} {
		eg.Go(func() error {
			if err := o.delete(ctx, state, prefix); err != nil {
				return NewDeleteError(err)
			}
			return nil
		})
	}

	return eg.Wait()
}

// Rewrap wraps repo key with the current master key for both files and patterns.
// It preserves object body and other metadata but updates the master key id and
// wrapped repo key atomically using If-Match.
func (o *Operator) Rewrap(ctx context.Context, state *manager.State) error {
	eg, ctx := errgroup.WithContext(ctx)
	for _, prefix := range []string{o.prefixFiles, o.prefixPatterns} {
		eg.Go(func() error {
			for i := range MaxRetryAttemptsConditionalError {
				attempt := func() error {
					// Step 1: Download encrypted object and metadata from S3
					body, meta, etag, err := o.download(ctx, state, prefix)
					if err != nil {
						if isObjectNotFound(err) {
							return nil
						}
						return NewDownloadError(err)
					}

					// Step 2: Validate metadata KeyID matches current state
					if meta.KeyID == state.KeyID {
						return nil
					}
					oldMK, ok := state.MasterKeys[meta.KeyID]
					if !ok {
						return NewValidateError(fmt.Errorf("required master key id %q not available in local state", meta.KeyID))
					}

					// Step 3: Decrypt CloudKey with AWS KMS Decrypt API
					oldCK, err := o.decryptCloudKey(ctx, state, meta.WrappedCloudKey)
					if err != nil {
						return NewDecryptError(err)
					}
					defer clear(oldCK)

					// Step 4: Derive LocalKey from old MasterKey
					oldLK, err := deriveLocalKey(oldMK)
					if err != nil {
						return NewDeriveError(err)
					}
					defer clear(oldLK)

					// Step 5: Derive WrapKey from LocalKey and CloudKey
					oldWK, err := deriveWrapKey(oldLK, oldCK)
					if err != nil {
						return NewDeriveError(err)
					}
					defer clear(oldWK)

					// Step 6: Decrypt RepoKey with WrapKey
					oldAAD, err := buildAAD(prefix, "repo", nil)
					if err != nil {
						return NewDecryptError(err)
					}
					rk, err := decryptKey(oldWK, meta.WrappedRepoKey, oldAAD)
					if err != nil {
						return NewDecryptError(err)
					}
					defer clear(rk)

					// Step 7: Derive LocalKey from new MasterKey
					newMK, ok := state.MasterKeys[state.KeyID]
					if !ok {
						return NewValidateError(fmt.Errorf("current master key id %q not present in local state", state.KeyID))
					}
					newLK, err := deriveLocalKey(newMK)
					if err != nil {
						return NewDeriveError(err)
					}
					defer clear(newLK)

					// Step 8: Derive WrapKey from new LocalKey and CloudKey
					newWK, err := deriveWrapKey(newLK, oldCK)
					if err != nil {
						return NewDeriveError(err)
					}
					defer clear(newWK)

					// Step 9: Encrypt RepoKey with new WrapKey
					newAAD, err := buildAAD(prefix, "repo", nil)
					if err != nil {
						return NewEncryptError(err)
					}
					newWRK, err := encryptKey(newWK, rk, newAAD)
					if err != nil {
						return NewEncryptError(err)
					}

					// Step 10: Upload encrypted body and metadata
					metadata := map[string]string{
						metadataKeySchemeVersion: strconv.Itoa(schemeVersion),
						metadataKeyKeyID:         state.KeyID,
						metadataKeyCloudKey:      base64.StdEncoding.EncodeToString(meta.WrappedCloudKey),
						metadataKeyRepoKey:       base64.StdEncoding.EncodeToString(newWRK),
						metadataKeyDataKey:       base64.StdEncoding.EncodeToString(meta.WrappedDataKey),
						metadataKeyBaseNonce:     base64.StdEncoding.EncodeToString(meta.BaseNonce),
						metadataKeyChunkSize:     strconv.Itoa(meta.ChunkSize),
						metadataKeyGitUser:       meta.GitUser,
					}
					if err := o.upload(ctx, state, body, prefix, metadata, etag); err != nil {
						_ = body.Close()
						return NewUploadError(err)
					}
					_ = body.Close()

					return nil
				}()

				if attempt == nil {
					return nil
				}

				if isConditionalError(attempt) && i < MaxRetryAttemptsConditionalError-1 {
					time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
					continue
				}
				return attempt
			}
			return nil
		})
	}

	return eg.Wait()
}

// loadOrCreateKeySet tries to load the RepoKey from S3, and if it fails due to object not found, it creates a new RepoKey and uploads it to S3.
func (o *Operator) loadOrCreateKeySet(ctx context.Context, state *manager.State, prefix string) (*keySet, error) {
	set, err := o.loadKeySet(ctx, state, prefix)
	if err == nil {
		return set, nil
	}
	if !isObjectNotFound(err) {
		return nil, err
	}
	return o.createKeySet(ctx, state, prefix)
}

// createKeySet generates a RepoKey through a series of workflows and returns it along with the necessary wrapping keys.
func (o *Operator) createKeySet(ctx context.Context, state *manager.State, prefix string) (*keySet, error) {
	// Step 1: Check MasterKey availability
	mk, ok := state.MasterKeys[state.KeyID]
	if !ok {
		return nil, NewValidateError(fmt.Errorf("current master key id %q not present in local state", state.KeyID))
	}

	// Step 2: Generate CloudKey using AWS KMS GenerateDataKey API
	ck, wck, err := o.generateCloudKey(ctx, state)
	if err != nil {
		return nil, NewGenerateError(err)
	}
	defer clear(ck)

	// Step 3: Derive LocalKey from MasterKey
	lk, err := deriveLocalKey(mk)
	if err != nil {
		return nil, NewDeriveError(err)
	}
	defer clear(lk)

	// Step 4: Derive WrapKey from LocalKey and CloudKey
	wk, err := deriveWrapKey(lk, ck)
	if err != nil {
		return nil, NewDeriveError(err)
	}
	defer clear(wk)

	// Step 5: Generate RepoKey
	rk, err := generateKey()
	if err != nil {
		return nil, NewGenerateError(err)
	}

	// Step 6: Encrypt RepoKey with WrapKey
	aad, err := buildAAD(prefix, "repo", nil)
	if err != nil {
		return nil, NewEncryptError(err)
	}
	wrk, err := encryptKey(wk, rk, aad)
	if err != nil {
		return nil, NewEncryptError(err)
	}

	// Step 7: Return RepoKey, WrappedWrapKey, and WrappedCloudKey
	return &keySet{
		rk:  rk,
		wrk: wrk,
		wck: wck,
	}, nil
}

// loadKeySet restores the RepoKey through a series of workflows and returns it along with the necessary wrapping keys.
func (o *Operator) loadKeySet(ctx context.Context, state *manager.State, prefix string) (*keySet, error) {
	// Step 1: Download and parse metadata from existing object in S3
	m, etag, err := o.head(ctx, state, prefix)
	if err != nil {
		return nil, NewDownloadError(err)
	}

	// Step 2: Check MasterKey availability
	if state.KeyID != m.KeyID {
		return nil, NewValidateError(fmt.Errorf("key id mismatch, rewrap required: current=%s, metadata=%s", state.KeyID, m.KeyID))
	}
	mk, ok := state.MasterKeys[m.KeyID]
	if !ok {
		return nil, NewValidateError(fmt.Errorf("required master key id %q not available in local state", m.KeyID))
	}

	// Step 3: Decrypt CloudKey using AWS KMS Decrypt API
	ck, err := o.decryptCloudKey(ctx, state, m.WrappedCloudKey)
	if err != nil {
		return nil, NewDecryptError(err)
	}
	defer clear(ck)

	// Step 4: Derive LocalKey from MasterKey
	lk, err := deriveLocalKey(mk)
	if err != nil {
		return nil, NewDeriveError(err)
	}
	defer clear(lk)

	// Step 5: Derive WrapKey from LocalKey and CloudKey
	wk, err := deriveWrapKey(lk, ck)
	if err != nil {
		return nil, NewDeriveError(err)
	}
	defer clear(wk)

	// Step 6: Decrypt RepoKey with WrapKey
	aad, err := buildAAD(prefix, "repo", nil)
	if err != nil {
		return nil, NewDecryptError(err)
	}
	rk, err := decryptKey(wk, m.WrappedRepoKey, aad)
	if err != nil {
		return nil, NewDecryptError(err)
	}

	// Step 7: Return RepoKey, WrappedWrapKey, WrappedCloudKey, and ETag
	return &keySet{
		rk:   rk,
		wrk:  m.WrappedRepoKey,
		wck:  m.WrappedCloudKey,
		etag: etag,
	}, nil
}
