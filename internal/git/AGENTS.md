# GIT PACKAGE KNOWLEDGE BASE

**Scope:** `internal/git/`

## OVERVIEW

Go-git wrapper for backup repository operations. 6 files, ~600 lines. **No shell execution** - pure Go implementation.

## STRUCTURE

```
git/
├── repo.go         # Repository initialization and opening
├── operations.go   # Commit, push, pull operations
├── status.go       # Status checks and diff
├── repo_test.go
├── operations_test.go
└── status_test.go
```

## KEY PRINCIPLE

**NEVER shell out to git binary.** Always use `github.com/go-git/go-git/v5` library.

```go
// CORRECT - use go-git
r, err := git.PlainOpen(path)

// WRONG - never do this
exec.Command("git", "add", ".").Run()
```

## API

```go
// Repository management
func InitRepository(path string) (*git.Repository, error)
func OpenRepository(path string) (*git.Repository, error)

// Operations
func CommitChanges(r *git.Repository, message string) error
func PushToRemote(r *git.Repository, remoteName string) error

// Status
func HasUncommittedChanges(r *git.Repository) (bool, error)
func GetStatus(r *git.Repository) (git.Status, error)
```

## CONVENTIONS

- **No shell**: Use go-git exclusively
- **No force push**: Force push is forbidden
- **Atomic commits**: Commit after each backup operation
- **Error wrapping**: `fmt.Errorf("git operation failed: %w", err)`
- **Worktree**: Use `r.Worktree()` for file operations

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add git operation | `operations.go` | Use go-git methods |
| Change commit behavior | `operations.go:CommitChanges()` | Commit message format |
| Status checking | `status.go` | Porcelain vs plumbing |
| Auth handling | `operations.go` | SSH key vs HTTPS token |

## ANTI-PATTERNS

- **NEVER** use `os/exec` to run git commands
- **NEVER** force push to remote
- **NEVER** store credentials in code
- **Don't** use porcelain commands - use go-git objects

## SECURITY NOTES

- Credentials handled via standard git credential helpers
- SSH keys loaded from standard locations
- No credential storage in dotkeeper code
