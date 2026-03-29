package manager

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/nekrassov01/ignoresync"
	"github.com/zalando/go-keyring"
)

// State holds the generated data state.
type State struct {
	KeyID      string            // The current master key ID
	MasterKeys map[string][]byte // The map of master key ID to master key bytes
	Account    string            //	The AWS account ID
	Region     string            //	The AWS region
	Bucket     *BucketInfo       //	The S3 bucket information
	SSEKey     *KMSKeyInfo       //	The server-side encryption key information
	CSEKey     *KMSKeyInfo       //	The client-side encryption key information
}

// BucketInfo holds information about the S3 bucket.
type BucketInfo struct {
	ARN arn.ARN
}

// KMSKeyInfo holds information about an encryption key.
type KMSKeyInfo struct {
	ARN               arn.ARN
	EncryptionContext map[string]string
}

// KeyEntry represents a master key ID and active status.
type KeyEntry struct {
	KeyID    string `json:"keyID"`
	IsActive bool   `json:"isActive"`
}

// AddCredential adds the new credential to the state and updates the keyring.
func (o *Manager) AddCredential(id string, key []byte) error {
	// Load stored state
	state, err := o.LoadState()
	if err != nil {
		return err
	}

	// If a credential with the same id already exists, ensure the key matches.
	if old, ok := state.MasterKeys[id]; ok {
		if !bytes.Equal(old, key) {
			return NewStateError(fmt.Errorf("key id conflict: %s", id))
		}
	}

	// Add or update credential in state
	state.KeyID = id
	state.MasterKeys[id] = key

	// Store state in the keyring
	if err := o.StoreState(state); err != nil {
		return err
	}

	return nil
}

// RemoveCredential removes the credential from the state and updates the keyring.
func (o *Manager) RemoveCredential(id string, key []byte) error {
	// Load stored state
	state, err := o.LoadState()
	if err != nil {
		return err
	}

	// Check if the credential to be deleted is currently active
	if state.KeyID == id {
		return NewStateError(errors.New("failed to delete active credential from state: rotate required"))
	}

	// If a credential with the same id already exists, ensure the key matches.
	if old, ok := state.MasterKeys[id]; ok {
		if !bytes.Equal(old, key) {
			return NewStateError(fmt.Errorf("same id, but master key does not match: %s", id))
		}
	}

	// Remove credential from state
	delete(state.MasterKeys, id)

	// Store state in the keyring
	if err := o.StoreState(state); err != nil {
		return err
	}

	return nil
}

// ListCredentials list the master key IDs.
// Sensitive information such as the master key will not be displayed.
func (o *Manager) ListCredentials() ([]KeyEntry, error) {
	// Load stored state
	state, err := o.LoadState()
	if err != nil {
		return nil, err
	}

	// Extract key IDs and active status
	entries := make([]KeyEntry, 0, len(state.MasterKeys))
	for id := range state.MasterKeys {
		key := KeyEntry{
			KeyID:    id,
			IsActive: id == state.KeyID,
		}
		entries = append(entries, key)
	}
	slices.SortFunc(entries, func(a, b KeyEntry) int {
		return strings.Compare(a.KeyID, b.KeyID)
	})
	return entries, nil
}

