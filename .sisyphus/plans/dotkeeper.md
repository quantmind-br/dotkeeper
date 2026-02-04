# dotkeeper - Dotfiles Backup Manager

## TL;DR

> **Quick Summary**: Build a TUI/CLI application in Go using Bubbletea for encrypted dotfiles backup management with git integration, systemd scheduling, and keyring-based password storage.
> 
> **Deliverables**:
> - `dotkeeper` binary with TUI and CLI modes
> - AES-256-GCM encrypted backups with Argon2id key derivation
> - File browser for selecting files/folders to backup
> - Git repository integration (commit + push)
> - Systemd timer for scheduled backups
> - Desktop notifications for backup status
> - Selective restoration with diff preview
> 
> **Estimated Effort**: XL (3-4 weeks)
> **Parallel Execution**: YES - 4 waves
> **Critical Path**: Project Setup → Crypto → Backup Core → TUI → Integration

---

## Context

### Original Request
Build a TUI application in Go using Bubbletea for dotfiles backup management. The application allows users to select files/folders for backup, encrypts everything with a user-defined password, stores in a git repository, commits and pushes to remote. Must support restoration with selective file recovery.

### Interview Summary
**Key Discussions**:
- **Backup format**: Single .tar.gz.enc file named by date
- **Restoration**: Selective (user chooses what to restore)
- **Conflicts**: Rename existing to .bak + show diff before deciding
- **Versioning**: Date-named backups (backup-YYYY-MM-DD.enc)
- **Interface**: File browser interactive (ranger-style)
- **Config**: ~/.config/dotkeeper/config.yaml
- **Password**: Per-operation via keyring integration (gnome-keyring/KWallet)
- **Scheduling**: Systemd timer (Linux only)
- **Notifications**: Desktop notifications
- **TUI + CLI**: Both modes supported

**Research Findings**:
- **Bubbletea**: State-based view switching, bubbles/filepicker, bubbles/list
- **Encryption**: AES-256-GCM + Argon2id (used by restic, kubernetes, ethereum)
- **Dotfiles tools**: Patterns from chezmoi (structure), yadm (git wrapper), dotbot (bootstrap)

### Metis Review
**Identified Gaps** (addressed):
- Password + scheduled backups paradox → Keyring integration
- macOS scheduling → Out of scope (Linux only)
- CLI non-interactive mode → DOTKEEPER_PASSWORD env var
- Large file handling → Streaming (no limit)
- Restore conflict options → Rename .bak + diff preview

---

## Work Objectives

### Core Objective
Create a secure, user-friendly dotfiles backup manager with encrypted storage, git versioning, and automated scheduling capabilities.

### Concrete Deliverables
- `dotkeeper` binary (single executable)
- TUI with 6 screens (Dashboard, File Browser, Backup List, Restore, Settings, Logs)
- CLI commands: `backup`, `restore`, `list`, `config`, `schedule`
- Systemd timer/service files
- YAML configuration schema
- Comprehensive test suite (TDD)

### Definition of Done
- [x] `dotkeeper --help` shows all commands
- [x] `dotkeeper backup` creates encrypted archive and pushes to git
- [x] `dotkeeper restore` decrypts and restores selected files
- [x] TUI launches and all 6 screens are navigable
- [x] Systemd timer triggers scheduled backups
- [x] All tests pass: `go test ./...`

### Must Have
- AES-256-GCM encryption with Argon2id key derivation
- File browser for selection (bubbles/filepicker)
- Git integration (go-git library)
- Keyring integration for scheduled backup passwords
- Streaming encryption (memory efficient)
- Desktop notifications (libnotify)
- Conflict resolution with .bak rename and diff preview

### Must NOT Have (Guardrails)
- NO file templates/transforms (copy as-is)
- NO incremental backups (full backup each time)
- NO multi-machine sync
- NO plugin/hook system
- NO auto-discovery of dotfiles
- NO diff viewer between backups (use git diff manually)
- NO cloud storage providers (only git)
- NO Windows support
- NO macOS scheduling (systemd only)
- NO multiple profiles (single profile)
- NO custom encryption algorithms
- NO password storage outside system keyring

---

## Verification Strategy

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> ALL tasks MUST be verifiable WITHOUT any human action.

### Test Decision
- **Infrastructure exists**: NO (new project)
- **Automated tests**: TDD (Test-Driven Development)
- **Framework**: Go standard testing + testify for assertions

