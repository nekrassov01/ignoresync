package operator

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// encryptBody encrypts body with AES-256-GCM in fixed-size chunks.
func encryptBody(body io.ReadCloser, key []byte, baseNonce []byte, aad []byte, chunkSize int) (io.ReadCloser, error) {
	if len(key) != keySize {
		_ = body.Close()
		return nil, fmt.Errorf("invalid data key size: %d", len(key))
	}
	if len(baseNonce) != baseNonceSize {
		_ = body.Close()
		return nil, fmt.Errorf("invalid base nonce size: %d", len(baseNonce))
	}

	aead, err := newGCM(key)
	if err != nil {
		_ = body.Close()
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		var err error
		defer func() {
			if closeErr := body.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("failed to encrypt body: %w", err))
			} else {
				_ = pw.Close()
			}
		}()

		if chunkSize <= 0 {
			err = fmt.Errorf("invalid chunk size: %d", chunkSize)
			return
		}
		buf := make([]byte, chunkSize)
		var idx uint32
		for {
			// Avoid uint32 overflow of chunk index
			if idx == math.MaxUint32 {
				err = fmt.Errorf("too many chunks: index overflow")
				return
			}

			n, readErr := io.ReadFull(body, buf)
			if errors.Is(readErr, io.EOF) {
				break
			}

			isLast := errors.Is(readErr, io.ErrUnexpectedEOF)
			if readErr != nil && !isLast {
				err = fmt.Errorf("failed to read plaintext: %w", readErr)
				return
			}

			plaintext := buf[:n]
			var nonce [chunkNonceSize]byte
			setChunkNonce(&nonce, baseNonce, idx)

			ciphertext := aead.Seal(nil, nonce[:], plaintext, aad)
			if _, err = pw.Write(ciphertext); err != nil {
				err = fmt.Errorf("failed to write ciphertext chunk: %w", err)
				return
			}

			idx++
			if isLast {
				break
			}
		}
	}()

	return pr, nil
}

// decryptBody decrypts the stream produced by encryptBody.
func decryptBody(body io.ReadCloser, key []byte, baseNonce []byte, aad []byte, chunkSize int) (io.ReadCloser, error) {
	if len(key) != keySize {
		_ = body.Close()
		return nil, fmt.Errorf("invalid data key size: %d", len(key))
	}
	if len(baseNonce) != baseNonceSize {
		_ = body.Close()
		return nil, fmt.Errorf("invalid base nonce size: %d", len(baseNonce))
	}

	aead, err := newGCM(key)
	if err != nil {
		_ = body.Close()
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		var err error
		defer func() {
			if closeErr := body.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("failed to decrypt body: %w", err))
			} else {
				_ = pw.Close()
			}
		}()

		if chunkSize <= 0 {
			err = fmt.Errorf("invalid chunk size: %d", chunkSize)
			return
		}
		buf := make([]byte, chunkSize+tagSize)
		var idx uint32
		for {
			// Avoid uint32 overflow of chunk index
			if idx == math.MaxUint32 {
				err = fmt.Errorf("too many chunks: index overflow")
				return
			}

			n, readErr := io.ReadFull(body, buf)
			if errors.Is(readErr, io.EOF) {
				break
			}

			isLast := errors.Is(readErr, io.ErrUnexpectedEOF)
			if readErr != nil && !isLast {
				err = fmt.Errorf("failed to read ciphertext: %w", readErr)
				return
			}
			if isLast && n < tagSize {
				err = fmt.Errorf("ciphertext truncated: %w", readErr)
				return
			}

			ciphertext := buf[:n]
			var nonce [chunkNonceSize]byte
			setChunkNonce(&nonce, baseNonce, idx)

			plaintext, openErr := aead.Open(nil, nonce[:], ciphertext, aad)
			if openErr != nil {
				err = fmt.Errorf("failed to decrypt chunk: integrity check failed: %w", openErr)
				return
			}

			if _, err = pw.Write(plaintext); err != nil {
				err = fmt.Errorf("failed to write plaintext chunk: %w", err)
				return
			}

			idx++
			if isLast {
				break
			}
		}
	}()

	return pr, nil
}

// encryptKey wraps plaintext with KEK using AES-GCM.
func encryptKey(key []byte, plaintext []byte, aad []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("invalid key size: %d", len(key))
	}

	aead, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	wrapped := make([]byte, nonceSize, nonceSize+len(plaintext)+aead.Overhead())
	nonce := wrapped[:nonceSize]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to read random nonce for key wrapping: %w", err)
	}

	wrapped = aead.Seal(wrapped, nonce, plaintext, aad)
	return wrapped, nil
}

// decryptKey unwraps ciphertext with KEK using AES-GCM.
func decryptKey(key []byte, ciphertext []byte, aad []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("invalid key size: %d", len(key))
	}

	aead, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize+tagSize {
		return nil, fmt.Errorf("invalid wrapped key size: %d", len(ciphertext))
	}

	plaintext, err := aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], aad)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap key: %w", err)
	}
	return plaintext, nil
}

