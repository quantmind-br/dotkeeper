package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/backup"
	"github.com/diogo/dotkeeper/internal/config"
)

// createTestBackup creates a test backup for use in restore tests
func createTestBackup(t *testing.T, tmpDir string, files map[string]string) (string, string) {
	t.Helper()

	backupDir := filepath.Join(tmpDir, "backups")
	sourceDir := filepath.Join(tmpDir, "source")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	var filePaths []string
	for name, content := range files {
		filePath := filepath.Join(sourceDir, name)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		filePaths = append(filePaths, filePath)
	}

	cfg := &config.Config{
		BackupDir: backupDir,
		Files:     filePaths,
		Folders:   []string{},
	}

	password := "test-password-123"

	result, err := backup.Backup(cfg, password)
	if err != nil {
		t.Fatalf("Failed to create test backup: %v", err)
	}

	return result.BackupPath, password
}

// setupTestConfig creates a temporary config file and sets XDG_CONFIG_HOME
func setupTestConfig(t *testing.T, tmpDir string) {
	t.Helper()

	configDir := filepath.Join(tmpDir, "config", "dotkeeper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	backupDir := filepath.Join(tmpDir, "backups")
	cfg := &config.Config{
		BackupDir:     backupDir,
		Files:         []string{},
		Folders:       []string{},
		Notifications: false,
	}

	configPath := filepath.Join(configDir, "config.yaml")
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatal(err)
	}

	// Set XDG_CONFIG_HOME to our temp directory
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
}