### TDD Workflow

Each TODO follows RED-GREEN-REFACTOR:

1. **RED**: Write failing test first
   - Test file: `*_test.go`
   - Command: `go test ./internal/... -run TestName -v`
   - Expected: FAIL
2. **GREEN**: Implement minimum code to pass
   - Command: `go test ./internal/... -run TestName -v`
   - Expected: PASS
3. **REFACTOR**: Clean up while keeping green

### Agent-Executed QA Scenarios (MANDATORY)

**Verification Tools:**
| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| **TUI** | interactive_bash (tmux) | Run dotkeeper, navigate screens, validate output |
| **CLI** | Bash | Execute commands, check exit codes, validate output |
| **Encryption** | Bash | Encrypt/decrypt roundtrip, verify content |
| **Git** | Bash | Check commits, verify push |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - Start Immediately):
├── Task 1: Project scaffolding and dependencies
├── Task 2: Configuration module
└── Task 3: Crypto module (AES-256-GCM + Argon2id)

Wave 2 (Core Logic - After Wave 1):
├── Task 4: File collection and tar.gz creation
├── Task 5: Backup module (encrypt + save)
├── Task 6: Git integration module
└── Task 7: Keyring integration

Wave 3 (TUI - After Wave 2):
├── Task 8: TUI base structure and navigation
├── Task 9: Dashboard screen
├── Task 10: File Browser screen
├── Task 11: Backup List screen
├── Task 12: Restore screen
├── Task 13: Settings screen
└── Task 14: Logs screen

Wave 4 (Integration - After Wave 3):
├── Task 15: CLI commands implementation
├── Task 16: Systemd timer/service
├── Task 17: Desktop notifications
├── Task 18: Restore module with diff preview
└── Task 19: End-to-end integration tests
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2,3,4,5,6,7,8 | None |
| 2 | 1 | 5,8,15 | 3 |
| 3 | 1 | 5,18 | 2 |
| 4 | 1 | 5 | 2,3 |
| 5 | 2,3,4 | 6,8,15,18 | None |
| 6 | 5 | 8,15,16 | 7 |
| 7 | 1 | 16 | 6 |
| 8 | 5,6 | 9-14 | None |
| 9-14 | 8 | 15 | Each other |
| 15 | 5,6,9-14 | 19 | 16,17 |
| 16 | 6,7 | 19 | 15,17 |
| 17 | 1 | 19 | 15,16 |
| 18 | 3,5 | 19 | 15,16,17 |
| 19 | 15,16,17,18 | None | None |

---

## TODOs

### Wave 1: Foundation