// EnsureState ensures the state is loaded from the keyring if cred is empty,
// otherwise generates a new state from the given credential.
func (o *Manager) EnsureState(cred string) (*State, error) {
	if cred == "" {
		state, err := o.LoadState()
		if err != nil {
			return nil, err
		}
		return state, nil
	}
	id, key, err := DecodeCredential(cred)
	if err != nil {
		return nil, err
	}
	state, err := o.GenerateState(id, key)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// StoreState stores the passed state in the keyring.
func (o *Manager) StoreState(state *State) error {
	buf := bytes.Buffer{}
	if err := gob.NewEncoder(&buf).Encode(state); err != nil {
		return NewStateError(fmt.Errorf("failed to store state: %w", err))
	}
	if err := keyring.Set(ignoresync.CanonicalName, o.OSUser, buf.String()); err != nil {
		return NewStateError(fmt.Errorf("failed to store state: %w", err))
	}
	return nil
}

// LoadState loads the passed state from the keyring.
// After loading data, verify that the current profile matches the loaded state.
func (o *Manager) LoadState() (*State, error) {
	state := &State{}
	s, err := keyring.Get(ignoresync.CanonicalName, o.OSUser)
	if err != nil {
		return nil, NewStateError(fmt.Errorf("failed to load state: %w", err))
	}
	r := strings.NewReader(s)
	if err := gob.NewDecoder(r).Decode(state); err != nil {
		return nil, NewStateError(fmt.Errorf("failed to load state: %w", err))
	}
	if o.Account != state.Account {
		return nil, NewStateError(fmt.Errorf("prevent operational error: unmatch aws account: %s", o.Account))
	}
	if o.Region != state.Region {
		return nil, NewStateError(fmt.Errorf("prevent operational error: unmatch aws region: %s", o.Region))
	}
	return state, nil
}

// DeleteState deletes the stored data state.
func (o *Manager) DeleteState() error {
	if err := keyring.Delete(ignoresync.CanonicalName, o.OSUser); err != nil {
		return NewStateError(fmt.Errorf("failed to delete state: %w", err))
	}
	return nil
}

// CheckStateExist checks if the state exists in the keyring.
func (o *Manager) CheckStateExist() (bool, error) {
	_, err := keyring.Get(ignoresync.CanonicalName, o.OSUser)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		return false, nil
	}
	return false, NewStateError(fmt.Errorf("failed to check state present: %w", err))
}

// GenerateState generates the state from the given master key.
// The state holds the informations to uniquely identify the environment.
func (o *Manager) GenerateState(id string, key []byte) (*State, error) {
	if id == "" {
		return nil, NewStateError(errors.New("no id provided"))
	}
	if len(key) != ignoresync.MasterKeySize {
		return nil, NewStateError(errors.New("invalid key length"))
	}
	bucket, err := o.generateBucketInfo(infoBucket)
	if err != nil {
		return nil, NewStateError(err)
	}
	sseKey, err := o.generateKMSKeyInfo(infoSSEKey, bucket.ARN.Resource)
	if err != nil {
		return nil, NewStateError(err)
	}
	cseKey, err := o.generateKMSKeyInfo(infoCSEKey, bucket.ARN.Resource)
	if err != nil {
		return nil, NewStateError(err)
	}
	state := &State{
		KeyID:      id,
		MasterKeys: map[string][]byte{id: key},
		Account:    o.Account,
		Region:     o.Region,
		Bucket:     bucket,
		SSEKey:     sseKey,
		CSEKey:     cseKey,
	}
	return state, nil
}

// generateBucketInfo generates BucketInfo from the given id, info, and bucket name.
func (o *Manager) generateBucketInfo(info []byte) (*BucketInfo, error) {
	hash := o.generateHash(info)
	prefix := "arn:aws:s3:::" // #nosec G101
	var b strings.Builder
	b.Grow(len(prefix) + len(hash))
	b.WriteString(prefix)
	b.WriteString(hash)
	bucketARN, err := arn.Parse(b.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate bucket ARN: %w", err)
	}
	bucketInfo := &BucketInfo{
		ARN: bucketARN,
	}
	return bucketInfo, nil
}

// generateKMSKeyInfo generates KMSKeyInfo from the given id, info, and bucket name.
func (o *Manager) generateKMSKeyInfo(info []byte, bucket string) (*KMSKeyInfo, error) {
	hash := o.generateHash(info)
	token0 := "arn:aws:kms:" // #nosec G101
	token1 := ":alias/"
	var b strings.Builder
	b.Grow(len(token0) + len(o.Region) + 1 + len(o.Account) + len(token1) + len(hash))
	b.WriteString(token0)
	b.WriteString(o.Region)
	b.WriteByte(':')
	b.WriteString(o.Account)
	b.WriteString(token1)
	b.WriteString(hash)
	keyARN, err := arn.Parse(b.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate kms key alias arn: %w", err)
	}
	b.Reset()
	b.Grow(len(bucket) + 1 + len(keyARN.Resource))
	b.WriteString(bucket)
	b.WriteByte('/')
	b.WriteString(keyARN.Resource)
	keyInfo := &KMSKeyInfo{
		ARN: keyARN,
		EncryptionContext: map[string]string{
			encryptionContextKey: b.String(),
		},
	}
	return keyInfo, nil
}

// generateHash generates a hash from the given id and info.
func (o *Manager) generateHash(info []byte) string {
	h := sha256.New()
	_, _ = h.Write(o.salt)
	_, _ = h.Write(info)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:resourceNameSize])
}
