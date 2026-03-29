package operator

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

// metadata represents the paramete stored in S3 object metadata.
type metadata struct {
	SchemeVersion   int    `json:"schemeVersion"`
	KeyID           string `json:"KeyID"`
	WrappedCloudKey []byte `json:"wrappedCloudKey"`
	WrappedRepoKey  []byte `json:"wrappedRepoKey"`
	WrappedDataKey  []byte `json:"wrappedDataKey"`
	BaseNonce       []byte `json:"baseNonce"`
	ChunkSize       int    `json:"chunkSize"`
	GitUser         string `json:"gitUser"`
}

// parseMetadata parses metadata from S3 object and returns a metadata struct.
func parseMetadata(m map[string]string) (*metadata, error) {
	schemeVersion, err := parseMetadataNumber(m, metadataKeySchemeVersion)
	if err != nil {
		return nil, err
	}

	KeyID, ok := m[metadataKeyKeyID]
	if !ok || KeyID == "" {
		return nil, fmt.Errorf("metadata key not found: %s", metadataKeyKeyID)
	}

	wrappedCloudKey, err := parseMetadataBase64(m, metadataKeyCloudKey)
	if err != nil {
		return nil, err
	}

	wrappedRepoKey, err := parseMetadataBase64(m, metadataKeyRepoKey)
	if err != nil {
		return nil, err
	}

	wrappedDataKey, err := parseMetadataBase64(m, metadataKeyDataKey)
	if err != nil {
		return nil, err
	}

	chunkSize, err := parseMetadataNumber(m, metadataKeyChunkSize)
	if err != nil {
		return nil, err
	}

	baseNonce, err := parseMetadataBase64(m, metadataKeyBaseNonce)
	if err != nil {
		return nil, err
	}

	gitUser, ok := m[metadataKeyGitUser]
	if !ok || gitUser == "" {
		return nil, fmt.Errorf("metadata key not found: %s", metadataKeyGitUser)
	}

	data := &metadata{
		SchemeVersion:   schemeVersion,
		KeyID:           KeyID,
		WrappedCloudKey: wrappedCloudKey,
		WrappedRepoKey:  wrappedRepoKey,
		WrappedDataKey:  wrappedDataKey,
		BaseNonce:       baseNonce,
		ChunkSize:       chunkSize,
		GitUser:         gitUser,
	}

	return data, nil
}

// parseMetadataNumber retrieves a numeric metadata value by key and converts it to int.
func parseMetadataNumber(m map[string]string, k string) (int, error) {
	s, ok := m[k]
	if !ok || s == "" {
		return 0, fmt.Errorf("metadata key not found: %s", k)
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid number in metadata key %s: %w", k, err)
	}
	return v, nil
}

// parseMetadataBase64 retrieves a base64-encoded metadata value by key and decodes it.
func parseMetadataBase64(m map[string]string, k string) ([]byte, error) {
	s, ok := m[k]
	if !ok || s == "" {
		return nil, fmt.Errorf("metadata key not found: %s", k)
	}
	v, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 in metadata key %s: %w", k, err)
	}
	return v, nil
}
