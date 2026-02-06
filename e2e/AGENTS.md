# E2E TEST KNOWLEDGE BASE

**Scope:** `e2e/`

## OVERVIEW

End-to-end integration tests. 3 files, ~1,400 lines. Tests full backup/restore/encryption workflows.

## STRUCTURE

```
e2e/
├── e2e_cli_test.go       # CLI command integration tests
├── e2e_encryption_test.go # Crypto integration tests
└── e2e_tui_test.go       # TUI workflow tests
```

## TEST PATTERNS

E2E tests use real filesystem operations and actual encryption:

```go
func TestFullBackupRestore(t *testing.T) {
    // Create temp dirs
    sourceDir := t.TempDir()
    backupDir := t.TempDir()
    restoreDir := t.TempDir()
    
    // Run full workflow
    // ... backup ...
    // ... restore ...
    // Verify files match
}
```

## CONVENTIONS

- **Temp dirs**: Use `t.TempDir()` for automatic cleanup
- **Real crypto**: Tests use actual AES-256-GCM + Argon2id
- **Password**: Use test passwords, never real credentials
- **Cleanup**: Always defer cleanup of test backups
- **Timeout**: E2E tests have longer timeouts (30s+)

## WHERE TO LOOK

| Task | File | Notes |
|------|------|-------|
| Add CLI workflow test | `e2e_cli_test.go` | Full command execution |
| Add crypto test | `e2e_encryption_test.go` | Encryption/decryption roundtrip |
| Add TUI test | `e2e_tui_test.go` | BubbleTea program testing |
| Test fixtures | Create `testdata/` | Large files, sample configs |

## ANTI-PATTERNS

- **Never** use production paths in tests
- **Never** commit real passwords
- **Don't** skip cleanup - leaves temp files
- **Don't** mock crypto in E2E - test real implementation

## RUNNING E2E TESTS

```bash
# Run all tests including E2E
go test -v ./...

# Run only E2E
go test -v ./e2e/...

# Run specific E2E test
go test -v ./e2e/... -run TestFullBackupRestore
```

## NOTES

- E2E tests are slower than unit tests (crypto operations)
- Tests create real tar.gz.enc files
- Uses test-specific keyring entries
- Parallel execution limited due to keyring access