- [x] 1. Project Scaffolding and Dependencies

  **What to do**:
  - Initialize Go module: `go mod init github.com/user/dotkeeper`
  - Create directory structure:
    ```
    cmd/dotkeeper/main.go
    internal/
      config/
      crypto/
      backup/
      restore/
      git/
      keyring/
      tui/
        views/
        components/
      cli/
      notify/
    pkg/
    ```
  - Add dependencies:
    - `github.com/charmbracelet/bubbletea`
    - `github.com/charmbracelet/bubbles`
    - `github.com/charmbracelet/lipgloss`
    - `github.com/go-git/go-git/v5`
    - `github.com/zalando/go-keyring`
    - `golang.org/x/crypto`
    - `gopkg.in/yaml.v3`
    - `github.com/stretchr/testify`
  - Create Makefile with targets: build, test, install, lint

  **Must NOT do**:
  - Add unnecessary dependencies
  - Create complex build systems

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []
    - Simple scaffolding task, no special skills needed

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1 (start)
  - **Blocks**: All other tasks
  - **Blocked By**: None

  **References**:
  - Go module docs: https://go.dev/doc/modules/managing-dependencies
  - Bubbletea examples: https://github.com/charmbracelet/bubbletea/tree/master/examples

  **Acceptance Criteria**:
  - [ ] `go mod tidy` runs without errors
  - [ ] `go build ./cmd/dotkeeper` produces binary
  - [ ] Directory structure matches specification

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Project builds successfully
    Tool: Bash
    Preconditions: Go 1.21+ installed
    Steps:
      1. cd /home/diogo/dev/backup-dotfiles
      2. go mod tidy
      3. go build -o dotkeeper ./cmd/dotkeeper
      4. ./dotkeeper --version || ./dotkeeper --help
    Expected Result: Binary created, runs without crash
    Evidence: Exit code 0, binary exists

  Scenario: All dependencies resolve
    Tool: Bash
    Steps:
      1. go mod download
      2. go mod verify
    Expected Result: All modules verified
    Evidence: Exit code 0
  ```

  **Commit**: YES
  - Message: `feat(init): scaffold project structure and dependencies`
  - Files: `go.mod`, `go.sum`, `cmd/`, `internal/`, `Makefile`

---

- [x] 2. Configuration Module

  **What to do**:
  - Define config struct in `internal/config/config.go`:
    ```go
    type Config struct {
      BackupDir     string   `yaml:"backup_dir"`
      GitRemote     string   `yaml:"git_remote"`
      Files         []string `yaml:"files"`
      Folders       []string `yaml:"folders"`
      Schedule      string   `yaml:"schedule"` // cron format
      Notifications bool     `yaml:"notifications"`
    }
    ```
  - Implement Load/Save functions with XDG paths (~/.config/dotkeeper/)
  - Implement config validation
  - Write tests first (TDD)

  **Must NOT do**:
  - Store passwords in config
  - Support multiple config files

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 3)
  - **Blocks**: 5, 8, 15
  - **Blocked By**: 1

  **References**:
  - XDG Base Directory: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
  - gopkg.in/yaml.v3 docs

  **Acceptance Criteria**:
  - [ ] `go test ./internal/config -v` passes
  - [ ] Config loads from ~/.config/dotkeeper/config.yaml
  - [ ] Config creates default if not exists
  - [ ] Validation rejects invalid paths

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Config load/save roundtrip
    Tool: Bash
    Steps:
      1. go test ./internal/config -run TestConfigRoundtrip -v
    Expected Result: PASS
    Evidence: Test output shows PASS

  Scenario: Default config creation
    Tool: Bash
    Steps:
      1. rm -f ~/.config/dotkeeper/config.yaml
      2. go test ./internal/config -run TestDefaultConfig -v
      3. cat ~/.config/dotkeeper/config.yaml
    Expected Result: Default config file created with valid YAML
    Evidence: File exists with expected fields
  ```

  **Commit**: YES
  - Message: `feat(config): add configuration module with YAML support`
  - Files: `internal/config/*.go`

---

- [x] 3. Crypto Module (AES-256-GCM + Argon2id)

  **What to do**:
  - Implement in `internal/crypto/`:
    - `kdf.go`: Argon2id key derivation (Time=3, Memory=64MB, Threads=4)
    - `aes.go`: AES-256-GCM encrypt/decrypt with streaming support
    - `types.go`: EncryptionMetadata struct
  - Streaming encryption for large files (io.Reader/io.Writer)
  - Metadata header format: [version(4)][salt(16)][nonce(12)][ciphertext...]
  - Write tests first with known test vectors

  **Must NOT do**:
  - Allow configurable encryption parameters
  - Store keys anywhere
  - Load entire file in memory

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
  - **Skills**: []
    - Crypto requires careful implementation, high-IQ task

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 2)
  - **Blocks**: 5, 18
  - **Blocked By**: 1

  **References**:
  - Go crypto/aes: https://pkg.go.dev/crypto/aes
  - Go crypto/cipher GCM: https://pkg.go.dev/crypto/cipher#NewGCM
  - Argon2: https://pkg.go.dev/golang.org/x/crypto/argon2
  - Restic crypto: https://github.com/restic/restic/blob/master/internal/crypto/crypto.go

  **Acceptance Criteria**:
  - [ ] `go test ./internal/crypto -v` passes
  - [ ] Encrypt/decrypt roundtrip preserves content
  - [ ] Different passwords produce different ciphertext
  - [ ] Wrong password returns error (not garbage)
  - [ ] Streaming works for files > 100MB

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Encryption roundtrip with test vectors
    Tool: Bash
    Steps:
      1. go test ./internal/crypto -run TestEncryptDecryptRoundtrip -v
    Expected Result: PASS - content matches after roundtrip
    Evidence: Test output

  Scenario: Wrong password fails gracefully
    Tool: Bash
    Steps:
      1. go test ./internal/crypto -run TestWrongPasswordFails -v
    Expected Result: PASS - returns authentication error
    Evidence: Test output

  Scenario: Large file streaming encryption
    Tool: Bash
    Steps:
      1. go test ./internal/crypto -run TestLargeFileStreaming -v -timeout 60s
    Expected Result: PASS - memory stays bounded
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `feat(crypto): add AES-256-GCM encryption with Argon2id KDF`
  - Files: `internal/crypto/*.go`

