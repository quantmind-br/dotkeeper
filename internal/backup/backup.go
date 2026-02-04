package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/crypto"
)

// BackupResult contains information about a completed backup
type BackupResult struct {
	BackupPath   string
	MetadataPath string
	BackupName   string
	FileCount    int
	TotalSize    int64
	Duration     time.Duration
	Checksum     string
}

// Backup performs a full backup: collect files → create archive → encrypt → save
func Backup(cfg *config.Config, password string) (*BackupResult, error) {
	start := time.Now()

	// Generate backup name with timestamp
	backupName := fmt.Sprintf("backup-%s.tar.gz.enc", time.Now().Format("2006-01-02-150405"))
	backupPath := filepath.Join(cfg.BackupDir, backupName)
	metadataPath := backupPath + ".meta.json"

	// Ensure backup directory exists
	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Collect all files
	files, err := CollectFiles(append(cfg.Files, cfg.Folders...))
	if err != nil {
		return nil, fmt.Errorf("failed to collect files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files to backup")
	}

	// Create temp file for archive
	tempFile, err := os.CreateTemp("", "dotkeeper-backup-*.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Create archive
	if err := CreateArchive(files, tempFile); err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to create archive: %w", err)
	}
	tempFile.Close()

	// Read archive for encryption
	archiveData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read archive: %w", err)
	}

	// Calculate checksum of plaintext
	checksum := sha256.Sum256(archiveData)
	checksumHex := hex.EncodeToString(checksum[:])

	// Generate salt and derive key
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := crypto.DeriveKey(password, salt)

	// Encrypt
	encrypted, err := crypto.Encrypt(archiveData, key, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	// Write encrypted backup
	if err := os.WriteFile(backupPath, encrypted, 0600); err != nil {
		return nil, fmt.Errorf("failed to write backup: %w", err)
	}

	// Create metadata
	metadata := crypto.EncryptionMetadata{
		Version:      crypto.Version,
		Algorithm:    "AES-256-GCM",
		KDF:          "Argon2id",
		Salt:         salt,
		KDFTime:      crypto.Argon2Time,
		KDFMemory:    crypto.Argon2Memory,
		KDFThreads:   crypto.Argon2Threads,
		Timestamp:    time.Now(),
		OriginalSize: int64(len(archiveData)),
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	// Calculate total size
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	return &BackupResult{
		BackupPath:   backupPath,
		MetadataPath: metadataPath,
		BackupName:   backupName,
		FileCount:    len(files),
		TotalSize:    totalSize,
		Duration:     time.Since(start),
		Checksum:     checksumHex,
	}, nil
}
