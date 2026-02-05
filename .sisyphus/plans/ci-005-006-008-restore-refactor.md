# CLI Restore Refactor: CI-008, CI-005, CI-006

## TL;DR

> **Quick Summary**: Refactor CLI restore command to use the internal/restore package, eliminating ~100 lines of duplicated code and gaining atomic writes, proper conflict resolution, and new `--dry-run` and `--diff` flags.
> 
> **Deliverables**:
> - Refactored `internal/cli/restore.go` using `restore.Restore()` API
> - New `--dry-run` flag for preview mode
> - New `--diff` flag for unified diff output
> - CLI unit tests in `internal/cli/restore_test.go`
> 
> **Estimated Effort**: Medium (1-2 days)
> **Parallel Execution**: NO - sequential (CI-008 is prerequisite for CI-005/006)
> **Critical Path**: Task 1 → Task 2 → Task 3 → Task 4

---

## Context

### Original Request
Implement CI-008, CI-005, CI-006 from IDEATION_CODE_IMPROVEMENTS.md:
- CI-008: Refactor CLI Restore to Use internal/restore Package
- CI-005: Add Dry-Run Flag to CLI Restore
- CI-006: Add Diff Preview to CLI Restore

### Interview Summary
**Key Discussions**:
- Progress output: Summary only, no per-file progress
- Diff behavior: `--diff` allowed without `--dry-run` (show diffs, then restore)
- Test strategy: Add CLI unit tests following project patterns
- Conflict behavior: Change from "skip" to "create .bak backup + overwrite"
- `--force` semantics: Keep flag but always create backups (safer behavior)

**Research Findings**:
- `internal/restore` package has complete API: `Restore()`, `RestoreOptions`, `RestoreResult`
- TUI already uses this pattern correctly (views/restore.go:191-202)
- Current CLI duplicates tar extraction logic (lines 95-200)
- Existing tests in `internal/restore/` cover all features (969 lines)
- No CLI unit tests for restore.go (gap to fill)

### Metis Review
**Identified Gaps** (addressed):
- Behavior change on conflict: User chose new behavior (create .bak)
- Force flag semantics mismatch: User chose "overwrite with backup"
- Exit code mapping: Keep current semantics (0/1/2)
- Missing acceptance criteria: Added detailed verification commands

---

## Work Objectives

### Core Objective
Replace duplicated restore logic in CLI with the battle-tested `internal/restore` package, add `--dry-run` and `--diff` flags, and create comprehensive unit tests.

### Concrete Deliverables
- `internal/cli/restore.go` refactored (~100 lines removed, ~30 lines added)
- `internal/cli/restore_test.go` created (new file, ~200 lines)
- Updated help text with new flags

### Definition of Done
- [x] `make test` passes with all new tests
- [x] `make build` succeeds
- [x] `./dotkeeper restore --help` shows `--dry-run` and `--diff` flags
- [x] Manual verification of all flag combinations

### Must Have
- Atomic writes (temp file + rename) for all restored files
- `.bak.TIMESTAMP` backups for existing files
- `--dry-run` flag that makes no file changes
- `--diff` flag that shows unified diffs
- Exit code 0 for success, 1 for error, 2 for partial success

### Must NOT Have (Guardrails)
- No changes to `internal/restore/*` package (it's complete)
- No per-file progress output (summary only)
- No `--target-dir` or `--select-files` flags (scope creep)
- No colored diff output
- No interactive conflict prompts
- No changes to password handling (`getPassword()` stays as-is)

---

## Verification Strategy (MANDATORY)

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> ALL tasks in this plan MUST be verifiable WITHOUT any human action.
> This is NOT conditional — it applies to EVERY task, regardless of test strategy.

### Test Decision
- **Infrastructure exists**: YES (Go testing)
- **Automated tests**: YES (Tests-after approach)
- **Framework**: Standard Go testing (`go test`)

### Agent-Executed QA Scenarios (MANDATORY — ALL tasks)

Each task includes specific verification scenarios using:
- **CLI**: Bash commands with curl-style assertions
- **TUI/CLI**: interactive_bash (tmux) for interactive testing

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately):
└── Task 1: CI-008 - Refactor CLI restore to use internal/restore

Wave 2 (After Wave 1):
└── Task 2: CI-005+006 - Add --dry-run and --diff flags

Wave 3 (After Wave 2):
└── Task 3: Create CLI unit tests

Wave 4 (After Wave 3):
└── Task 4: Final verification and cleanup

