package crypto

import (
	"crypto/rand"
	"io"

	"golang.org/x/crypto/argon2"
)

// DeriveKey derives a 32-byte key from password and salt using Argon2id
func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLen,
	)
}

// GenerateSalt generates a random salt of SaltLength bytes
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}