// generateKey generates a random key for encryption.
func generateKey() ([]byte, error) {
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// generateNonce generates a random base nonce for chunked AES-GCM.
func generateNonce() ([]byte, error) {
	nonce := make([]byte, baseNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate base nonce: %w", err)
	}
	return nonce, nil
}

// setChunkNonce sets the nonce for a chunk by combining baseNonce and chunk index.
func setChunkNonce(dst *[chunkNonceSize]byte, baseNonce []byte, idx uint32) {
	// If baseNonce is smaller than nonce size, place baseNonce then append chunk idx.
	if baseNonceSize < chunkNonceSize {
		copy(dst[:baseNonceSize], baseNonce)
		binary.BigEndian.PutUint32(dst[baseNonceSize:], idx)
		return
	}

	// If baseNonce size equals nonce size, derive per-chunk nonce by XOR'ing
	// last 4 bytes with the chunk index. Do not mutate input baseNonce.
	if baseNonceSize == chunkNonceSize {
		copy(dst[:], baseNonce)
		off := chunkNonceSize - 4
		last := binary.BigEndian.Uint32(dst[off:])
		last ^= idx
		binary.BigEndian.PutUint32(dst[off:], last)
		return
	}

	// Fallback: if baseNonce is larger than nonce size, truncate.
	copy(dst[:], baseNonce[:chunkNonceSize])
}

// deriveLocalKey derives localKey from masterKey.
func deriveLocalKey(mk []byte) ([]byte, error) {
	if len(mk) == 0 {
		return nil, errors.New("master key is empty")
	}
	return hkdf.Key(sha256.New, mk, nil, infoLocalKey, keySize)
}

// deriveWrapKey derives wrapKey from localKey and cloudKey.
func deriveWrapKey(lk, ck []byte) ([]byte, error) {
	if len(lk) == 0 {
		return nil, errors.New("local key is empty")
	}
	if len(ck) == 0 {
		return nil, errors.New("cloud key is empty")
	}
	return hkdf.Key(sha256.New, lk, ck, infoWrapKey, keySize)
}

// newGCM creates a new AES-GCM cipher with the given key.
func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm: %w", err)
	}
	if aead.NonceSize() != chunkNonceSize {
		return nil, fmt.Errorf("unexpected gcm nonce size: %d", aead.NonceSize())
	}
	return aead, nil
}

// buildAAD constructs a deterministic AAD byte sequence from
// s3 object prefix, baseNonce, scheme version and purpose.
func buildAAD(prefix string, purpose string, baseNonce []byte) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	// Validate input lengths to prevent overflow
	// Avoid int overflow on 32bit
	if n := len(prefix); n == 0 {
		return nil, fmt.Errorf("invalid prefix length: %d", n)
	}
	if uint64(len(prefix)) > math.MaxUint32 {
		return nil, fmt.Errorf("prefix too long: %d", len(prefix))
	}
	if n := len(purpose); n == 0 {
		return nil, fmt.Errorf("invalid purpose length: %d", n)
	}
	if uint64(len(purpose)) > math.MaxUint32 {
		return nil, fmt.Errorf("purpose too long: %d", len(purpose))
	}

	// baseNonce may be omitted for some purposes.
	if len(baseNonce) == 0 && purpose != "repo" {
		return nil, fmt.Errorf("baseNonce is required for purpose: %s", purpose)
	}

	// #nosec G115: Write scheme version first (broadest scope)
	err = binary.Write(&buf, binary.BigEndian, uint32(schemeVersion))
	if err != nil {
		return nil, fmt.Errorf("failed to write scheme version: %w", err)
	}

	// #nosec G115: Write prefix length and prefix
	err = binary.Write(&buf, binary.BigEndian, uint32(len(prefix)))
	if err != nil {
		return nil, fmt.Errorf("failed to write prefix length: %w", err)
	}
	_, err = buf.WriteString(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to write prefix: %w", err)
	}

	// #nosec G115: Write purpose length and purpose
	err = binary.Write(&buf, binary.BigEndian, uint32(len(purpose)))
	if err != nil {
		return nil, fmt.Errorf("failed to write purpose length: %w", err)
	}
	_, err = buf.WriteString(purpose)
	if err != nil {
		return nil, fmt.Errorf("failed to write purpose: %w", err)
	}

	// #nosec G115: Write baseNonce length and baseNonce (may be zero-length)
	err = binary.Write(&buf, binary.BigEndian, uint32(len(baseNonce)))
	if err != nil {
		return nil, fmt.Errorf("failed to write baseNonce length: %w", err)
	}
	_, err = buf.Write(baseNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to write baseNonce: %w", err)
	}

	return buf.Bytes(), nil
}
