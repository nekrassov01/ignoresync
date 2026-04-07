package manager

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/nekrassov01/ignoresync/internal/params"
)

// GenerateCredential generates the underlying master key ID and master key.
func GenerateCredential() (string, []byte, error) {
	cred := make([]byte, params.CredentialSize)
	if _, err := rand.Read(cred); err != nil {
		return "", nil, NewCredentialError(fmt.Errorf("failed to generate credential: %w", err))
	}
	id := hex.EncodeToString(cred[:params.KeyIDSize])
	key := cred[params.KeyIDSize:]
	return id, key, nil
}

// EncodeCredential encodes the credential to a hexadecimal string for display.
func EncodeCredential(id string, key []byte) string {
	return id + hex.EncodeToString(key)
}

// DecodeCredential decodes the hexadecimal string back to the credential bytes.
func DecodeCredential(cred string) (string, []byte, error) {
	idLen := hex.EncodedLen(params.KeyIDSize)
	if len(cred) != idLen+hex.EncodedLen(params.MasterKeySize) {
		return "", nil, NewCredentialError(fmt.Errorf("failed to validate credential: invalid key length"))
	}
	id := cred[:idLen]
	kh := cred[idLen:]
	_, err := decode(id, params.KeyIDSize)
	if err != nil {
		return "", nil, NewCredentialError(fmt.Errorf("failed to decode key id: %w", err))
	}
	key, err := decode(kh, params.MasterKeySize)
	if err != nil {
		return "", nil, NewCredentialError(fmt.Errorf("failed to decode master key: %w", err))
	}
	return id, key, nil
}

// ValidateCredential checks if the given input is a valid credential.
// Since this is used to validate the string passed from the prompt,
// it does not return detailed error messages.
func ValidateCredential(cred string) error {
	_, _, err := DecodeCredential(cred)
	if err != nil {
		return errors.New("invalid credential format")
	}
	return nil
}

// decode decodes the given input to bytes and checks its length.
func decode(s string, size int) ([]byte, error) {
	if len(s) != hex.EncodedLen(size) {
		return nil, fmt.Errorf("invalid key length")
	}
	dst, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid key format")
	}
	return dst, nil
}
