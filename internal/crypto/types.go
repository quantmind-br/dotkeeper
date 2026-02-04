package crypto

import "time"

// Constants for Argon2id KDF
const (
	Argon2Time    = 3
	Argon2Memory  = 64 * 1024 // 64 MB
	Argon2Threads = 4
	Argon2KeyLen  = 32 // 256-bit key for AES-256
	SaltLength    = 16 // 128-bit salt
)

// Constants for AES-GCM
const (
	AESKeySize   = 32 // 256 bits
	AESNonceSize = 12 // GCM standard nonce size
	Version      = 1  // Encryption format version
)

// EncryptionMetadata stores metadata about encrypted data
type EncryptionMetadata struct {
	Version      int       `json:"version"`
	Algorithm    string    `json:"algorithm"`
	KDF          string    `json:"kdf"`
	Salt         []byte    `json:"salt"`
	KDFTime      int       `json:"kdf_time"`
	KDFMemory    int       `json:"kdf_memory"`
	KDFThreads   int       `json:"kdf_threads"`
	Timestamp    time.Time `json:"timestamp"`
	OriginalSize int64     `json:"original_size"`
}

// DefaultMetadata returns a new metadata with default values
func DefaultMetadata() EncryptionMetadata {
	return EncryptionMetadata{
		Version:    Version,
		Algorithm:  "AES-256-GCM",
		KDF:        "Argon2id",
		KDFTime:    Argon2Time,
		KDFMemory:  Argon2Memory,
		KDFThreads: Argon2Threads,
		Timestamp:  time.Now(),
	}
}
