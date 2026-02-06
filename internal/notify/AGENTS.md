# NOTIFY PACKAGE KNOWLEDGE BASE

**Scope:** `internal/notify/`

## OVERVIEW

Desktop notification sender using notify-send. 3 files, ~38 lines. Gracefully degrades if notify-send unavailable.

## STRUCTURE

```
notify/
├── notify.go        # Send/SendSuccess/SendError
└── *_test.go        # Unit tests
```

## API

```go
// Send raw notification
func Send(title, message string) error

// Send backup success notification
func SendSuccess(backupName string, duration time.Duration) error

// Send backup failure notification
func SendError(err error) error
```

## BEHAVIOR

- Uses `notify-send` command via `os/exec`
- **Graceful degradation**: Returns nil if notify-send not found
- Reports error only if notify-send exists but fails

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add notification type | `notify.go` | Wrap Send() with context |
| Change notify-send args | `Send()` | Currently: title, message |
| Handle missing notify-send | `Send()` | Returns nil silently |

## ANTI-PATTERNS

- **Never** block on notification (fire-and-forget only)
- **Don't** fail operations if notification fails
