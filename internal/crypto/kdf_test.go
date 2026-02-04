package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	tests := []struct {
		name     string
		password string
		salt     []byte
	}{
		{
			name:     "basic derivation",
			password: "test-password",
			salt:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			name:     "empty password",
			password: "",
			salt:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			name:     "long password",
			password: "this-is-a-very-long-password-with-many-characters-for-testing",
			salt:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := DeriveKey(tt.password, tt.salt)

			// Verify key length is correct
			if len(key) != Argon2KeyLen {
				t.Errorf("key length: got %d, want %d", len(key), Argon2KeyLen)
			}

			// Verify key is not nil
			if key == nil {
				t.Error("key should not be nil")
			}

			// Verify key is not all zeros
			allZeros := true
			for _, b := range key {
				if b != 0 {
					allZeros = false
					break
				}
			}
			if allZeros {
				t.Error("key should not be all zeros")
			}
		})
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	password := "test-password"
	salt := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	key1 := DeriveKey(password, salt)
	key2 := DeriveKey(password, salt)

	// Same password and salt should produce same key
	if !bytes.Equal(key1, key2) {
		t.Error("same password and salt should produce same key")
	}
}

func TestDeriveKeyDifferentSalts(t *testing.T) {
	password := "test-password"
	salt1 := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	salt2 := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	key1 := DeriveKey(password, salt1)
	key2 := DeriveKey(password, salt2)

	// Different salts should produce different keys
	if bytes.Equal(key1, key2) {
		t.Error("different salts should produce different keys")
	}
}

func TestDeriveKeyDifferentPasswords(t *testing.T) {
	salt := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	password1 := "password1"
	password2 := "password2"

	key1 := DeriveKey(password1, salt)
	key2 := DeriveKey(password2, salt)

	// Different passwords should produce different keys
	if bytes.Equal(key1, key2) {
		t.Error("different passwords should produce different keys")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt()

	// Verify no error
	if err != nil {
		t.Errorf("GenerateSalt should not return error: %v", err)
	}

	// Verify salt length is correct
	if len(salt) != SaltLength {
		t.Errorf("salt length: got %d, want %d", len(salt), SaltLength)
	}

	// Verify salt is not nil
	if salt == nil {
		t.Error("salt should not be nil")
	}

	// Verify salt is not all zeros
	allZeros := true
	for _, b := range salt {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		t.Error("salt should not be all zeros")
	}
}

func TestGenerateSaltRandomness(t *testing.T) {
	salt1, err1 := GenerateSalt()
	salt2, err2 := GenerateSalt()

	if err1 != nil || err2 != nil {
		t.Fatalf("GenerateSalt should not return error")
	}

	// Different calls should produce different salts
	if bytes.Equal(salt1, salt2) {
		t.Error("GenerateSalt should produce different salts on each call")
	}
}

func TestGenerateSaltMultiple(t *testing.T) {
	salts := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		salt, err := GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt failed: %v", err)
		}
		salts[i] = salt
	}

	// Verify all salts are unique
	for i := 0; i < len(salts); i++ {
		for j := i + 1; j < len(salts); j++ {
			if bytes.Equal(salts[i], salts[j]) {
				t.Errorf("salt %d and %d should be different", i, j)
			}
		}
	}
}