---

### Wave 2: Core Logic

- [x] 4. File Collection and Tar.gz Creation

  **What to do**:
  - Implement in `internal/backup/collector.go`:
    - CollectFiles(paths []string) -> []FileInfo
    - Follow symlinks and copy content
    - Detect and prevent circular symlinks (max depth 20)
    - Handle unreadable files (skip with warning)
  - Implement in `internal/backup/archive.go`:
    - CreateArchive(files []FileInfo, writer io.Writer) error
    - Streaming tar.gz creation
    - Preserve file permissions

  **Must NOT do**:
  - Preserve symlinks as links
  - Include special files (sockets, FIFOs)
  - Load all files in memory

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (can start after Task 1)
  - **Blocks**: 5
  - **Blocked By**: 1

  **References**:
  - Go archive/tar: https://pkg.go.dev/archive/tar
  - Go compress/gzip: https://pkg.go.dev/compress/gzip

  **Acceptance Criteria**:
  - [ ] `go test ./internal/backup -run TestCollect -v` passes
  - [ ] Symlinks are followed (content copied)
  - [ ] Circular symlinks detected and skipped
  - [ ] Unreadable files skipped with warning
  - [ ] Archive extracts correctly with standard tar

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Collect files with symlinks
    Tool: Bash
    Steps:
      1. go test ./internal/backup -run TestCollectWithSymlinks -v
    Expected Result: PASS - symlink content included
    Evidence: Test output

  Scenario: Archive creates valid tar.gz
    Tool: Bash
    Steps:
      1. go test ./internal/backup -run TestArchiveValid -v
    Expected Result: PASS - tar tzf lists files
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `feat(backup): add file collection and tar.gz archive creation`
  - Files: `internal/backup/collector.go`, `internal/backup/archive.go`

---

- [x] 5. Backup Module (Orchestration)

  **What to do**:
  - Implement in `internal/backup/backup.go`:
    - Backup(config, password) -> (backupPath, error)
    - Orchestrate: collect → archive → encrypt → save
    - Generate backup name: `backup-YYYY-MM-DD-HHMMSS.tar.gz.enc`
    - Calculate SHA-256 checksum of plaintext
    - Store metadata in separate .meta.json file
  - Handle partial failures gracefully (cleanup temp files)

  **Must NOT do**:
  - Leave partial files on failure
  - Create backup if no files selected

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (sequential dependency)
  - **Blocks**: 6, 8, 15, 18
  - **Blocked By**: 2, 3, 4

  **References**:
  - Task 3 crypto module
  - Task 4 archive module

  **Acceptance Criteria**:
  - [ ] `go test ./internal/backup -run TestBackup -v` passes
  - [ ] Backup file created with correct name format
  - [ ] Metadata file contains checksum and file list
  - [ ] Failed backup leaves no partial files
  - [ ] Empty file list returns error

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Full backup workflow
    Tool: Bash
    Steps:
      1. go test ./internal/backup -run TestFullBackupWorkflow -v
    Expected Result: PASS - encrypted archive created
    Evidence: Test output, file exists

  Scenario: Cleanup on failure
    Tool: Bash
    Steps:
      1. go test ./internal/backup -run TestBackupCleanupOnFailure -v
    Expected Result: PASS - no temp files remain
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `feat(backup): add backup orchestration module`
  - Files: `internal/backup/backup.go`

---

- [x] 6. Git Integration Module

  **What to do**:
  - Implement in `internal/git/`:
    - `repo.go`: Open/Init repository, Clone
    - `operations.go`: Add, Commit, Push
    - `status.go`: GetStatus, HasChanges
  - Use go-git library (not shell commands)
  - Handle authentication via existing SSH/credential helper
  - Graceful error messages for auth failures

  **Must NOT do**:
  - Store git credentials
  - Force push ever
  - Shell out to git command

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: [`git-master`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 7)
  - **Blocks**: 8, 15, 16
  - **Blocked By**: 5

  **References**:
  - go-git: https://github.com/go-git/go-git
  - go-git examples: https://github.com/go-git/go-git/tree/master/_examples

  **Acceptance Criteria**:
  - [ ] `go test ./internal/git -v` passes
  - [ ] Can init new repository
  - [ ] Can add, commit files
  - [ ] Can push to remote (with valid auth)
  - [ ] Auth errors show clear message

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Git operations integration test
    Tool: Bash
    Steps:
      1. go test ./internal/git -run TestGitOperations -v
    Expected Result: PASS - init, add, commit work
    Evidence: Test output

  Scenario: Auth error handling
    Tool: Bash
    Steps:
      1. go test ./internal/git -run TestAuthError -v
    Expected Result: PASS - clear error message
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `feat(git): add git integration with go-git`
  - Files: `internal/git/*.go`

