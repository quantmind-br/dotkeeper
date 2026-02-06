package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/crypto"
)

func TestFindBackups(t *testing.T) {
	t.Run("directory not found", func(t *testing.T) {
		_, err := findBackups(filepath.Join(t.TempDir(), "missing"))
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("filters and reads metadata", func(t *testing.T) {
		dir := t.TempDir()

		encA := filepath.Join(dir, "backup-a.tar.gz.enc")
		encB := filepath.Join(dir, "backup-b.tar.gz.enc")
		nonBackup := filepath.Join(dir, "note.txt")

		if err := os.WriteFile(encA, []byte("abc"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(encB, []byte("abcdef"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(nonBackup, []byte("ignore"), 0600); err != nil {
			t.Fatal(err)
		}

		meta := crypto.DefaultMetadata()
		meta.Timestamp = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		meta.OriginalSize = 999
		metaJSON, _ := json.Marshal(meta)
		if err := os.WriteFile(encA+".meta.json", metaJSON, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(encB+".meta.json", []byte("not-json"), 0600); err != nil {
			t.Fatal(err)
		}

		backups, err := findBackups(dir)
		if err != nil {
			t.Fatalf("findBackups error: %v", err)
		}
		if len(backups) != 2 {
			t.Fatalf("len(backups) = %d", len(backups))
		}

		var foundA bool
		for _, b := range backups {
			if b.Name == "backup-a.tar.gz.enc" {
				foundA = true
				if b.OriginalSize != 999 {
					t.Fatalf("OriginalSize = %d", b.OriginalSize)
				}
				if !b.Created.Equal(meta.Timestamp) {
					t.Fatalf("Created = %v, want %v", b.Created, meta.Timestamp)
				}
			}
		}
		if !foundA {
			t.Fatal("backup-a.tar.gz.enc not found")
		}
	})
}

func TestListCommand(t *testing.T) {
	setup := func(t *testing.T, backupDir string) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		writeCLIConfig(t, tmp, &config.Config{BackupDir: backupDir, Files: []string{".zshrc"}})
	}

	t.Run("no backup dir", func(t *testing.T) {
		backupDir := filepath.Join(t.TempDir(), "no-dir")
		setup(t, backupDir)

		var exit int
		stdout, stderr := captureStdoutStderr(t, func() {
			exit = ListCommand(nil)
		})
		if exit != 0 {
			t.Fatalf("exit = %d, stderr=%s", exit, stderr)
		}
		if !strings.Contains(stdout, "backup directory does not exist") {
			t.Fatalf("stdout = %q", stdout)
		}
	})

	t.Run("empty dir json", func(t *testing.T) {
		backupDir := t.TempDir()
		setup(t, backupDir)

		var exit int
		stdout, stderr := captureStdoutStderr(t, func() {
			exit = ListCommand([]string{"--json"})
		})
		if exit != 0 {
			t.Fatalf("exit = %d, stderr=%s", exit, stderr)
		}
		if strings.TrimSpace(stdout) != "[]" {
			t.Fatalf("stdout = %q", stdout)
		}
	})

	t.Run("with backups plain", func(t *testing.T) {
		backupDir := t.TempDir()
		setup(t, backupDir)
		if err := os.WriteFile(filepath.Join(backupDir, "backup-2026-01-01-010101.tar.gz.enc"), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}

		var exit int
		stdout, stderr := captureStdoutStderr(t, func() {
			exit = ListCommand(nil)
		})
		if exit != 0 {
			t.Fatalf("exit = %d, stderr=%s", exit, stderr)
		}
		for _, want := range []string{"NAME", "CREATED", "SIZE", "Total: 1 backup(s)"} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("stdout missing %q: %s", want, stdout)
			}
		}
	})

	t.Run("with backups json", func(t *testing.T) {
		backupDir := t.TempDir()
		setup(t, backupDir)
		name := "backup-2026-01-01-010101.tar.gz.enc"
		if err := os.WriteFile(filepath.Join(backupDir, name), []byte("xyz"), 0600); err != nil {
			t.Fatal(err)
		}

		var exit int
		stdout, stderr := captureStdoutStderr(t, func() {
			exit = ListCommand([]string{"--json"})
		})
		if exit != 0 {
			t.Fatalf("exit = %d, stderr=%s", exit, stderr)
		}

		var backups []BackupInfo
		if err := json.Unmarshal([]byte(stdout), &backups); err != nil {
			t.Fatalf("invalid json output: %v\n%s", err, stdout)
		}
		if len(backups) != 1 || backups[0].Name != name {
			t.Fatalf("unexpected backups: %#v", backups)
		}
	})
}
