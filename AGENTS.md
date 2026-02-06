# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-05
**Commit:** 064eb68
**Branch:** master

## OVERVIEW

dotkeeper — Encrypted dotfiles backup manager with TUI and CLI interfaces. Go + BubbleTea + AES-256-GCM + Argon2id. Backs up to local directory with optional git sync.

## STRUCTURE

```
dotkeeper/
├── cmd/dotkeeper/main.go    # Entry: TUI (no args) or CLI (backup|restore|list|config)
├── internal/
│   ├── backup/              # Collect → Archive → Encrypt → Save
│   ├── restore/             # Decrypt → Extract → Conflict resolution → Atomic write
│   ├── crypto/              # AES-256-GCM + Argon2id KDF
│   ├── config/              # YAML config at ~/.config/dotkeeper/config.yaml
│   ├── cli/                 # CLI command handlers
│   ├── tui/                 # BubbleTea TUI (see TUI PATTERNS below)
│   ├── git/                 # go-git wrapper (no shell-out)
│   ├── keyring/             # System keyring for headless password
│   └── notify/              # Desktop notifications
├── e2e/                     # End-to-end tests
└── contrib/systemd/         # Service + timer for scheduled backups
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add CLI command | `internal/cli/` + `cmd/dotkeeper/main.go` | Add handler, add case to switch |
| Add TUI view | `internal/tui/views/` + `internal/tui/model.go` | Add ViewState, add sub-model, wire Update/View |
| Change encryption | `internal/crypto/` | NEVER change defaults (security) |
| Modify backup flow | `internal/backup/backup.go` | collect → archive → encrypt → write |
| Modify restore flow | `internal/restore/restore.go` | decrypt → extract → conflict → atomic write |
| Add config field | `internal/config/config.go` | Add to struct + yaml tag + Validate() |
| Password sources | `internal/cli/backup.go:getPassword()` | Priority: file → env → keyring |
| Git operations | `internal/git/` | go-git library only, no shell-out |

## TUI PATTERNS

BubbleTea architecture with eager-initialized sub-models:

```go
// Main model holds ALL views in memory (fast switching)
type Model struct {
    state     ViewState              // DashboardView, BackupListView, etc.
    dashboard views.DashboardModel   // All views initialized in NewModel()
    // ...
}

// Each view implements tea.Model
type XxxModel struct {
    config *config.Config
    width, height int
}
func (m XxxModel) Init() tea.Cmd { ... }
func (m XxxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }
func (m XxxModel) View() string { ... }
```

**Conventions:**
- Private messages (`statusMsg`) for internal async
- Public messages (`BackupSuccessMsg`) for cross-view events
- Async via closures: `return func() tea.Msg { /* blocking I/O */ }`
- Type assertions after Update: `m.dashboard = model.(views.DashboardModel)`
- Tab cycles views: `m.state = (m.state + 1) % viewCount`
- Bubbles components embedded directly (list, textinput, filepicker)

## CONVENTIONS

- **Error wrapping**: Always `fmt.Errorf("context: %w", err)`
- **Atomic writes**: temp file + rename (see `restoreFileAtomic`)
- **Config location**: XDG_CONFIG_HOME/dotkeeper/config.yaml
- **Backup naming**: `backup-YYYY-MM-DD-HHMMSS.tar.gz.enc` + `.meta.json`
- **Tests co-located**: `foo.go` → `foo_test.go` (same package)
- **No globals in TUI**: Pass config via constructor
- **Git operations**: Use go-git library, never shell out
- **Permissions**: Backup files always `0600`
- **Symlinks**: Followed and content copied (not preserved as links)

## ANTI-PATTERNS (THIS PROJECT)

**NEVER:**
- Store passwords outside system keyring
- Use `as any` / type assertions without checking
- Shell out to git (use go-git library)
- Force push
- Load entire backup files in memory (use streaming where possible)
- Add `@ts-ignore` equivalent (`//nolint` without justification)
- Change Argon2id parameters (breaks backward compatibility)
- Roll your own crypto

**DON'T:**
- Add incremental backups (full backup each time by design)
- Add cloud storage providers (git only)
- Add Windows support (Linux/systemd only)
- Add multiple config profiles (single config by design)
- Add file auto-discovery (explicit paths only)
- Block in TUI `Update()` - always use `tea.Cmd`
- Access globals in TUI - pass config via constructor

## SECURITY

- **Encryption**: AES-256-GCM (authenticated)
- **KDF**: Argon2id (3 iterations, 64MB, 4 threads)
- **Salt**: 16 bytes random per backup
- **Nonce**: 12 bytes random per encryption
- **Ciphertext format**: `[version(1)][salt(16)][nonce(12)][ciphertext...][tag(16)]`
- **Metadata**: Stored in `.meta.json` (not encrypted - contains salt/KDF params only)
- **Keyring**: zalando/go-keyring for headless mode

## COMMANDS

```bash
# Development
make build      # → ./bin/dotkeeper
make test       # -race -coverprofile=coverage.out
make lint       # golangci-lint required
make clean

# Usage
./dotkeeper                        # Launch TUI
./dotkeeper backup                 # CLI backup (needs password)
./dotkeeper backup --password-file ~/.pw
DOTKEEPER_PASSWORD=xxx ./dotkeeper backup
./dotkeeper restore --backup-id <path>
./dotkeeper list
./dotkeeper config
```

## SYSTEMD DEPLOYMENT

```bash
# Install service + timer
cp contrib/systemd/dotkeeper.* ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now dotkeeper.timer

# Requires password in keyring first:
# Store via TUI settings or: secret-tool store --label='dotkeeper' service dotkeeper username backup
```

Timer runs daily at 2:00 AM with 1-hour randomized delay.

## NOTES

- **Password paradox**: Scheduled backups need password → solved via system keyring
- **Symlinks**: Followed and content copied (not preserved as links)
- **Large files**: Currently loads in memory; streaming TODO for very large backups
- **Restore conflicts**: Existing files renamed to `.bak` with diff preview option
- **Git integration**: Pure go-git, no shell exec for security/reliability
- **E2E tests**: Full integration tests in `e2e/` directory
