package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
)

// TestCLIBackupWorkflow tests the complete CLI backup workflow
func TestCLIBackupWorkflow(t *testing.T) {
	skipIfShort(t)

	// Skip if binary not available
	binaryPath := getBinaryPath(t)
	if binaryPath == "" {
		t.Skip("dotkeeper binary not found, skipping CLI tests")
	}

	// Create test environment
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	backupDir := filepath.Join(tempDir, "backups")
	configDir := filepath.Join(tempDir, "config")
	configPath := filepath.Join(configDir, "config.yaml")

	// Create directories
	for _, dir := range []string{sourceDir, backupDir, configDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Create test files
	testFile := filepath.Join(sourceDir, "test.conf")
	if err := os.WriteFile(testFile, []byte("test=value"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create config
	cfg := &config.Config{
		BackupDir: backupDir,
		GitRemote: "https://github.com/test/dotfiles.git",
		Files:     []string{testFile},
	}
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Set up environment for CLI
	env := append(os.Environ(),
		"XDG_CONFIG_HOME="+tempDir,
		"DOTKEEPER_PASSWORD="+testPassword,
	)

	// Test 1: Run backup command
	t.Log("Test 1: Running backup via CLI...")
	cmd := exec.Command(binaryPath, "backup")
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Create dotkeeper config dir at expected location
	dotKeeperConfigDir := filepath.Join(tempDir, "dotkeeper")
	if err := os.MkdirAll(dotKeeperConfigDir, 0755); err != nil {
		t.Fatalf("failed to create dotkeeper config dir: %v", err)
	}
	dotKeeperConfigPath := filepath.Join(dotKeeperConfigDir, "config.yaml")
	if err := cfg.SaveToPath(dotKeeperConfigPath); err != nil {
		t.Fatalf("failed to save dotkeeper config: %v", err)
	}

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("backup command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Backup completed") {
		t.Errorf("expected 'Backup completed' in output, got: %s", output)
	}

	// Verify backup was created
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("failed to read backup dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("no backup files created")
	}

	t.Logf("Backup created: %d files in backup dir", len(entries))

	// Test 2: Run list command
	t.Log("Test 2: Running list via CLI...")
	cmd = exec.Command(binaryPath, "list")
	cmd.Env = env
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("list command failed: %v", err)
	}

	output = stdout.String()
	if !strings.Contains(output, "backup-") || !strings.Contains(output, ".tar.gz.enc") {
		t.Errorf("expected backup listing in output, got: %s", output)
	}

	// Test 3: Run list --json
	t.Log("Test 3: Running list --json via CLI...")
	cmd = exec.Command(binaryPath, "list", "--json")
	cmd.Env = env
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("list --json command failed: %v", err)
	}

	output = stdout.String()
	if !strings.Contains(output, "[") || !strings.Contains(output, "name") {
		t.Errorf("expected JSON output, got: %s", output)
	}
}

// TestCLIConfigWorkflow tests the config CLI commands
func TestCLIConfigWorkflow(t *testing.T) {
	skipIfShort(t)

	binaryPath := getBinaryPath(t)
	if binaryPath == "" {
		t.Skip("dotkeeper binary not found, skipping CLI tests")
	}

	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "dotkeeper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Create initial config
	cfg := &config.Config{
		BackupDir:     "/tmp/backups",
		GitRemote:     "https://github.com/test/dotfiles.git",
		Schedule:      "0 2 * * *",
		Notifications: true,
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	env := append(os.Environ(), "XDG_CONFIG_HOME="+tempDir)

	// Test config list
	t.Log("Test: config list...")
	cmd := exec.Command(binaryPath, "config", "list")
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("config list failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "backup_dir") {
		t.Errorf("expected 'backup_dir' in config list output, got: %s", output)
	}

	// Test config get
	t.Log("Test: config get backup_dir...")
	cmd = exec.Command(binaryPath, "config", "get", "backup_dir")
	cmd.Env = env
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("config get failed: %v", err)
	}

	output = strings.TrimSpace(stdout.String())
	if output != "/tmp/backups" {
		t.Errorf("expected '/tmp/backups', got: %s", output)
	}

	// Test config set
	t.Log("Test: config set backup_dir...")
	cmd = exec.Command(binaryPath, "config", "set", "backup_dir", "/new/backup/dir")
	cmd.Env = env
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("config set failed: %v", err)
	}

	// Verify the value was set
	cmd = exec.Command(binaryPath, "config", "get", "backup_dir")
	cmd.Env = env
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("config get (verification) failed: %v", err)
	}

	output = strings.TrimSpace(stdout.String())
	if output != "/new/backup/dir" {
		t.Errorf("expected '/new/backup/dir', got: %s", output)
	}
}

