# CLI PACKAGE KNOWLEDGE BASE

**Scope:** `internal/cli/`

## OVERVIEW

CLI command handlers for non-TUI usage. 5 files, ~2,200 lines.

## STRUCTURE

```
cli/
├── backup.go    # `dotkeeper backup` command
├── restore.go   # `dotkeeper restore` command
├── list.go      # `dotkeeper list` command
├── config.go    # `dotkeeper config` command
└── schedule.go  # `dotkeeper schedule` command (systemd)
```

## PASSWORD PRIORITY

`backup.go:getPassword()` implements cascade:

```
1. --password-file flag (highest)
2. DOTKEEPER_PASSWORD env var
3. System keyring (via internal/keyring)
4. Interactive prompt (if TTY)
```

## COMMANDS

| Command | File | Key Flags |
|---------|------|-----------|
| backup | `backup.go` | `--password-file`, `--dry-run` |
| restore | `restore.go` | `--backup-id`, `--password-file`, `--conflict` |
| list | `list.go` | `--json`, `--format` |
| config | `config.go` | `--edit`, `--show` |
| schedule | `schedule.go` | `--enable`, `--disable` |

## CONVENTIONS

- **Flag parsing**: Standard `flag` package
- **Error output**: `fmt.Fprintf(os.Stderr, ...)` then `os.Exit(1)`
- **JSON output**: Use `encoding/json` for `--json` flag
- **Headless support**: All commands work without TTY (for systemd)

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add CLI flag | Respective `xxx.go` | Add to `flag.*Var()` calls |
| Change password flow | `backup.go:getPassword()` | Priority order defined here |
| Add output format | `list.go` | JSON vs table vs plain |
| Modify systemd integration | `schedule.go` | User systemd, not system |

## ANTI-PATTERNS

- **Never** prompt for password in non-TTY (breaks automation)
- **Never** shell out to git (use `internal/git`)
- **Never** store password in command history
