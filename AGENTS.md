# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-06
**Commit:** a0cc9df
**Branch:** feat/settings-inline-actions
**Go:** 1.25.6

## OVERVIEW

dotkeeper — Encrypted dotfiles backup manager with TUI and CLI interfaces. Go + BubbleTea + AES-256-GCM + Argon2id. Backs up to local directory with optional git sync.

## STRUCTURE

```
dotkeeper/
├── cmd/dotkeeper/main.go    # Entry point: dual TUI/CLI router
├── internal/
│   ├── backup/              # Collect → Archive → Encrypt → Save (6 files)
│   ├── restore/             # Decrypt → Extract → Conflict → Atomic write (7 files)
│   ├── crypto/              # AES-256-GCM + Argon2id KDF (6 files)
│   ├── config/              # YAML config with XDG paths (2 files)
│   ├── cli/                 # CLI command handlers (5 files, ~2200 lines)
│   ├── tui/                 # BubbleTea framework + views (~4300 lines total)
│   │   ├── views/           # Dashboard, BackupList, Restore, Settings, Logs
│   │   └── components/      # Reusable UI components
│   ├── git/                 # go-git wrapper (6 files)
│   ├── keyring/             # System keyring for headless password (2 files)
│   ├── history/             # JSONL operation history with file locking (2 files)
│   ├── pathutil/            # Path utilities, scanner, presets (8 files)
│   └── notify/              # Desktop notifications (2 files)
├── e2e/                     # End-to-end integration tests (3 files)
└── contrib/systemd/         # Service + timer for scheduled backups
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add CLI command | `internal/cli/<command>.go` + `cmd/dotkeeper/main.go` | Add handler, add case to switch |
| Add TUI view | `internal/tui/views/<view>.go` + wire in model/update/view | Add ViewState, eager-init in NewModel() |
| Change encryption | `internal/crypto/` | NEVER change defaults (breaks compatibility) |
| Modify backup flow | `internal/backup/backup.go` | collect → archive → encrypt → write |
| Modify restore flow | `internal/restore/restore.go` | decrypt → extract → conflict → atomic write |
| Add config field | `internal/config/config.go` | Add to struct + yaml tag + Validate() |
| Password sources | `internal/cli/backup.go:getPassword()` | Priority: file → env → keyring → prompt |
| Git operations | `internal/git/` | go-git library only, no shell-out |
| Operation history | `internal/history/` | JSONL format with advisory file locking |
| Path scanning | `internal/pathutil/scanner.go` | Glob + preset expansion |

## TUI ARCHITECTURE

BubbleTea with **eager-initialized** sub-models (all views in memory for fast switching):

```go
type Model struct {
    state     ViewState              // DashboardView, BackupListView, etc.
    dashboard views.DashboardModel   // Created in NewModel(), not on-demand
    // ... all views always in memory
}

// After Update(), REQUIRED type assertion:
model, cmd = m.dashboard.Update(msg)
m.dashboard = model.(views.DashboardModel)
```

**Conventions:**
- Private messages (`statusMsg`) for internal async
- Public messages (`BackupSuccessMsg`) for cross-view events
- Async via closures: `return func() tea.Msg { /* blocking I/O */ }`
- Tab cycles views: `m.state = (m.state + 1) % viewCount`

## CRYPTO SPECIFICATION

```
Ciphertext format: [version(1)][salt(16)][nonce(12)][ciphertext...][tag(16)]
```

- **Encryption**: AES-256-GCM (authenticated)
- **KDF**: Argon2id (3 iterations, 64MB, 4 threads)
- **Salt**: 16 bytes random per backup
- **Nonce**: 12 bytes random per encryption
- **KeySize**: 32 bytes
- **Metadata**: Stored in `.meta.json` (unencrypted - contains salt/KDF params only)

**NEVER CHANGE CONSTANTS** - Would break backward compatibility with existing backups.

## CONVENTIONS

- **Error wrapping**: Always `fmt.Errorf("context: %w", err)`
- **Atomic writes**: temp file + rename pattern
- **Config location**: `$XDG_CONFIG_HOME/dotkeeper/config.yaml`
- **Backup naming**: `backup-YYYY-MM-DD-HHMMSS.tar.gz.enc` + `.meta.json`
- **Permissions**: Backup files always `0600`
- **Tests co-located**: `foo.go` → `foo_test.go` (same package)
- **Symlinks**: Followed and content copied (not preserved as links)
- **No globals in TUI**: Pass config via constructor
- **Git operations**: Use go-git library, never shell out
- **File locking**: `syscall.Flock` for history.jsonl concurrent access

## ANTI-PATTERNS (THIS PROJECT)

**NEVER:**
- Store passwords outside system keyring
- Use `as any` / type assertions without checking
- Shell out to git (use go-git library)
- Force push
- Load entire backup files in memory (use streaming where possible)
- Use `//nolint` without justification
- Change Argon2id parameters (breaks backward compatibility)
- Roll your own crypto
- Write to history without file locking
- Block in TUI `Update()` - always use `tea.Cmd`
- Access globals in TUI - pass via Model constructor

**DON'T:**
- Add incremental backups (full backup each time by design)
- Add cloud storage providers (git only)
- Add Windows support (Linux/systemd only)
- Add multiple config profiles (single config by design)
- Add file auto-discovery (explicit paths only)
- Lazy-initialize TUI views (eager init in NewModel())
- Modify other views' state directly - use messages

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

## DEPENDENCIES

Key external packages:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/go-git/go-git/v5` - Git operations
- `github.com/zalando/go-keyring` - System keyring access
- `golang.org/x/crypto` - Argon2id

## NOTES

- **Password paradox**: Scheduled backups need password → solved via system keyring
- **Large files**: Currently loads in memory; streaming TODO for very large backups
- **Restore conflicts**: Existing files renamed to `.bak` with diff preview option
- **Git integration**: Pure go-git, no shell exec for security/reliability
- **E2E tests**: Full integration tests in `e2e/` directory using real crypto
- **Concurrency**: File locking prevents history corruption when manual + systemd run simultaneously
