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

// TestEncrypt_InvalidKeyLength tests encryption with wrong key size
func TestEncrypt_InvalidKeyLength(t *testing.T) {
	plaintext := []byte("test")
	invalidKey := []byte("short") // Wrong key size

	_, err := Encrypt(plaintext, invalidKey, make([]byte, SaltLength))
	if err == nil {
		t.Error("Encrypt with invalid key length should fail")
	}
}

// TestDecrypt_InvalidKeyLength tests decryption with wrong key size
func TestDecrypt_InvalidKeyLength(t *testing.T) {
	invalidKey := []byte("short") // Wrong key size
	ciphertext := make([]byte, 100) // Any ciphertext

	_, err := Decrypt(ciphertext, invalidKey)
	if err == nil {
		t.Error("Decrypt with invalid key length should fail")
	}
}

// TestDecrypt_CiphertextTooShort tests various too-short ciphertexts
func TestDecrypt_CiphertextTooShort(t *testing.T) {
	key := make([]byte, AESKeySize)

	tests := []struct {
		name        string
		ciphertext  []byte
	}{
		{"empty", []byte{}},
		{"only version", []byte{1}},
		{"version + partial salt", []byte{1, 2, 3}},
		{"missing nonce", make([]byte, 1+SaltLength)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.ciphertext, key)
			if err == nil {
				t.Error("Decrypt should fail with too-short ciphertext")
			}
		})
	}
}

// TestDecrypt_CorruptedData tests that decryption fails with corrupted data
func TestDecrypt_CorruptedData(t *testing.T) {
	plaintext := []byte("Secret data")
	password := "password"

	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	key := DeriveKey(password, salt)
	ciphertext, err := Encrypt(plaintext, key, salt)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Corrupt the ciphertext by changing some bytes
	// Format: [version(1)][salt(16)][nonce(12)][ciphertext+tag...]
	// We need to corrupt bytes after the nonce (after position 1+16+12=29)
	dataStart := 1 + SaltLength + AESNonceSize
	if len(ciphertext) > dataStart+2 {
		ciphertext[dataStart] ^= 0xFF   // Corrupt first byte of encrypted data
		ciphertext[dataStart+1] ^= 0xAA // Corrupt second byte
	}

	_, err = Decrypt(ciphertext, key)
	if err == nil {
		t.Error("Decrypt with corrupted ciphertext should fail")
	}
}

// TestEncrypt_EmptyPlaintext tests encryption of empty data
func TestEncrypt_EmptyPlaintext(t *testing.T) {
	plaintext := []byte{}
	password := "password"

	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	key := DeriveKey(password, salt)
	ciphertext, err := Encrypt(plaintext, key, salt)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Should be able to decrypt empty data
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty decrypted data, got %d bytes", len(decrypted))
	}
}

// TestEncrypt_LargePlaintext tests encryption of large data
func TestEncrypt_LargePlaintext(t *testing.T) {
	password := "password"
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	key := DeriveKey(password, salt)

	// Create a large plaintext (10KB)
	largePlaintext := make([]byte, 10*1024)
	for i := range largePlaintext {
		largePlaintext[i] = byte(i % 256)
	}

	ciphertext, err := Encrypt(largePlaintext, key, salt)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Verify ciphertext is not empty
	if len(ciphertext) == 0 {
		t.Fatal("ciphertext is empty")
	}

	// Decrypt and verify
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if string(decrypted) != string(largePlaintext) {
		t.Fatal("decrypted text does not match original plaintext")
	}
}

// TestEncrypt_SingleByte tests encryption of single byte
func TestEncrypt_SingleByte(t *testing.T) {
	plaintext := []byte("X")
	password := "password"

	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	key := DeriveKey(password, salt)
	ciphertext, err := Encrypt(plaintext, key, salt)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("decrypted text does not match: got %q, want %q", string(decrypted), string(plaintext))
	}
}

// TestGenerateSalt_MultipleCalls tests that GenerateSalt produces different salts
func TestGenerateSalt_MultipleCalls(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}
	if len(salt1) != SaltLength {
		t.Errorf("salt length: got %d, want %d", len(salt1), SaltLength)
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}
	if len(salt2) != SaltLength {
		t.Errorf("salt length: got %d, want %d", len(salt2), SaltLength)
	}

	// Salts should be different (statistically, unless very unlucky)
	if string(salt1) == string(salt2) {
		t.Error("two calls to GenerateSalt should produce different salts")
	}
}