Critical Path: Task 1 → Task 2 → Task 3 → Task 4
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2, 3, 4 | None |
| 2 | 1 | 3, 4 | None |
| 3 | 2 | 4 | None |
| 4 | 3 | None | None (final) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Agents |
|------|-------|-------------------|
| 1 | Task 1 | delegate_task(category="unspecified-low", load_skills=[]) |
| 2 | Task 2 | delegate_task(category="quick", load_skills=[]) |
| 3 | Task 3 | delegate_task(category="unspecified-low", load_skills=[]) |
| 4 | Task 4 | delegate_task(category="quick", load_skills=[]) |

---

## TODOs

- [x] 1. CI-008: Refactor CLI restore to use internal/restore package

  **What to do**:
  - Add import: `"github.com/diogo/dotkeeper/internal/restore"`
  - Replace `restoreBackup()` function body with call to `restore.Restore()`
  - Delete `extractArchive()` function entirely (lines 133-200)
  - Delete unused imports (`archive/tar`, `compress/gzip`, `strings`)
  - Map `RestoreResult` fields to CLI output format
  - Keep existing `--force` and `--password-file` flags
  - Update output to show `.bak` files created

  **Must NOT do**:
  - Do NOT modify `internal/restore/*` package
  - Do NOT change password handling (`getPassword()`)
  - Do NOT add new flags yet (that's Task 2)
  - Do NOT add per-file progress output

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
    - Reason: Straightforward refactoring following established pattern
  - **Skills**: `[]`
    - No special skills needed - standard Go refactoring
  - **Skills Evaluated but Omitted**:
    - `git-master`: Not needed - single commit at end

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1 (solo)
  - **Blocks**: Tasks 2, 3, 4
  - **Blocked By**: None (can start immediately)

  **References** (CRITICAL - Be Exhaustive):

  **Pattern References** (existing code to follow):
  - `internal/tui/views/restore.go:191-202` - TUI's usage of restore.Restore() - FOLLOW THIS PATTERN
  - `internal/cli/backup.go:89-116` - getPassword() function to reuse for password handling

  **API/Type References** (contracts to implement against):
  - `internal/restore/types.go:6-27` - RestoreOptions struct definition
  - `internal/restore/types.go:29-39` - RestoreResult struct to map to CLI output
  - `internal/restore/restore.go:18` - Restore() function signature

  **Code to Replace**:
  - `internal/cli/restore.go:95-130` - restoreBackup() function - REPLACE body
  - `internal/cli/restore.go:133-200` - extractArchive() function - DELETE entirely

  **WHY Each Reference Matters**:
  - TUI pattern shows exactly how to call Restore() and handle RestoreResult
  - RestoreOptions shows available options (use Force=false per user decision)
  - RestoreResult shows what fields to display in CLI output

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios (MANDATORY):**

  ```
  Scenario: Basic restore works after refactor
    Tool: Bash
    Preconditions: Backup exists at ~/.config/dotkeeper/backups/, password file at /tmp/testpw
    Steps:
      1. Create test file: echo "test content" > /tmp/test-restore-file.txt
      2. Create backup: ./dotkeeper backup --password-file /tmp/testpw
      3. Modify test file: echo "modified" > /tmp/test-restore-file.txt
      4. Run restore: ./dotkeeper restore backup-*.tar.gz.enc --password-file /tmp/testpw
      5. Check exit code: echo $?
      6. Check output contains "Restore completed"
      7. Check .bak file created: ls /tmp/test-restore-file.txt.bak.*
    Expected Result: Exit 0, restore succeeds, .bak file exists
    Evidence: Terminal output captured

  Scenario: Wrong password returns error
    Tool: Bash
    Preconditions: Backup exists, wrong password in /tmp/wrongpw
    Steps:
      1. echo "wrongpassword" > /tmp/wrongpw
      2. ./dotkeeper restore backup-*.tar.gz.enc --password-file /tmp/wrongpw
      3. Check exit code: echo $?
      4. Check stderr contains "wrong password" or "decrypt"
    Expected Result: Exit 1, error message about password
    Evidence: Terminal output captured

  Scenario: Output shows .bak files created
    Tool: Bash
    Preconditions: File exists that will be backed up during restore
    Steps:
      1. Run restore on existing files
      2. Check output contains ".bak" or "backup" count
    Expected Result: Output mentions backup files created
    Evidence: Terminal output captured
  ```

  **Commit**: YES
  - Message: `refactor(cli): use internal/restore package for CLI restore command (CI-008)`
  - Files: `internal/cli/restore.go`
  - Pre-commit: `make test`

---

- [x] 2. CI-005 + CI-006: Add --dry-run and --diff flags

  **What to do**:
  - Add `--dry-run` flag: `dryRun := fs.Bool("dry-run", false, "Preview restore without making changes")`
  - Add `--diff` flag: `showDiff := fs.Bool("diff", false, "Show differences between backup and current files")`
  - Pass flags to RestoreOptions: `opts.DryRun = *dryRun`, `opts.ShowDiff = *showDiff`
  - Set `opts.DiffWriter = os.Stdout` when `--diff` is enabled
  - Update usage help text with new flags
  - Handle combined `--dry-run --diff` case

  **Must NOT do**:
  - Do NOT require `--dry-run` with `--diff` (user chose: allow --diff alone)
  - Do NOT add colored output
  - Do NOT add per-file progress

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple flag additions following established pattern
  - **Skills**: `[]`
    - No special skills needed

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (solo)
  - **Blocks**: Tasks 3, 4
  - **Blocked By**: Task 1

  **References** (CRITICAL - Be Exhaustive):

  **Pattern References**:
  - `internal/cli/restore.go:20-22` - Existing flag definitions (--force, --password-file)
  - `internal/cli/list.go:29,81-87` - JSON output flag pattern (similar flag handling)

  **API/Type References**:
  - `internal/restore/types.go:8-9` - DryRun and ShowDiff fields
  - `internal/restore/types.go:23` - DiffWriter field

  **Documentation References**:
  - `internal/restore/restore.go:62-73` - How DryRun mode works (files added to SkippedFiles)
  - `internal/restore/restore.go:46-59` - How ShowDiff mode works (diffs written to DiffWriter)

  **WHY Each Reference Matters**:
  - Existing flags show exact pattern to follow
  - types.go shows RestoreOptions fields to set
  - restore.go shows how the package handles these options

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios (MANDATORY):**

  ```
  Scenario: --dry-run makes no file changes
    Tool: Bash
    Preconditions: Backup exists, test file modified
    Steps:
      1. Create marker file: echo "before" > /tmp/marker-dry-run.txt
      2. Backup the marker file
      3. Modify marker: echo "modified" > /tmp/marker-dry-run.txt
      4. Run: ./dotkeeper restore backup-*.tar.gz.enc --password-file /tmp/testpw --dry-run
      5. Check exit code: echo $?
      6. Check marker content: cat /tmp/marker-dry-run.txt
      7. Assert marker still contains "modified" (NOT restored)
      8. Assert no .bak file created: ls /tmp/marker-dry-run.txt.bak.* 2>/dev/null || echo "no bak"
    Expected Result: Exit 0, marker unchanged, no .bak files
    Evidence: Terminal output + file contents

  Scenario: --diff shows unified diff output
    Tool: Bash
    Preconditions: Backup exists with file that differs from current
    Steps:
      1. Run: ./dotkeeper restore backup-*.tar.gz.enc --password-file /tmp/testpw --diff 2>&1
      2. Check output contains "---" (diff header)
      3. Check output contains "+++" (diff header)
      4. Check output contains "@@" (hunk header)
    Expected Result: Unified diff format in output
    Evidence: Terminal output captured

  Scenario: --diff without --dry-run restores files
    Tool: Bash
    Preconditions: Backup exists
    Steps:
      1. Modify tracked file
      2. Run: ./dotkeeper restore backup-*.tar.gz.enc --password-file /tmp/testpw --diff
      3. Check diff output shown
      4. Check file was actually restored (content matches backup)
    Expected Result: Diff shown AND file restored
    Evidence: Terminal output + file contents

  Scenario: --dry-run --diff shows diff without changes
    Tool: Bash
    Preconditions: Backup exists with modifications
    Steps:
      1. Run: ./dotkeeper restore ... --dry-run --diff
      2. Check diff output shown
      3. Check NO files modified
    Expected Result: Diff shown, no file changes
    Evidence: Terminal output
  ```

  **Commit**: YES
  - Message: `feat(cli): add --dry-run and --diff flags to restore command (CI-005, CI-006)`
  - Files: `internal/cli/restore.go`
  - Pre-commit: `make test`

---

- [x] 3. Create CLI unit tests for restore command

  **What to do**:
  - Create `internal/cli/restore_test.go`
  - Follow project patterns: fixture-based with `createTestBackup()`, table-driven, `t.TempDir()`
  - Test cases:
    1. `TestRestoreCommand_Basic` - happy path restore
    2. `TestRestoreCommand_DryRun` - verify no file changes
    3. `TestRestoreCommand_Diff` - verify diff output
    4. `TestRestoreCommand_DryRunDiff` - combined flags
    5. `TestRestoreCommand_Force` - force flag behavior
    6. `TestRestoreCommand_WrongPassword` - error handling
    7. `TestRestoreCommand_MissingBackup` - error handling
    8. `TestRestoreCommand_MissingBackupName` - usage error

  **Must NOT do**:
  - Do NOT use mocks (project uses real I/O)
  - Do NOT skip cleanup (use t.TempDir())

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
    - Reason: Test writing following established patterns
  - **Skills**: `[]`
    - No special skills needed

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (solo)
  - **Blocks**: Task 4
  - **Blocked By**: Task 2

  **References** (CRITICAL - Be Exhaustive):

  **Test Pattern References**:
  - `internal/restore/restore_test.go:1-50` - Test setup pattern with createTestBackup() helper
  - `internal/restore/restore_test.go:52-100` - TestRestore_Basic showing happy path pattern
  - `internal/restore/diff_test.go:15-40` - Table-driven test pattern with t.Run()
  - Note: No existing CLI tests - this will be the first CLI test file

  **Fixture References**:
  - `internal/restore/restore_test.go` - createTestBackup() helper function

  **WHY Each Reference Matters**:
  - restore_test.go shows how to create test backups with real encryption
  - diff_test.go shows table-driven pattern for multiple test cases
  - Follow project conventions exactly for consistency

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios (MANDATORY):**

  ```
  Scenario: All CLI tests pass
    Tool: Bash
    Preconditions: Test file created
    Steps:
      1. Run: go test ./internal/cli/... -v -run TestRestore
      2. Check exit code: echo $?
      3. Count PASS lines in output
    Expected Result: Exit 0, all tests pass
    Evidence: Test output captured

  Scenario: Test coverage > 70%
    Tool: Bash
    Preconditions: Tests exist
    Steps:
      1. Run: go test ./internal/cli/... -coverprofile=/tmp/coverage.out
      2. Run: go tool cover -func=/tmp/coverage.out | grep restore.go
      3. Extract coverage percentage
    Expected Result: Coverage >= 70% for restore.go
    Evidence: Coverage report
  ```

  **Commit**: YES
  - Message: `test(cli): add unit tests for restore command`
  - Files: `internal/cli/restore_test.go`
  - Pre-commit: `make test`

---

- [x] 4. Final verification and cleanup

  **What to do**:
  - Run full test suite: `make test`
  - Run build: `make build`
  - Verify help text: `./dotkeeper restore --help`
  - Manual smoke test of all flag combinations
  - Run linter: `make lint`

  **Must NOT do**:
  - Do NOT skip any verification step

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple verification commands
  - **Skills**: `[]`
    - No special skills needed

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4 (final)
  - **Blocks**: None
  - **Blocked By**: Task 3

  **References**:
  - `Makefile` - test, build, lint targets

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios (MANDATORY):**

  ```
  Scenario: Full test suite passes
    Tool: Bash
    Steps:
      1. Run: make test
      2. Check exit code
    Expected Result: Exit 0
    Evidence: Test output

  Scenario: Build succeeds
    Tool: Bash
    Steps:
      1. Run: make build
      2. Check exit code
      3. Check binary exists: ls -la ./bin/dotkeeper
    Expected Result: Exit 0, binary exists
    Evidence: Build output

  Scenario: Help text shows new flags
    Tool: Bash
    Steps:
      1. Run: ./bin/dotkeeper restore --help
      2. Check output contains "--dry-run"
      3. Check output contains "--diff"
    Expected Result: Both flags documented
    Evidence: Help output

  Scenario: Linter passes
    Tool: Bash
    Steps:
      1. Run: make lint
      2. Check exit code
    Expected Result: Exit 0, no lint errors
    Evidence: Lint output
  ```

  **Commit**: NO (verification only, changes already committed)

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `refactor(cli): use internal/restore package for CLI restore command (CI-008)` | internal/cli/restore.go | make test |
| 2 | `feat(cli): add --dry-run and --diff flags to restore command (CI-005, CI-006)` | internal/cli/restore.go | make test |
| 3 | `test(cli): add unit tests for restore command` | internal/cli/restore_test.go | make test |

---

## Success Criteria

### Verification Commands
```bash
# All tests pass
make test
# Expected: exit 0

# Build succeeds
make build
# Expected: exit 0, ./bin/dotkeeper exists

# Help shows new flags
./bin/dotkeeper restore --help | grep -E "(dry-run|diff)"
# Expected: Both flags shown

# Dry-run makes no changes
touch /tmp/before && ./bin/dotkeeper restore <backup> --password-file /tmp/pw --dry-run && cat /tmp/before
# Expected: File unchanged

# Diff shows output
./bin/dotkeeper restore <backup> --password-file /tmp/pw --diff | grep -E "^(---|\\+\\+\\+|@@)"
# Expected: Unified diff format
```

### Final Checklist
- [x] All "Must Have" present (atomic writes, .bak backups, --dry-run, --diff)
- [x] All "Must NOT Have" absent (no progress output, no scope creep flags)
- [x] All tests pass (`make test`)
- [x] Lint passes (`make lint`) - for new code
- [x] Help text updated
