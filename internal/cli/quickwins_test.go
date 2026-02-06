package cli

import (
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
)

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "lowercase", in: "backup_dir", want: "backup_dir"},
		{name: "uppercase", in: "BACKUP_DIR", want: "backup_dir"},
		{name: "mixed with hyphen", in: "BackUp-Dir", want: "backup_dir"},
		{name: "underscore", in: "git_remote", want: "git_remote"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeKey(tt.in); got != tt.want {
				t.Fatalf("normalizeKey(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGetConfigValue(t *testing.T) {
	cfg := &config.Config{
		BackupDir:     "/tmp/backups",
		GitRemote:     "https://example.com/repo.git",
		Schedule:      "0 2 * * *",
		Notifications: true,
		Files:         []string{".zshrc", ".gitconfig"},
		Folders:       []string{".config", ".ssh"},
	}

	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{name: "backup_dir", key: "backup_dir", want: "/tmp/backups"},
		{name: "git_remote", key: "git_remote", want: "https://example.com/repo.git"},
		{name: "schedule", key: "schedule", want: "0 2 * * *"},
		{name: "notifications", key: "notifications", want: "true"},
		{name: "files", key: "files", want: ".zshrc,.gitconfig"},
		{name: "folders", key: "folders", want: ".config,.ssh"},
		{name: "normalized hyphen", key: "backup-dir", want: "/tmp/backups"},
		{name: "unknown", key: "unknown", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getConfigValue(cfg, tt.key)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("getConfigValue(%q) expected error", tt.key)
				}
				return
			}
			if err != nil {
				t.Fatalf("getConfigValue(%q) returned error: %v", tt.key, err)
			}
			if got != tt.want {
				t.Fatalf("getConfigValue(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestSetConfigValue(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		assertion func(t *testing.T, cfg *config.Config)
		wantErr   bool
	}{
		{
			name:  "backup_dir",
			key:   "backup_dir",
			value: "/tmp/new-backups",
			assertion: func(t *testing.T, cfg *config.Config) {
				if cfg.BackupDir != "/tmp/new-backups" {
					t.Fatalf("BackupDir = %q", cfg.BackupDir)
				}
			},
		},
		{
			name:  "git_remote",
			key:   "git-remote",
			value: "git@example.com:repo.git",
			assertion: func(t *testing.T, cfg *config.Config) {
				if cfg.GitRemote != "git@example.com:repo.git" {
					t.Fatalf("GitRemote = %q", cfg.GitRemote)
				}
			},
		},
		{
			name:  "notifications true aliases",
			key:   "notifications",
			value: "yes",
			assertion: func(t *testing.T, cfg *config.Config) {
				if !cfg.Notifications {
					t.Fatal("Notifications should be true")
				}
			},
		},
		{
			name:  "notifications false aliases",
			key:   "notifications",
			value: "off",
			assertion: func(t *testing.T, cfg *config.Config) {
				if cfg.Notifications {
					t.Fatal("Notifications should be false")
				}
			},
		},
		{
			name:  "files empty",
			key:   "files",
			value: "",
			assertion: func(t *testing.T, cfg *config.Config) {
				if len(cfg.Files) != 0 {
					t.Fatalf("Files should be empty: %#v", cfg.Files)
				}
			},
		},
		{
			name:  "files list",
			key:   "files",
			value: ".zshrc,.bashrc",
			assertion: func(t *testing.T, cfg *config.Config) {
				if len(cfg.Files) != 2 || cfg.Files[0] != ".zshrc" || cfg.Files[1] != ".bashrc" {
					t.Fatalf("Files = %#v", cfg.Files)
				}
			},
		},
		{
			name:  "folders list",
			key:   "folders",
			value: ".config,.local/share",
			assertion: func(t *testing.T, cfg *config.Config) {
				if len(cfg.Folders) != 2 || cfg.Folders[0] != ".config" || cfg.Folders[1] != ".local/share" {
					t.Fatalf("Folders = %#v", cfg.Folders)
				}
			},
		},
		{name: "invalid bool", key: "notifications", value: "maybe", wantErr: true},
		{name: "unknown key", key: "not_a_key", value: "x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			err := setConfigValue(cfg, tt.key, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("setConfigValue(%q, %q) expected error", tt.key, tt.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("setConfigValue(%q, %q) unexpected error: %v", tt.key, tt.value, err)
			}
			if tt.assertion != nil {
				tt.assertion(t, cfg)
			}
		})
	}
}

