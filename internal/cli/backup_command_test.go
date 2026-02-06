package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
)

func setupBackupCommandConfig(t *testing.T, cfg *config.Config) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	writeCLIConfig(t, tmp, cfg)
}

func TestBackupCommand_ConfigLoadError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	var exit int
	_, stderr := captureStdoutStderr(t, func() {
		exit = BackupCommand(nil)
	})

	if exit != 1 {
		t.Fatalf("exit = %d", exit)
	}
	if !strings.Contains(stderr, "Error loading config") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestBackupCommand_InvalidConfig(t *testing.T) {
	setupBackupCommandConfig(t, &config.Config{
		BackupDir: "/tmp/backups",
		Files:     []string{},
		Folders:   []string{},
	})
	t.Setenv("DOTKEEPER_PASSWORD", "pw")

	var exit int
	_, stderr := captureStdoutStderr(t, func() {
		exit = BackupCommand(nil)
	})

	if exit != 1 {
		t.Fatalf("exit = %d", exit)
	}
	if !strings.Contains(stderr, "Invalid configuration") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestBackupCommand_GetPasswordError(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(source, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	setupBackupCommandConfig(t, &config.Config{
		BackupDir: filepath.Join(tmp, "backups"),
		Files:     []string{source},
	})
	t.Setenv("DOTKEEPER_PASSWORD", "")

	missingPwFile := filepath.Join(tmp, "missing.pw")
	var exit int
	_, stderr := captureStdoutStderr(t, func() {
		exit = BackupCommand([]string{"--password-file", missingPwFile})
	})

	if exit != 1 {
		t.Fatalf("exit = %d", exit)
	}
	if !strings.Contains(stderr, "Error getting password") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestBackupCommand_BackupFailure(t *testing.T) {
	tmp := t.TempDir()
	setupBackupCommandConfig(t, &config.Config{
		BackupDir: filepath.Join(tmp, "backups"),
		Files:     []string{filepath.Join(tmp, "does-not-exist.txt")},
	})
	t.Setenv("DOTKEEPER_PASSWORD", "pw")
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = BackupCommand(nil)
	})

	if exit != 1 {
		t.Fatalf("exit = %d, stdout=%s stderr=%s", exit, stdout, stderr)
	}
	if !strings.Contains(stdout, "Starting backup") {
		t.Fatalf("stdout = %q", stdout)
	}
	if !strings.Contains(stderr, "Backup failed") {
		t.Fatalf("stderr = %q", stderr)
	}

	store, err := history.NewStore()
	if err != nil {
		t.Fatalf("history.NewStore: %v", err)
	}
	entries, err := store.Read(10)
	if err != nil {
		t.Fatalf("history.Read: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected history entry")
	}
	if entries[0].Operation != "backup" || entries[0].Status != "error" {
		t.Fatalf("unexpected history entry: %#v", entries[0])
	}
}

func TestBackupCommand_Success(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(source, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	backupDir := filepath.Join(tmp, "backups")

	setupBackupCommandConfig(t, &config.Config{
		BackupDir: backupDir,
		Files:     []string{source},
	})
	t.Setenv("DOTKEEPER_PASSWORD", "pw")
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = BackupCommand(nil)
	})

	if exit != 0 {
		t.Fatalf("exit = %d, stdout=%s stderr=%s", exit, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr should be empty on success, got %q", stderr)
	}
	for _, want := range []string{"Starting backup", "Backup completed successfully", "Files backed up:"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q: %s", want, stdout)
		}
	}

	matches, err := filepath.Glob(filepath.Join(backupDir, "*.tar.gz.enc"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 backup file, got %d", len(matches))
	}
	if _, err := os.Stat(matches[0] + ".meta.json"); err != nil {
		t.Fatalf("metadata file missing: %v", err)
	}

	store, err := history.NewStore()
	if err != nil {
		t.Fatalf("history.NewStore: %v", err)
	}
	entries, err := store.Read(10)
	if err != nil {
		t.Fatalf("history.Read: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected history entry")
	}
	if entries[0].Operation != "backup" || entries[0].Status != "success" {
		t.Fatalf("unexpected history entry: %#v", entries[0])
	}
}

func TestBackupCommand_HistoryStoreUnavailable(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(source, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	backupDir := filepath.Join(tmp, "backups")

	setupBackupCommandConfig(t, &config.Config{
		BackupDir: backupDir,
		Files:     []string{source},
	})
	t.Setenv("DOTKEEPER_PASSWORD", "pw")

	badStateHome := filepath.Join(tmp, "state-file")
	if err := os.WriteFile(badStateHome, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_STATE_HOME", badStateHome)

	var exit int
	_, stderr := captureStdoutStderr(t, func() {
		exit = BackupCommand(nil)
	})

	if exit != 0 {
		t.Fatalf("exit = %d", exit)
	}
	if !strings.Contains(stderr, "Warning: failed to log history") {
		t.Fatalf("stderr = %q", stderr)
	}
}
