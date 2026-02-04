package e2e

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/diogo/dotkeeper/internal/backup"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/crypto"
	"github.com/diogo/dotkeeper/internal/restore"
)

// TestEncryptionRoundtrip tests that data can be encrypted and decrypted correctly
func TestEncryptionRoundtrip(t *testing.T) {
	skipIfShort(t)

	passwords := []string{
		"simple",
		"Complex!P@ssw0rd#123",
		"unicode-密码-пароль",
		"spaces in password",
		"very-very-very-long-password-that-exceeds-normal-length-expectations",
	}

	for _, password := range passwords {
		t.Run("password="+password[:min(10, len(password))]+"...", func(t *testing.T) {
			// Generate test data
			plaintext := []byte("This is test data to encrypt")

			// Generate salt
			salt, err := crypto.GenerateSalt()
			if err != nil {
				t.Fatalf("failed to generate salt: %v", err)
			}

			// Derive key
			key := crypto.DeriveKey(password, salt)

			// Encrypt
			ciphertext, err := crypto.Encrypt(plaintext, key, salt)
			if err != nil {
				t.Fatalf("encryption failed: %v", err)
			}

			// Ciphertext should be different from plaintext
			if bytes.Equal(plaintext, ciphertext) {
				t.Error("ciphertext should differ from plaintext")
			}

			// Decrypt
			decrypted, err := crypto.Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decryption failed: %v", err)
			}

			// Verify roundtrip
			if !bytes.Equal(plaintext, decrypted) {
				t.Errorf("decrypted data doesn't match original:\n  expected: %q\n  got: %q",
					plaintext, decrypted)
			}
		})
	}
}

// TestEncryptionWithLargeData tests encryption of large data
func TestEncryptionWithLargeData(t *testing.T) {
	skipIfShort(t)

	sizes := []int{
		1,           // 1 byte
		1024,        // 1 KB
		1024 * 1024, // 1 MB
	}

	for _, size := range sizes {
		t.Run("size="+formatTestSize(size), func(t *testing.T) {
			// Generate random data
			plaintext := make([]byte, size)
			if _, err := rand.Read(plaintext); err != nil {
				t.Fatalf("failed to generate random data: %v", err)
			}

			salt, err := crypto.GenerateSalt()
			if err != nil {
				t.Fatalf("failed to generate salt: %v", err)
			}

			key := crypto.DeriveKey(testPassword, salt)

			// Encrypt
			ciphertext, err := crypto.Encrypt(plaintext, key, salt)
			if err != nil {
				t.Fatalf("encryption failed for %d bytes: %v", size, err)
			}

			// Decrypt
			decrypted, err := crypto.Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decryption failed for %d bytes: %v", size, err)
			}

			// Verify
			if !bytes.Equal(plaintext, decrypted) {
				t.Errorf("data mismatch for %d bytes", size)
			}
		})
	}
}

// TestEncryptionWrongPassword verifies decryption fails with wrong password
func TestEncryptionWrongPassword(t *testing.T) {
	skipIfShort(t)

	plaintext := []byte("secret data")

	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	// Encrypt with correct password
	correctKey := crypto.DeriveKey("correct-password", salt)
	ciphertext, err := crypto.Encrypt(plaintext, correctKey, salt)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Try to decrypt with wrong password
	wrongKey := crypto.DeriveKey("wrong-password", salt)
	_, err = crypto.Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Error("decryption should fail with wrong password")
	}
}

// TestEncryptionDataIntegrity verifies that tampering is detected
func TestEncryptionDataIntegrity(t *testing.T) {
	skipIfShort(t)

	plaintext := []byte("important data that must not be tampered with")

	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	key := crypto.DeriveKey(testPassword, salt)

	ciphertext, err := crypto.Encrypt(plaintext, key, salt)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Tamper with the ciphertext
	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	// Flip a bit somewhere in the encrypted data (not the header)
	if len(tampered) > 30 {
		tampered[30] ^= 0x01
	}

	// Decryption should fail due to authentication
	_, err = crypto.Decrypt(tampered, key)
	if err == nil {
		t.Error("decryption should fail for tampered data")
	}
}

