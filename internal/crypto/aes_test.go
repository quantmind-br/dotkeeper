package crypto

import (
	"testing"
)

// TestEncryptDecrypt tests basic encryption and decryption
func TestEncryptDecrypt(t *testing.T) {
	plaintext := []byte("Hello, World! This is a secret message.")
	password := "my-secure-password"

	// Generate salt
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	// Derive key from password
	key := DeriveKey(password, salt)

	// Encrypt
	ciphertext, err := Encrypt(plaintext, key, salt)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Verify ciphertext is not empty and has expected structure
	if len(ciphertext) == 0 {
		t.Fatal("ciphertext is empty")
	}

	// Verify format: [version(1)][salt(16)][nonce(12)][ciphertext...]
	expectedMinLen := 1 + SaltLength + AESNonceSize
	if len(ciphertext) < expectedMinLen {
		t.Fatalf("ciphertext too short: got %d, want at least %d", len(ciphertext), expectedMinLen)
	}

	// Verify version byte
	if ciphertext[0] != byte(Version) {
		t.Fatalf("version mismatch: got %d, want %d", ciphertext[0], Version)
	}

	// Verify salt is embedded correctly
	embeddedSalt := ciphertext[1 : 1+SaltLength]
	if string(embeddedSalt) != string(salt) {
		t.Fatal("embedded salt does not match original salt")
	}

	// Decrypt
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// Verify plaintext matches
	if string(decrypted) != string(plaintext) {
		t.Fatalf("decrypted text does not match: got %q, want %q", string(decrypted), string(plaintext))
	}
}

// TestWrongPassword tests that decryption fails with wrong password
func TestWrongPassword(t *testing.T) {
	plaintext := []byte("Secret data")
	password := "correct-password"
	wrongPassword := "wrong-password"

	// Generate salt
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	// Encrypt with correct password
	key := DeriveKey(password, salt)
	ciphertext, err := Encrypt(plaintext, key, salt)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Try to decrypt with wrong password
	wrongKey := DeriveKey(wrongPassword, salt)
	_, err = Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Fatal("decryption with wrong password should fail")
	}
}

// TestEncryptDifferentNonces tests that encrypting same plaintext produces different ciphertexts
func TestEncryptDifferentNonces(t *testing.T) {
	plaintext := []byte("Same message")
	password := "password"

	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	key := DeriveKey(password, salt)

	// Encrypt same plaintext twice
	ciphertext1, err := Encrypt(plaintext, key, salt)
	if err != nil {
		t.Fatalf("first encryption failed: %v", err)
	}

	ciphertext2, err := Encrypt(plaintext, key, salt)
	if err != nil {
		t.Fatalf("second encryption failed: %v", err)
	}

	// Ciphertexts should be different (due to random nonce)
	if string(ciphertext1) == string(ciphertext2) {
		t.Fatal("two encryptions of same plaintext should produce different ciphertexts")
	}

	// But both should decrypt to same plaintext
	decrypted1, err := Decrypt(ciphertext1, key)
	if err != nil {
		t.Fatalf("first decryption failed: %v", err)
	}

	decrypted2, err := Decrypt(ciphertext2, key)
	if err != nil {
		t.Fatalf("second decryption failed: %v", err)
	}

	if string(decrypted1) != string(plaintext) || string(decrypted2) != string(plaintext) {
		t.Fatal("both decryptions should match original plaintext")
	}
}

// TestEmptyCiphertext tests that decryption fails with too-short ciphertext
func TestEmptyCiphertext(t *testing.T) {
	key := make([]byte, AESKeySize)
	_, err := Decrypt([]byte{}, key)
	if err == nil {
		t.Fatal("decryption of empty ciphertext should fail")
	}
}
