package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
)

func writeCLIConfig(t *testing.T, xdgHome string, cfg *config.Config) {
	t.Helper()
	configDir := filepath.Join(xdgHome, "dotkeeper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := cfg.SaveToPath(filepath.Join(configDir, "config.yaml")); err != nil {
		t.Fatalf("save config: %v", err)
	}
}

func captureStdoutStderr(t *testing.T, fn func()) (string, string) {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	fn()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, stdoutR)
	_, _ = io.Copy(&stderrBuf, stderrR)
	return stdoutBuf.String(), stderrBuf.String()
}

func TestConfigGet_Success(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	writeCLIConfig(t, tmp, &config.Config{BackupDir: "/tmp/backups", Notifications: true})

	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = configGet([]string{"backup_dir"})
	})

	if exit != 0 {
		t.Fatalf("configGet exit = %d, stderr=%s", exit, stderr)
	}
	if strings.TrimSpace(stdout) != "/tmp/backups" {
		t.Fatalf("stdout = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr should be empty, got %q", stderr)
	}
}

func TestConfigGet_Errors(t *testing.T) {
	t.Run("missing key", func(t *testing.T) {
		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = configGet(nil)
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "KEY required") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("load error", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = configGet([]string{"backup_dir"})
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "Error loading config") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("unknown key", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		writeCLIConfig(t, tmp, &config.Config{BackupDir: "/tmp/backups", Files: []string{".zshrc"}})

		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = configGet([]string{"does_not_exist"})
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "unknown key") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		configDir := filepath.Join(tmp, "dotkeeper")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("mkdir config dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("backup_dir: ["), 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = configGet([]string{"backup_dir"})
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "Error loading config") {
			t.Fatalf("stderr = %q", stderr)
		}
	})
}

func TestConfigSet_SuccessAndPersist(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = configSet([]string{"backup-dir", "/tmp/new-backups"})
	})

	if exit != 0 {
		t.Fatalf("configSet exit = %d, stderr=%s", exit, stderr)
	}
	if !strings.Contains(stdout, "Set backup-dir = /tmp/new-backups") {
		t.Fatalf("stdout = %q", stdout)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load after set: %v", err)
	}
	if cfg.BackupDir != "/tmp/new-backups" {
		t.Fatalf("BackupDir = %q", cfg.BackupDir)
	}
}

func TestConfigSet_Errors(t *testing.T) {
	t.Run("missing args", func(t *testing.T) {
		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = configSet([]string{"backup_dir"})
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "KEY and VALUE required") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("invalid bool", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		writeCLIConfig(t, tmp, &config.Config{BackupDir: "/tmp/backups", Files: []string{".zshrc"}})

		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = configSet([]string{"notifications", "maybe"})
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "invalid boolean value") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("unknown key", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		writeCLIConfig(t, tmp, &config.Config{BackupDir: "/tmp/backups", Files: []string{".zshrc"}})

		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = configSet([]string{"not_real", "value"})
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "unknown key") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("save error", func(t *testing.T) {
		tmp := t.TempDir()
		xdgAsFile := filepath.Join(tmp, "xdg-config-home-file")
		if err := os.WriteFile(xdgAsFile, []byte("x"), 0644); err != nil {
			t.Fatalf("write xdg file: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", xdgAsFile)

		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = configSet([]string{"backup_dir", "/tmp/new"})
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "Error saving config") {
			t.Fatalf("stderr = %q", stderr)
		}
	})
}

func TestConfigGet_EdgeCases(t *testing.T) {
	t.Run("normalized key with hyphen", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		writeCLIConfig(t, tmp, &config.Config{BackupDir: "/tmp/backups", Files: []string{".zshrc"}})

		var exit int
		stdout, stderr := captureStdoutStderr(t, func() {
			exit = configGet([]string{"backup-dir"})
		})
		if exit != 0 {
			t.Fatalf("exit = %d, stderr=%s", exit, stderr)
		}
		if strings.TrimSpace(stdout) != "/tmp/backups" {
			t.Fatalf("stdout = %q", stdout)
		}
	})

	t.Run("files and folders empty", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		writeCLIConfig(t, tmp, &config.Config{
			BackupDir: "/tmp/backups",
			Files:     []string{},
			Folders:   []string{},
		})

		var filesExit, foldersExit int
		filesOut, filesErr := captureStdoutStderr(t, func() {
			filesExit = configGet([]string{"files"})
		})
		foldersOut, foldersErr := captureStdoutStderr(t, func() {
			foldersExit = configGet([]string{"folders"})
		})

		if filesExit != 0 || foldersExit != 0 {
			t.Fatalf("expected exits 0, got files=%d folders=%d", filesExit, foldersExit)
		}
		if filesErr != "" || foldersErr != "" {
			t.Fatalf("expected empty stderr, got filesErr=%q foldersErr=%q", filesErr, foldersErr)
		}
		if strings.TrimSpace(filesOut) != "" {
			t.Fatalf("expected empty files output, got %q", filesOut)
		}
		if strings.TrimSpace(foldersOut) != "" {
			t.Fatalf("expected empty folders output, got %q", foldersOut)
		}
	})
}