func TestRestoreCommand_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir)

	// Create a test backup
	files := map[string]string{
		"test1.txt": "content1",
		"test2.txt": "content2",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	// Write password to temp file
	pwFile := filepath.Join(tmpDir, "password")
	if err := os.WriteFile(pwFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}

	// Get just the backup name (filename only)
	backupName := filepath.Base(backupPath)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Call RestoreCommand (backup name first, then flags)
	exitCode := RestoreCommand([]string{
		"--password-file", pwFile,
		"--force",
		backupName,
	})

	// Restore stdout and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify output contains success message
	if !strings.Contains(output, "Restore completed successfully") {
		t.Errorf("Expected success message in output, got: %s", output)
	}

	// Verify files were restored to their original locations
	sourceDir := filepath.Join(tmpDir, "source")
	for name, expectedContent := range files {
		restoredPath := filepath.Join(sourceDir, name)
		content, err := os.ReadFile(restoredPath)
		if err != nil {
			t.Errorf("Failed to read restored file %s: %v", name, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("File %s: expected content %q, got %q", name, expectedContent, string(content))
		}
	}
}

func TestRestoreCommand_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir)

	// Create a test backup
	files := map[string]string{
		"test.txt": "original content",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	// Modify the source file
	sourceDir := filepath.Join(tmpDir, "source")
	testFile := filepath.Join(sourceDir, "test.txt")
	modifiedContent := "modified content"
	if err := os.WriteFile(testFile, []byte(modifiedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write password to temp file
	pwFile := filepath.Join(tmpDir, "password")
	if err := os.WriteFile(pwFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}

	backupName := filepath.Base(backupPath)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	exitCode := RestoreCommand([]string{
		"--password-file", pwFile,
		"--dry-run",
		backupName,
	})

	// Restore stdout and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify output contains dry run message
	if !strings.Contains(output, "Dry run completed") {
		t.Errorf("Expected dry run message in output, got: %s", output)
	}

	// Verify file was NOT restored (still has modified content)
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != modifiedContent {
		t.Errorf("File should not be modified in dry run. Expected %q, got %q", modifiedContent, string(content))
	}
}

func TestRestoreCommand_Diff(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir)

	// Create a test backup
	files := map[string]string{
		"test.txt": "line1\nline2\nline3\n",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	// Modify the source file
	sourceDir := filepath.Join(tmpDir, "source")
	testFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("line1\nmodified\nline3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Write password to temp file
	pwFile := filepath.Join(tmpDir, "password")
	if err := os.WriteFile(pwFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}

	backupName := filepath.Base(backupPath)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	exitCode := RestoreCommand([]string{
		"--password-file", pwFile,
		"--diff",
		"--force",
		backupName,
	})

	// Restore stdout and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify output contains diff markers
	if !strings.Contains(output, "---") {
		t.Errorf("Expected diff to contain '---' marker, got: %s", output)
	}
	if !strings.Contains(output, "+++") {
		t.Errorf("Expected diff to contain '+++' marker, got: %s", output)
	}
	if !strings.Contains(output, "@@") {
		t.Errorf("Expected diff to contain '@@' marker, got: %s", output)
	}
}

func TestRestoreCommand_DryRunDiff(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir)

	// Create a test backup
	files := map[string]string{
		"test.txt": "line1\nline2\nline3\n",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	// Modify the source file
	sourceDir := filepath.Join(tmpDir, "source")
	testFile := filepath.Join(sourceDir, "test.txt")
	modifiedContent := "line1\nmodified line\nline3\n"
	if err := os.WriteFile(testFile, []byte(modifiedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write password to temp file
	pwFile := filepath.Join(tmpDir, "password")
	if err := os.WriteFile(pwFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}

	backupName := filepath.Base(backupPath)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	exitCode := RestoreCommand([]string{
		"--password-file", pwFile,
		"--dry-run",
		"--diff",
		backupName,
	})

	// Restore stdout and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify output contains both dry run and diff
	if !strings.Contains(output, "Dry run completed") {
		t.Errorf("Expected dry run message in output, got: %s", output)
	}
	if !strings.Contains(output, "---") {
		t.Errorf("Expected diff markers in output, got: %s", output)
	}

	// Verify file was NOT restored
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != modifiedContent {
		t.Errorf("File should not be modified in dry run. Expected %q, got %q", modifiedContent, string(content))
	}
}

func TestRestoreCommand_Force(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir)

	// Create a test backup
	files := map[string]string{
		"test.txt": "backup content",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	// Create a conflicting file
	sourceDir := filepath.Join(tmpDir, "source")
	testFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("existing content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Write password to temp file
	pwFile := filepath.Join(tmpDir, "password")
	if err := os.WriteFile(pwFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}

	backupName := filepath.Base(backupPath)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	exitCode := RestoreCommand([]string{
		"--password-file", pwFile,
		"--force",
		backupName,
	})

	// Restore stdout and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify file was overwritten with backup content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "backup content" {
		t.Errorf("File should be overwritten with backup content. Expected %q, got %q", "backup content", string(content))
	}
}

func TestRestoreCommand_WrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir)

	// Create a test backup
	files := map[string]string{
		"test.txt": "content",
	}
	backupPath, _ := createTestBackup(t, tmpDir, files)

	// Write WRONG password to temp file
	pwFile := filepath.Join(tmpDir, "password")
	if err := os.WriteFile(pwFile, []byte("wrong-password"), 0600); err != nil {
		t.Fatal(err)
	}

	backupName := filepath.Base(backupPath)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	exitCode := RestoreCommand([]string{
		"--password-file", pwFile,
		backupName,
	})

	// Restore stderr and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for wrong password, got %d", exitCode)
	}

	// Verify error message
	if !strings.Contains(output, "Restore failed") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}

func TestRestoreCommand_MissingBackup(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir)

	// Write password to temp file
	pwFile := filepath.Join(tmpDir, "password")
	if err := os.WriteFile(pwFile, []byte("test-password"), 0600); err != nil {
		t.Fatal(err)
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	exitCode := RestoreCommand([]string{
		"--password-file", pwFile,
		"nonexistent-backup.tar.gz.enc",
	})

	// Restore stderr and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing backup, got %d", exitCode)
	}

	// Verify error message
	if !strings.Contains(output, "backup not found") {
		t.Errorf("Expected 'backup not found' in error message, got: %s", output)
	}
}

func TestRestoreCommand_MissingBackupName(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	// Call RestoreCommand without backup name
	exitCode := RestoreCommand([]string{})

	// Restore stderr and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing backup name, got %d", exitCode)
	}

	// Verify error message
	if !strings.Contains(output, "backup name required") {
		t.Errorf("Expected 'backup name required' in error message, got: %s", output)
	}
}