// TestEncryptedBackupIntegrity tests the full encryption path through backup/restore
func TestEncryptedBackupIntegrity(t *testing.T) {
	skipIfShort(t)

	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	backupDir := filepath.Join(tempDir, "backups")
	restoreDir := filepath.Join(tempDir, "restore")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatalf("failed to create restore dir: %v", err)
	}

	// Create test files with binary content
	binaryContent := make([]byte, 1024)
	if _, err := rand.Read(binaryContent); err != nil {
		t.Fatalf("failed to generate binary content: %v", err)
	}

	testFiles := map[string][]byte{
		"text.txt":   []byte("Hello, World!\nLine 2\n"),
		"binary.dat": binaryContent,
		"empty.txt":  {},
	}

	for name, content := range testFiles {
		path := filepath.Join(sourceDir, name)
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("failed to write test file %s: %v", name, err)
		}
	}

	// Create config
	cfg := &config.Config{
		BackupDir: backupDir,
		GitRemote: "https://github.com/test/dotfiles.git",
		Folders:   []string{sourceDir},
	}

	// Backup
	result, err := backup.Backup(cfg, testPassword)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Verify backup file is encrypted (doesn't contain plaintext)
	encryptedData, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}

	for _, content := range testFiles {
		if len(content) > 0 && bytes.Contains(encryptedData, content) {
			t.Error("encrypted backup contains plaintext content")
		}
	}

	// Restore and verify
	opts := restore.RestoreOptions{
		TargetDir: restoreDir,
		Force:     true,
	}

	_, err = restore.Restore(result.BackupPath, testPassword, opts)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	// Verify all files restored correctly
	for name, expectedContent := range testFiles {
		restoredPath := filepath.Join(restoreDir, name)
		content, err := os.ReadFile(restoredPath)
		if err != nil {
			t.Errorf("failed to read restored file %s: %v", name, err)
			continue
		}
		if !bytes.Equal(content, expectedContent) {
			t.Errorf("content mismatch for %s", name)
		}
	}
}

// TestEncryptionWithDifferentSalts tests that same password with different salts produces different keys
func TestEncryptionWithDifferentSalts(t *testing.T) {
	skipIfShort(t)

	password := "same-password"
	plaintext := []byte("test data")

	// Generate two different salts
	salt1, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt1: %v", err)
	}
	salt2, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt2: %v", err)
	}

	// Derive keys
	key1 := crypto.DeriveKey(password, salt1)
	key2 := crypto.DeriveKey(password, salt2)

	// Keys should be different
	if bytes.Equal(key1, key2) {
		t.Error("different salts should produce different keys")
	}

	// Encrypt with key1
	ciphertext1, err := crypto.Encrypt(plaintext, key1, salt1)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Encrypt with key2
	ciphertext2, err := crypto.Encrypt(plaintext, key2, salt2)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Ciphertexts should be different
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("same plaintext with different keys should produce different ciphertexts")
	}

	// Each should decrypt correctly with its own key
	decrypted1, err := crypto.Decrypt(ciphertext1, key1)
	if err != nil {
		t.Fatalf("decryption of ciphertext1 failed: %v", err)
	}
	decrypted2, err := crypto.Decrypt(ciphertext2, key2)
	if err != nil {
		t.Fatalf("decryption of ciphertext2 failed: %v", err)
	}

	if !bytes.Equal(decrypted1, plaintext) || !bytes.Equal(decrypted2, plaintext) {
		t.Error("decrypted data doesn't match original")
	}
}

// TestBackupValidationWithWrongPassword tests that backup validation fails with wrong password
func TestBackupValidationWithWrongPassword(t *testing.T) {
	skipIfShort(t)

	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	backupDir := filepath.Join(tempDir, "backups")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	testFile := filepath.Join(sourceDir, "secret.txt")
	if err := os.WriteFile(testFile, []byte("secret content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := &config.Config{
		BackupDir: backupDir,
		GitRemote: "https://github.com/test/dotfiles.git",
		Files:     []string{testFile},
	}

	// Create backup with correct password
	result, err := backup.Backup(cfg, "correct-password")
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Validation should pass with correct password
	if err := restore.ValidateBackup(result.BackupPath, "correct-password"); err != nil {
		t.Errorf("validation should pass with correct password: %v", err)
	}

	// Validation should fail with wrong password
	err = restore.ValidateBackup(result.BackupPath, "wrong-password")
	if err == nil {
		t.Error("validation should fail with wrong password")
	}
}

// TestEmptyFileEncryption tests encryption of empty files
func TestEmptyFileEncryption(t *testing.T) {
	skipIfShort(t)

	plaintext := []byte{}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	key := crypto.DeriveKey(testPassword, salt)

	// Encrypt empty data
	ciphertext, err := crypto.Encrypt(plaintext, key, salt)
	if err != nil {
		t.Fatalf("encryption of empty data failed: %v", err)
	}

	// Decrypt
	decrypted, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decryption of empty data failed: %v", err)
	}

	// Verify
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted empty data doesn't match:\n  expected len: %d\n  got len: %d",
			len(plaintext), len(decrypted))
	}
}

// formatTestSize formats a byte size for test names
func formatTestSize(size int) string {
	switch {
	case size >= 1024*1024:
		return fmt.Sprintf("%dMB", size/(1024*1024))
	case size >= 1024:
		return fmt.Sprintf("%dKB", size/1024)
	default:
		return fmt.Sprintf("%dB", size)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
