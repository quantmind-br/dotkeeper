# KEYRING PACKAGE KNOWLEDGE BASE

**Scope:** `internal/keyring/`

## OVERVIEW

System keyring wrapper for headless password storage. 2 files, ~67 lines. Enables scheduled backups without password prompts.

## STRUCTURE

```
keyring/
├── keyring.go       # Store/Retrieve/Delete/IsAvailable
└── keyring_test.go  # Unit tests
```

## API

```go
// Store password in system keyring
func Store(password string) error

// Retrieve password from keyring
func Retrieve() (string, error)

// Delete password from keyring
func Delete() error

// Check if keyring is accessible
func IsAvailable() bool
```

## SERVICE IDENTIFIER

```go
const (
    serviceName = "dotkeeper"
    userName    = "backup-password"
)
```

Used by zalando/go-keyring to identify entries.

## PASSWORD PRIORITY

In CLI backup flow (`internal/cli/backup.go`):
1. `--password-file` flag (highest)
2. `DOTKEEPER_PASSWORD` env var
3. System keyring (this package)
4. Interactive prompt (lowest)

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Change service name | `keyring.go:serviceName` | Would break existing entries |
| Add keyring check | `IsAvailable()` | Test write/read/delete cycle |
| Handle headless mode | `Retrieve()` | Returns ErrPasswordNotFound if empty |

## ANTI-PATTERNS

- **Never** store passwords elsewhere - this is the ONLY storage
- **Don't** change serviceName/username (breaks existing passwords)