// TestCLIRestoreWorkflow tests the restore CLI command
func TestCLIRestoreWorkflow(t *testing.T) {
	skipIfShort(t)

	binaryPath := getBinaryPath(t)
	if binaryPath == "" {
		t.Skip("dotkeeper binary not found, skipping CLI tests")
	}

	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	backupDir := filepath.Join(tempDir, "backups")
	restoreDir := filepath.Join(tempDir, "restore")
	configDir := filepath.Join(tempDir, "dotkeeper")

	// Create directories
	for _, dir := range []string{sourceDir, backupDir, restoreDir, configDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Create test file
	testFile := filepath.Join(sourceDir, "important.conf")
	testContent := "important=settings\nkey=value"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create config
	cfg := &config.Config{
		BackupDir: backupDir,
		GitRemote: "https://github.com/test/dotfiles.git",
		Files:     []string{testFile},
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	env := append(os.Environ(),
		"XDG_CONFIG_HOME="+tempDir,
		"DOTKEEPER_PASSWORD="+testPassword,
	)

	// Create backup first
	t.Log("Creating backup via CLI...")
	cmd := exec.Command(binaryPath, "backup")
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("backup command failed: %v", err)
	}

	// Get backup name
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("failed to read backup dir: %v", err)
	}
	var backupName string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar.gz.enc") {
			backupName = entry.Name()
			break
		}
	}
	if backupName == "" {
		t.Fatal("no backup file found")
	}

	// Delete original file
	os.Remove(testFile)

	// Run restore command
	t.Log("Running restore via CLI...")
	cmd = exec.Command(binaryPath, "restore", backupName, "--force")
	cmd.Env = env
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("restore command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Restore completed") {
		t.Errorf("expected 'Restore completed' in output, got: %s", output)
	}

	// Verify file was restored
	restoredContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(restoredContent) != testContent {
		t.Errorf("restored content mismatch:\n  expected: %q\n  got: %q",
			testContent, string(restoredContent))
	}

	t.Log("Restore workflow completed successfully")
}

// TestCLIHelpAndVersion tests help and version flags
func TestCLIHelpAndVersion(t *testing.T) {
	skipIfShort(t)

	binaryPath := getBinaryPath(t)
	if binaryPath == "" {
		t.Skip("dotkeeper binary not found, skipping CLI tests")
	}

	// Test --help
	t.Log("Test: --help flag...")
	cmd := exec.Command(binaryPath, "--help")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("--help failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "dotkeeper") || !strings.Contains(output, "backup") {
		t.Errorf("expected help text with 'dotkeeper' and 'backup', got: %s", output)
	}

	// Test --version
	t.Log("Test: --version flag...")
	cmd = exec.Command(binaryPath, "--version")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("--version failed: %v", err)
	}

	output = stdout.String()
	if !strings.Contains(output, "dotkeeper") || !strings.Contains(output, "version") {
		t.Errorf("expected version info, got: %s", output)
	}
}

// TestCLIErrorHandling tests CLI error scenarios
func TestCLIErrorHandling(t *testing.T) {
	skipIfShort(t)

	binaryPath := getBinaryPath(t)
	if binaryPath == "" {
		t.Skip("dotkeeper binary not found, skipping CLI tests")
	}

	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "dotkeeper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	env := append(os.Environ(), "XDG_CONFIG_HOME="+tempDir)

	// Test unknown command
	t.Log("Test: unknown command...")
	cmd := exec.Command(binaryPath, "unknown-command")
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Error("expected error for unknown command")
	}
	if !strings.Contains(stderr.String(), "Unknown command") {
		t.Errorf("expected 'Unknown command' in stderr, got: %s", stderr.String())
	}

	// Test restore without backup name
	t.Log("Test: restore without backup name...")
	// First create a config so restore can load it
	cfg := &config.Config{
		BackupDir: tempDir,
		GitRemote: "https://github.com/test/dotfiles.git",
		Files:     []string{"/tmp/test.txt"},
	}
	if err := cfg.SaveToPath(filepath.Join(configDir, "config.yaml")); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	cmd = exec.Command(binaryPath, "restore")
	cmd.Env = env
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		t.Error("expected error for restore without backup name")
	}
	if !strings.Contains(stderr.String(), "backup name required") {
		t.Errorf("expected 'backup name required' in stderr, got: %s", stderr.String())
	}
}

// getBinaryPath returns the path to the dotkeeper binary
// Returns empty string if not found
func getBinaryPath(t *testing.T) string {
	t.Helper()

	// Check if binary exists in project root
	projectRoot := getProjectRoot()
	binaryPath := filepath.Join(projectRoot, "dotkeeper")

	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath
	}

	// Check if binary exists in bin directory
	binaryPath = filepath.Join(projectRoot, "bin", "dotkeeper")
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath
	}

	// Check if binary is in PATH
	if path, err := exec.LookPath("dotkeeper"); err == nil {
		return path
	}

	return ""
}

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Get the directory of this test file and go up to project root
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// If we're in e2e directory, go up one level
	if filepath.Base(wd) == "e2e" {
		return filepath.Dir(wd)
	}

	return wd
}