func TestConfigSet_BooleanAliasesAndLists(t *testing.T) {
	trueAliases := []string{"true", "yes", "1", "on"}
	falseAliases := []string{"false", "no", "0", "off"}

	for _, val := range trueAliases {
		t.Run("notifications true alias "+val, func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", tmp)
			writeCLIConfig(t, tmp, &config.Config{BackupDir: "/tmp/backups", Files: []string{".zshrc"}})

			var exit int
			_, stderr := captureStdoutStderr(t, func() {
				exit = configSet([]string{"notifications", val})
			})
			if exit != 0 {
				t.Fatalf("exit = %d, stderr=%s", exit, stderr)
			}

			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("config.Load: %v", err)
			}
			if !cfg.Notifications {
				t.Fatalf("notifications should be true for alias %q", val)
			}
		})
	}

	for _, val := range falseAliases {
		t.Run("notifications false alias "+val, func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", tmp)
			writeCLIConfig(t, tmp, &config.Config{
				BackupDir:     "/tmp/backups",
				Files:         []string{".zshrc"},
				Notifications: true,
			})

			var exit int
			_, stderr := captureStdoutStderr(t, func() {
				exit = configSet([]string{"notifications", val})
			})
			if exit != 0 {
				t.Fatalf("exit = %d, stderr=%s", exit, stderr)
			}

			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("config.Load: %v", err)
			}
			if cfg.Notifications {
				t.Fatalf("notifications should be false for alias %q", val)
			}
		})
	}

	t.Run("set files and folders from comma list", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		writeCLIConfig(t, tmp, &config.Config{BackupDir: "/tmp/backups", Files: []string{".zshrc"}})

		var filesExit, foldersExit int
		_, filesErr := captureStdoutStderr(t, func() {
			filesExit = configSet([]string{"files", ".zshrc,.bashrc"})
		})
		_, foldersErr := captureStdoutStderr(t, func() {
			foldersExit = configSet([]string{"folders", ".config,.ssh"})
		})
		if filesExit != 0 || foldersExit != 0 {
			t.Fatalf("expected exits 0, got files=%d folders=%d", filesExit, foldersExit)
		}
		if filesErr != "" || foldersErr != "" {
			t.Fatalf("expected empty stderr, got filesErr=%q foldersErr=%q", filesErr, foldersErr)
		}

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("config.Load: %v", err)
		}
		if len(cfg.Files) != 2 || cfg.Files[0] != ".zshrc" || cfg.Files[1] != ".bashrc" {
			t.Fatalf("files = %#v", cfg.Files)
		}
		if len(cfg.Folders) != 2 || cfg.Folders[0] != ".config" || cfg.Folders[1] != ".ssh" {
			t.Fatalf("folders = %#v", cfg.Folders)
		}
	})
}

func TestConfigList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)
		writeCLIConfig(t, tmp, &config.Config{
			BackupDir:     "/tmp/backups",
			GitRemote:     "https://example.com/repo.git",
			Schedule:      "0 2 * * *",
			Notifications: true,
			Files:         []string{".zshrc"},
			Folders:       []string{".config"},
		})

		var exit int
		stdout, stderr := captureStdoutStderr(t, func() {
			exit = configList()
		})
		if exit != 0 {
			t.Fatalf("exit = %d, stderr=%s", exit, stderr)
		}
		for _, want := range []string{"Configuration:", "backup_dir:", "notifications:", "files:", "folders:"} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("stdout missing %q: %s", want, stdout)
			}
		}
	})

	t.Run("load error", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		var exit int
		_, stderr := captureStdoutStderr(t, func() {
			exit = configList()
		})
		if exit != 1 {
			t.Fatalf("exit = %d", exit)
		}
		if !strings.Contains(stderr, "Error loading config") {
			t.Fatalf("stderr = %q", stderr)
		}
	})
}

func TestConfigCommand_Dispatch(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	writeCLIConfig(t, tmp, &config.Config{BackupDir: "/tmp/backups", Files: []string{".zshrc"}})

	var getExit, setExit, listExit, badExit int
	_, getErr := captureStdoutStderr(t, func() { getExit = ConfigCommand([]string{"get", "backup_dir"}) })
	_, setErr := captureStdoutStderr(t, func() { setExit = ConfigCommand([]string{"set", "backup_dir", "/tmp/other"}) })
	_, listErr := captureStdoutStderr(t, func() { listExit = ConfigCommand([]string{"list"}) })
	_, badErr := captureStdoutStderr(t, func() { badExit = ConfigCommand([]string{"unknown"}) })

	if getExit != 0 || setExit != 0 || listExit != 0 {
		t.Fatalf("expected successful exits, got get=%d set=%d list=%d", getExit, setExit, listExit)
	}
	if getErr != "" || setErr != "" || listErr != "" {
		t.Fatalf("expected empty stderr for success paths")
	}
	if badExit != 1 {
		t.Fatalf("badExit = %d", badExit)
	}
	if !strings.Contains(badErr, "Unknown subcommand") {
		t.Fatalf("bad stderr = %q", badErr)
	}
}

func TestConfigCommand_NoSubcommand(t *testing.T) {
	var exit int
	_, stderr := captureStdoutStderr(t, func() {
		exit = ConfigCommand(nil)
	})
	if exit != 1 {
		t.Fatalf("exit = %d", exit)
	}
	if !strings.Contains(stderr, "Usage: dotkeeper config") {
		t.Fatalf("stderr = %q", stderr)
	}
}