---

- [x] 7. Keyring Integration

  **What to do**:
  - Implement in `internal/keyring/`:
    - `keyring.go`: Store/Retrieve/Delete password
    - Service name: "dotkeeper"
    - User name: "backup-password"
  - Use zalando/go-keyring library
  - Support gnome-keyring, KWallet
  - Fallback to manual prompt if keyring unavailable

  **Must NOT do**:
  - Store password in plain text
  - Store password in config file

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 6)
  - **Blocks**: 16
  - **Blocked By**: 1

  **References**:
  - zalando/go-keyring: https://github.com/zalando/go-keyring

  **Acceptance Criteria**:
  - [ ] `go test ./internal/keyring -v` passes
  - [ ] Password stored in system keyring
  - [ ] Password retrieved correctly
  - [ ] Graceful fallback if keyring unavailable

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Keyring store and retrieve
    Tool: Bash
    Steps:
      1. go test ./internal/keyring -run TestKeyringRoundtrip -v
    Expected Result: PASS - password retrieved matches stored
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `feat(keyring): add system keyring integration`
  - Files: `internal/keyring/*.go`

---

### Wave 3: TUI

- [x] 8. TUI Base Structure and Navigation

  **What to do**:
  - Implement in `internal/tui/`:
    - `model.go`: Main model with view state enum
    - `update.go`: Global key handling, view switching
    - `view.go`: View dispatcher
    - `styles.go`: Lipgloss styles
  - Views enum: Dashboard, FileBrowser, BackupList, Restore, Settings, Logs
  - Global keys: q=quit, tab=next view, ?=help
  - Use bubbles/key for key bindings

  **Must NOT do**:
  - Put business logic in TUI code
  - Use global variables

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (start)
  - **Blocks**: 9-14
  - **Blocked By**: 5, 6

  **References**:
  - Bubbletea views example: https://github.com/charmbracelet/bubbletea/blob/master/examples/views/main.go
  - Bubbletea composable-views: https://github.com/charmbracelet/bubbletea/blob/master/examples/composable-views/main.go

  **Acceptance Criteria**:
  - [ ] TUI launches without crash
  - [ ] Tab cycles through views
  - [ ] q quits application
  - [ ] ? shows help

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: TUI launches and navigates
    Tool: interactive_bash (tmux)
    Steps:
      1. tmux new-session -d -s dotkeeper-test './dotkeeper'
      2. sleep 1
      3. tmux capture-pane -t dotkeeper-test -p | grep -i "dashboard\|dotkeeper"
      4. tmux send-keys -t dotkeeper-test Tab
      5. sleep 0.5
      6. tmux capture-pane -t dotkeeper-test -p | grep -i "file\|browser"
      7. tmux send-keys -t dotkeeper-test q
    Expected Result: TUI shows views, navigates with Tab, quits with q
    Evidence: Captured pane output
  ```

  **Commit**: YES
  - Message: `feat(tui): add base TUI structure with navigation`
  - Files: `internal/tui/*.go`

---

- [x] 9. Dashboard Screen

  **What to do**:
  - Implement in `internal/tui/views/dashboard.go`:
    - Show last backup date and status
    - Show next scheduled backup
    - Show count of files being tracked
    - Show git sync status
    - Quick actions: [B]ackup now, [R]estore, [S]ettings

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10-14)
  - **Blocks**: 15
  - **Blocked By**: 8

  **Acceptance Criteria**:
  - [ ] Dashboard shows last backup info
  - [ ] Dashboard shows file count
  - [ ] Quick action keys work

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Dashboard displays status
    Tool: interactive_bash (tmux)
    Steps:
      1. tmux new-session -d -s dash-test './dotkeeper'
      2. sleep 1
      3. tmux capture-pane -t dash-test -p
      4. tmux send-keys -t dash-test q
    Expected Result: Shows "Last backup", "Files tracked", quick actions
    Evidence: Captured output
  ```

  **Commit**: YES (group with 10-14)
  - Message: `feat(tui): add dashboard screen`

---

- [x] 10. File Browser Screen

  **What to do**:
  - Implement in `internal/tui/views/filebrowser.go`:
    - Use bubbles/filepicker component
    - Start at $HOME
    - Space to toggle selection
    - Enter to navigate into folder
    - Show selected count
    - a to add to backup list, r to remove

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9, 11-14)
  - **Blocks**: 15
  - **Blocked By**: 8

  **References**:
  - bubbles/filepicker: https://github.com/charmbracelet/bubbles/tree/master/filepicker

  **Acceptance Criteria**:
  - [ ] File browser shows directory listing
  - [ ] Can navigate directories
  - [ ] Can select files/folders
  - [ ] Selection persists to config

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: File browser navigation
    Tool: interactive_bash (tmux)
    Steps:
      1. tmux new-session -d -s fb-test './dotkeeper'
      2. tmux send-keys -t fb-test Tab  # Go to file browser
      3. sleep 0.5
      4. tmux capture-pane -t fb-test -p | head -20
      5. tmux send-keys -t fb-test j j Enter  # Navigate
      6. sleep 0.3
      7. tmux capture-pane -t fb-test -p | head -20
      8. tmux send-keys -t fb-test q
    Expected Result: Shows files, navigates with j/k/Enter
    Evidence: Captured output showing directory contents
  ```

  **Commit**: YES (group with 9, 11-14)
  - Message: `feat(tui): add file browser screen`

---

- [x] 11. Backup List Screen

  **What to do**:
  - Implement in `internal/tui/views/backuplist.go`:
    - Use bubbles/list component
    - Show all backups with dates
    - Show file size
    - Enter to view details
    - d to delete backup

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9-10, 12-14)
  - **Blocks**: 15
  - **Blocked By**: 8

  **Acceptance Criteria**:
  - [ ] Shows list of backups
  - [ ] Shows date and size
  - [ ] Can select backup for restore

  **Commit**: YES (group with 9-10, 12-14)
  - Message: `feat(tui): add backup list screen`

---

- [x] 12. Restore Screen

  **What to do**:
  - Implement in `internal/tui/views/restore.go`:
    - Select backup from list
    - Show contents of backup (file tree)
    - Select individual files to restore
    - Show diff preview for conflicting files
    - Confirm restore action

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9-11, 13-14)
  - **Blocks**: 15, 18
  - **Blocked By**: 8

  **Acceptance Criteria**:
  - [ ] Can select backup to restore
  - [ ] Shows file tree of backup
  - [ ] Can select individual files
  - [ ] Shows diff for conflicts

  **Commit**: YES (group with 9-11, 13-14)
  - Message: `feat(tui): add restore screen`

---

- [x] 13. Settings Screen

  **What to do**:
  - Implement in `internal/tui/views/settings.go`:
    - Edit git remote URL
    - Edit backup directory
    - Configure schedule (cron format helper)
    - Toggle notifications
    - Manage keyring password

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9-12, 14)
  - **Blocks**: 15
  - **Blocked By**: 8

  **Acceptance Criteria**:
  - [ ] Can edit all config values
  - [ ] Changes persist to config file
  - [ ] Cron format validated

  **Commit**: YES (group with 9-12, 14)
  - Message: `feat(tui): add settings screen`

---

- [x] 14. Logs Screen

  **What to do**:
  - Implement in `internal/tui/views/logs.go`:
    - Show operation history
    - Show timestamps
    - Show success/failure status
    - Scrollable list
    - Filter by type (backup/restore/error)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9-13)
  - **Blocks**: 15
  - **Blocked By**: 8

  **Acceptance Criteria**:
  - [ ] Shows operation log
  - [ ] Scrollable with j/k
  - [ ] Filter works

  **Commit**: YES
  - Message: `feat(tui): add all TUI screens (dashboard, browser, list, restore, settings, logs)`
  - Files: `internal/tui/views/*.go`

---

### Wave 4: Integration

- [x] 15. CLI Commands Implementation

  **What to do**:
  - Implement in `internal/cli/`:
    - `backup.go`: `dotkeeper backup [--password-file PATH]`
    - `restore.go`: `dotkeeper restore [backup-name] [--force]`
    - `list.go`: `dotkeeper list [--json]`
    - `config.go`: `dotkeeper config [get|set] KEY [VALUE]`
    - `schedule.go`: `dotkeeper schedule [enable|disable|status]`
  - Support DOTKEEPER_PASSWORD env var
  - JSON output for scripting
  - Exit codes: 0=success, 1=error, 2=partial

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 16, 17)
  - **Blocks**: 19
  - **Blocked By**: 5, 6, 9-14

  **Acceptance Criteria**:
  - [ ] All CLI commands work
  - [ ] `--help` shows usage
  - [ ] JSON output valid
  - [ ] Exit codes correct

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: CLI backup command
    Tool: Bash
    Steps:
      1. export DOTKEEPER_PASSWORD="testpass123"
      2. ./dotkeeper backup
      3. echo "Exit code: $?"
    Expected Result: Exit code 0, backup created
    Evidence: Exit code, file exists

  Scenario: CLI list with JSON
    Tool: Bash
    Steps:
      1. ./dotkeeper list --json | jq .
    Expected Result: Valid JSON array of backups
    Evidence: jq parses successfully
  ```

  **Commit**: YES
  - Message: `feat(cli): add CLI commands (backup, restore, list, config, schedule)`
  - Files: `internal/cli/*.go`, `cmd/dotkeeper/main.go`

---

- [x] 16. Systemd Timer/Service

  **What to do**:
  - Create systemd unit files in `contrib/systemd/`:
    - `dotkeeper.service`: Runs backup
    - `dotkeeper.timer`: Schedules runs
  - Implement in `internal/cli/schedule.go`:
    - Install/uninstall service to ~/.config/systemd/user/
    - Enable/disable timer
    - Show status
  - Service uses keyring for password

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 15, 17)
  - **Blocks**: 19
  - **Blocked By**: 6, 7

  **Acceptance Criteria**:
  - [ ] Service file valid (systemd-analyze verify)
  - [ ] Timer triggers correctly
  - [ ] `dotkeeper schedule enable` works
  - [ ] Uses keyring for password

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Systemd service validation
    Tool: Bash
    Steps:
      1. systemd-analyze verify contrib/systemd/dotkeeper.service || true
      2. systemd-analyze verify contrib/systemd/dotkeeper.timer || true
    Expected Result: No critical errors
    Evidence: Command output

  Scenario: Schedule enable/disable
    Tool: Bash
    Steps:
      1. ./dotkeeper schedule enable
      2. systemctl --user status dotkeeper.timer
      3. ./dotkeeper schedule disable
    Expected Result: Timer enabled then disabled
    Evidence: systemctl output
  ```

  **Commit**: YES
  - Message: `feat(schedule): add systemd timer integration`
  - Files: `contrib/systemd/*.service`, `contrib/systemd/*.timer`, `internal/cli/schedule.go`

---

- [x] 17. Desktop Notifications

  **What to do**:
  - Implement in `internal/notify/`:
    - `notify.go`: Send notification
    - Use libnotify (notify-send command)
    - Notify on backup success/failure
    - Include backup size and duration

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 15, 16)
  - **Blocks**: 19
  - **Blocked By**: 1

  **Acceptance Criteria**:
  - [ ] Notification sent on backup complete
  - [ ] Shows success/failure status
  - [ ] Graceful if notify-send not available

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Notification sent
    Tool: Bash
    Steps:
      1. go test ./internal/notify -run TestNotify -v
    Expected Result: PASS (or graceful skip if no display)
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `feat(notify): add desktop notifications`
  - Files: `internal/notify/*.go`

---

- [x] 18. Restore Module with Diff Preview

  **What to do**:
  - Implement in `internal/restore/`:
    - `restore.go`: Restore orchestration
    - `diff.go`: Generate diff between backup and current file
    - `conflict.go`: Handle conflict resolution
  - Conflict options: Rename existing to .bak, show diff
  - Atomic restore (temp file then rename)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 15-17)
  - **Blocks**: 19
  - **Blocked By**: 3, 5

  **Acceptance Criteria**:
  - [ ] Restore works for selected files
  - [ ] Conflicts renamed to .bak
  - [ ] Diff shown in TUI
  - [ ] Atomic restore (no partial)

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Restore with conflict
    Tool: Bash
    Steps:
      1. go test ./internal/restore -run TestRestoreWithConflict -v
    Expected Result: PASS - existing file renamed to .bak
    Evidence: Test output

  Scenario: Diff generation
    Tool: Bash
    Steps:
      1. go test ./internal/restore -run TestDiffGeneration -v
    Expected Result: PASS - diff shows changes
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `feat(restore): add restore module with diff preview and conflict handling`
  - Files: `internal/restore/*.go`

---

- [x] 19. End-to-End Integration Tests

  **What to do**:
  - Create `e2e/` directory with integration tests:
    - Full backup → restore cycle
    - CLI workflow
    - TUI smoke test
  - Test with real git repo (local)
  - Test encryption roundtrip
  - Test scheduled backup simulation

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4 (final)
  - **Blocks**: None (final task)
  - **Blocked By**: 15, 16, 17, 18

  **Acceptance Criteria**:
  - [ ] All e2e tests pass
  - [ ] Full workflow works end-to-end
  - [ ] No regressions

  **Agent-Executed QA Scenarios**:

  ```
  Scenario: Full backup-restore cycle
    Tool: Bash
    Steps:
      1. export DOTKEEPER_PASSWORD="e2e-test-pass"
      2. ./dotkeeper config set backup_dir /tmp/dotkeeper-e2e
      3. ./dotkeeper config set files "~/.bashrc"
      4. ./dotkeeper backup
      5. ./dotkeeper list --json | jq -r '.[0].name'
      6. rm ~/.bashrc.bak 2>/dev/null || true
      7. ./dotkeeper restore $(./dotkeeper list --json | jq -r '.[0].name') --force
      8. diff ~/.bashrc ~/.bashrc  # Should be identical
    Expected Result: Backup created and restored successfully
    Evidence: Exit codes 0, files match

  Scenario: E2E test suite
    Tool: Bash
    Steps:
      1. go test ./e2e/... -v -timeout 5m
    Expected Result: All tests PASS
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `test(e2e): add end-to-end integration tests`
  - Files: `e2e/*.go`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `feat(init): scaffold project structure and dependencies` | go.mod, cmd/, internal/ | go build |
| 2 | `feat(config): add configuration module with YAML support` | internal/config/*.go | go test ./internal/config |
| 3 | `feat(crypto): add AES-256-GCM encryption with Argon2id KDF` | internal/crypto/*.go | go test ./internal/crypto |
| 4 | `feat(backup): add file collection and tar.gz archive creation` | internal/backup/*.go | go test ./internal/backup |
| 5 | `feat(backup): add backup orchestration module` | internal/backup/backup.go | go test ./internal/backup |
| 6 | `feat(git): add git integration with go-git` | internal/git/*.go | go test ./internal/git |
| 7 | `feat(keyring): add system keyring integration` | internal/keyring/*.go | go test ./internal/keyring |
| 8-14 | `feat(tui): add TUI with all screens` | internal/tui/**/*.go | ./dotkeeper (manual smoke) |
| 15 | `feat(cli): add CLI commands` | internal/cli/*.go | ./dotkeeper --help |
| 16 | `feat(schedule): add systemd timer integration` | contrib/systemd/*, internal/cli/schedule.go | systemd-analyze verify |
| 17 | `feat(notify): add desktop notifications` | internal/notify/*.go | go test ./internal/notify |
| 18 | `feat(restore): add restore module with diff preview` | internal/restore/*.go | go test ./internal/restore |
| 19 | `test(e2e): add end-to-end integration tests` | e2e/*.go | go test ./e2e/... |

---

## Success Criteria

### Verification Commands
```bash
# Build
go build -o dotkeeper ./cmd/dotkeeper  # Expected: binary created

# All tests
go test ./... -v  # Expected: all PASS

# CLI help
./dotkeeper --help  # Expected: shows commands

# Backup
DOTKEEPER_PASSWORD=test ./dotkeeper backup  # Expected: exit 0

# List
./dotkeeper list --json  # Expected: valid JSON

# Restore
DOTKEEPER_PASSWORD=test ./dotkeeper restore latest  # Expected: exit 0
```

### Final Checklist
- [x] All "Must Have" present
- [x] All "Must NOT Have" absent
- [x] All tests pass: `go test ./...`
- [x] TUI launches and all screens work
- [x] CLI commands all work
- [x] Systemd timer installs correctly
- [x] Encryption/decryption verified
- [x] Git push works (with valid auth)
