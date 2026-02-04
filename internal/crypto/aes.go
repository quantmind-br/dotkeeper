package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Encrypt encrypts plaintext using AES-256-GCM with the given key
// Returns: [version(1)][salt(16)][nonce(12)][ciphertext...]
func Encrypt(plaintext []byte, key []byte, salt []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Format: [version(1)][salt(16)][nonce(12)][ciphertext...]
	versionByte := []byte{byte(Version)}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	result := make([]byte, 0, 1+len(salt)+len(nonce)+len(ciphertext))
	result = append(result, versionByte...)
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the given key
// Expects format: [version(1)][salt(16)][nonce(12)][ciphertext...]
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(ciphertext) < 1+SaltLength+AESNonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Skip version byte (we could verify it here)
	salt := ciphertext[1 : 1+SaltLength]
	nonce := ciphertext[1+SaltLength : 1+SaltLength+AESNonceSize]
	encryptedData := ciphertext[1+SaltLength+AESNonceSize:]

	_ = salt // salt is for key derivation, already done by caller

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password or corrupted data): %w", err)
	}

	return plaintext, nil
}
