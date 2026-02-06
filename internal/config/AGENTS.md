# CONFIG PACKAGE KNOWLEDGE BASE

**Scope:** `internal/config/`

## OVERVIEW

YAML configuration management with XDG paths. 2 files, ~390 lines.

## STRUCTURE

```
config/
├── config.go       # Config struct, Load/Save/Validate
└── config_test.go  # Unit tests
```

## CONFIG FORMAT

Location: `$XDG_CONFIG_HOME/dotkeeper/config.yaml` (or `~/.config/dotkeeper/config.yaml`)

```yaml
backup_dir: /path/to/backups
files:
  - ~/.bashrc
  - ~/.vimrc
folders:
  - ~/.config/nvim
git_remote: git@github.com:user/dotfiles.git
notifications: true
```

## CONFIG STRUCT

```go
type Config struct {
    BackupDir     string   `yaml:"backup_dir"`      // Required
    GitRemote     string   `yaml:"git_remote"`      // Required
    Files         []string `yaml:"files"`           // At least one file OR folder
    Folders       []string `yaml:"folders"`
    Schedule      string   `yaml:"schedule"`        // Cron format
    Notifications bool     `yaml:"notifications"`
}
```

## API

```go
// Load from default XDG location
cfg, err := config.Load()

// Load from specific path
cfg, err := config.LoadFromPath("/custom/path/config.yaml")

// Save to default location
err := cfg.Save()

// Validate required fields
err := cfg.Validate()  // Checks backup_dir, git_remote, files/folders
```

## XDG PATHS

```go
Config:  $XDG_CONFIG_HOME/dotkeeper/config.yaml
State:   $XDG_STATE_HOME/dotkeeper/       (used by history/)
```

Fall back to `~/.config/` and `~/.local/state/` if XDG vars not set.

## VALIDATION RULES

```go
func (c *Config) Validate() error
```

- `backup_dir` must be non-empty
- `git_remote` must be non-empty
- At least one `file` OR `folder` must be specified

Returns descriptive errors for each failure.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add config field | `config.go:Config struct` | Add yaml tag |
| Change validation | `config.go:Validate()` | Check required fields |
| Change XDG path logic | `config.go:GetConfigDir()` | Config directory |
| Load defaults | `config.go:LoadOrDefault()` | Returns defaults if file missing |

## ANTI-PATTERNS

- **Never** hardcode config paths - use XDG functions
- **Never** ignore validation errors in production code
- **Don't** store sensitive data in config (passwords go to keyring)

## TESTING

```go
// Use test-specific paths
cfg := &config.Config{...}
err := cfg.SaveToPath(t.TempDir() + "/config.yaml")
```

Default config for tests:
- BackupDir: `~/.dotfiles`
- GitRemote: `https://github.com/user/dotfiles.git`
- Folders: `[".config"]`
- Schedule: `"0 2 * * *"` (2 AM daily)
