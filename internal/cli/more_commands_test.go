package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
)

func TestGetPassword(t *testing.T) {
	t.Run("password file trims newline", func(t *testing.T) {
		pwFile := filepath.Join(t.TempDir(), "pw")
		if err := os.WriteFile(pwFile, []byte("secret\n"), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := getPassword(pwFile)
		if err != nil {
			t.Fatalf("getPassword error: %v", err)
		}
		if got != "secret" {
			t.Fatalf("password = %q", got)
		}
	})

	t.Run("password file precedence over env", func(t *testing.T) {
		t.Setenv("DOTKEEPER_PASSWORD", "from-env")
		pwFile := filepath.Join(t.TempDir(), "pw")
		if err := os.WriteFile(pwFile, []byte("from-file"), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := getPassword(pwFile)
		if err != nil {
			t.Fatalf("getPassword error: %v", err)
		}
		if got != "from-file" {
			t.Fatalf("password = %q", got)
		}
	})

	t.Run("env fallback", func(t *testing.T) {
		t.Setenv("DOTKEEPER_PASSWORD", "from-env")
		got, err := getPassword("")
		if err != nil {
			t.Fatalf("getPassword error: %v", err)
		}
		if got != "from-env" {
			t.Fatalf("password = %q", got)
		}
	})

	t.Run("file read error", func(t *testing.T) {
		_, err := getPassword(filepath.Join(t.TempDir(), "missing"))
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "failed to read password file") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestDeleteCommand(t *testing.T) {
	setup := func(t *testing.T, backupDir string) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		writeCLIConfig(t, tmp, &config.Config{BackupDir: backupDir, Files: []string{".zshrc"}})
	}

	t.Run("missing name", func(t *testing.T) {
		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = DeleteCommand(nil)
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "backup name required") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("not found", func(t *testing.T) {
		backupDir := t.TempDir()
		setup(t, backupDir)
		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = DeleteCommand([]string{"--force", "missing-backup"})
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "backup not found") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("deletes backup and metadata", func(t *testing.T) {
		backupDir := t.TempDir()
		setup(t, backupDir)
		name := "backup-2026-01-01-010101.tar.gz.enc"
		encPath := filepath.Join(backupDir, name)
		metaPath := encPath + ".meta.json"
		if err := os.WriteFile(encPath, []byte("enc"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(metaPath, []byte("{}"), 0600); err != nil {
			t.Fatal(err)
		}

		var exit int
		stdout, stderr := captureStdoutStderr(t, func() {
			exit = DeleteCommand([]string{"--force", strings.TrimSuffix(name, ".tar.gz.enc")})
		})
		if exit != 0 {
			t.Fatalf("exit = %d, stderr=%s", exit, stderr)
		}
		if !strings.Contains(stdout, "Deleted "+name) {
			t.Fatalf("stdout = %q", stdout)
		}
		if _, err := os.Stat(encPath); !os.IsNotExist(err) {
			t.Fatalf("encrypted backup should be removed, err=%v", err)
		}
		if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
			t.Fatalf("metadata should be removed, err=%v", err)
		}
	})
}

func TestScheduleHelpersAndCommand(t *testing.T) {
	t.Run("getUserSystemdDir", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		dir, err := getUserSystemdDir()
		if err != nil {
			t.Fatalf("getUserSystemdDir: %v", err)
		}
		want := filepath.Join(home, ".config", "systemd", "user")
		if dir != want {
			t.Fatalf("dir = %q, want %q", dir, want)
		}
	})

	t.Run("copyFile success and missing source", func(t *testing.T) {
		tmp := t.TempDir()
		src := filepath.Join(tmp, "src")
		dst := filepath.Join(tmp, "dst")
		if err := os.WriteFile(src, []byte("hello"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := copyFile(src, dst); err != nil {
			t.Fatalf("copyFile error: %v", err)
		}
		data, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("read dst: %v", err)
		}
		if string(data) != "hello" {
			t.Fatalf("dst content = %q", data)
		}

		if err := copyFile(filepath.Join(tmp, "nope"), filepath.Join(tmp, "out")); err == nil {
			t.Fatal("expected error for missing source")
		}
	})

	t.Run("schedule command usage and unknown", func(t *testing.T) {
		var noSubExit, unknownExit int
		_, noSubErr := captureStdoutStderr(t, func() {
			noSubExit = ScheduleCommand(nil)
		})
		_, unknownErr := captureStdoutStderr(t, func() {
			unknownExit = ScheduleCommand([]string{"nope"})
		})

		if noSubExit != 1 {
			t.Fatalf("noSubExit = %d", noSubExit)
		}
		if !strings.Contains(noSubErr, "Usage: dotkeeper schedule") {
			t.Fatalf("stderr = %q", noSubErr)
		}
		if unknownExit != 1 {
			t.Fatalf("unknownExit = %d", unknownExit)
		}
		if !strings.Contains(unknownErr, "Unknown subcommand") {
			t.Fatalf("stderr = %q", unknownErr)
		}
	})
}
