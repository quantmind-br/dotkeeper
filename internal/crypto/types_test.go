package crypto

import (
	"testing"
	"time"
)

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"Argon2Time", Argon2Time, 3},
		{"Argon2Memory", Argon2Memory, 64 * 1024},
		{"Argon2Threads", Argon2Threads, 4},
		{"Argon2KeyLen", Argon2KeyLen, 32},
		{"SaltLength", SaltLength, 16},
		{"AESKeySize", AESKeySize, 32},
		{"AESNonceSize", AESNonceSize, 12},
		{"Version", Version, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("got %d, want %d", tt.value, tt.expected)
			}
		})
	}
}

func TestDefaultMetadata(t *testing.T) {
	metadata := DefaultMetadata()

	if metadata.Version != Version {
		t.Errorf("Version: got %d, want %d", metadata.Version, Version)
	}

	if metadata.Algorithm != "AES-256-GCM" {
		t.Errorf("Algorithm: got %s, want AES-256-GCM", metadata.Algorithm)
	}

	if metadata.KDF != "Argon2id" {
		t.Errorf("KDF: got %s, want Argon2id", metadata.KDF)
	}

	if metadata.KDFTime != Argon2Time {
		t.Errorf("KDFTime: got %d, want %d", metadata.KDFTime, Argon2Time)
	}

	if metadata.KDFMemory != Argon2Memory {
		t.Errorf("KDFMemory: got %d, want %d", metadata.KDFMemory, Argon2Memory)
	}

	if metadata.KDFThreads != Argon2Threads {
		t.Errorf("KDFThreads: got %d, want %d", metadata.KDFThreads, Argon2Threads)
	}

	if metadata.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	// Verify timestamp is recent (within last second)
	if time.Since(metadata.Timestamp) > time.Second {
		t.Error("Timestamp should be recent")
	}

	if metadata.Salt != nil {
		t.Error("Salt should be nil in default metadata")
	}

	if metadata.OriginalSize != 0 {
		t.Error("OriginalSize should be 0 in default metadata")
	}
}

func TestEncryptionMetadataStruct(t *testing.T) {
	metadata := EncryptionMetadata{
		Version:      1,
		Algorithm:    "AES-256-GCM",
		KDF:          "Argon2id",
		Salt:         []byte{1, 2, 3, 4},
		KDFTime:      3,
		KDFMemory:    65536,
		KDFThreads:   4,
		Timestamp:    time.Now(),
		OriginalSize: 1024,
	}

	if metadata.Version != 1 {
		t.Error("Version field not set correctly")
	}

	if len(metadata.Salt) != 4 {
		t.Error("Salt field not set correctly")
	}

	if metadata.OriginalSize != 1024 {
		t.Error("OriginalSize field not set correctly")
	}
}
