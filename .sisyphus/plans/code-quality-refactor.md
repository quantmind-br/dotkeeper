# Code Quality Refactoring — 14 Items

## TL;DR

> **Quick Summary**: Pure refactoring of 14 code quality items across the dotkeeper codebase. Deduplication, dead code removal, method extraction, file splitting, and structural improvements. Zero behavior changes, zero new features.
> 
> **Deliverables**:
> - 14 atomic commits, each passing `make test`
> - ~400 lines of dead code removed
> - ~35 lines of duplication eliminated
> - 857-line settings.go split into 4 focused files
> - 721-line restore.go reduced by ~150 lines via render extraction + component wiring
> - Generic `updateView[T]` helper eliminating 11 boilerplate blocks
> 
> **Estimated Effort**: Large (8-10 hours)
> **Parallel Execution**: YES — 3 waves within Phase 1, partial parallelism in Phase 2
> **Critical Path**: Task 0 → Phase 1 → CQ-008 → CQ-002 | CQ-003 → CQ-006

---

## Context

### Original Request
Implement all code quality improvements from `IDEATION_CODE_QUALITY.md` — a critically analyzed report with 14 actionable items organized in 3 phases plus suggestions.

### Interview Summary
**Key Decisions**:
- Scope: ALL 14 items (3 phases + suggestions)
- Commits: 1 atomic commit per CQ item (e.g., `refactor(tui): CQ-004 add generic updateView helper`)
- Test strategy: Existing Go stdlib tests as safety net — `make test` after every commit
- Include suggestions CQ-012 and CQ-013

**Research Findings**:
- All 14 items verified with exact file references and line numbers
- Test infrastructure: Go stdlib testing, ~45% coverage, table-driven subtests, `make test` runs with race detector
- All affected files already have test coverage
- Branch: `master`, Go 1.25.6, BubbleTea 1.3.10
- No CI pipeline (manual verification required)

### Metis Review
**Critical Gaps Identified & Resolved**:
- CQ-001: `pathutil.FormatSize()` only handles up to GB, while duplicates handle up to EB. **Resolution**: Upgrade canonical function to loop-based algorithm before replacing callers.
- CQ-007: The 3 "dead" aliases ARE used as value constructors (4 sites). **Resolution**: Replace constructor sites with `ErrorMsg{...}` directly when deleting aliases.
- CQ-006: Wiring components is riskier than deletion. **Resolution**: Default to deletion fallback if wiring is non-trivial (>30 min).
- Working tree: `IDEATION_CODE_QUALITY.md` is modified, coverage files exist. **Resolution**: Task 0 stashes/commits working tree before CQ work.
- `messages.go` post-deletion: Only `ErrorMsg` remains (~15 lines). **Resolution**: Acceptable — focused types file.

---

## Work Objectives

### Core Objective
Eliminate all identified code quality issues (duplication, dead code, complexity, oversized files) through pure refactoring — preserving identical behavior and passing all existing tests.

### Concrete Deliverables
- 14 atomic commits on a `refactor/code-quality` branch
- Each commit passes `make test` and `go vet ./...`

### Definition of Done
- [ ] All 14 commits created and passing
- [ ] `make test` passes with same package count as baseline
- [ ] `make build` succeeds
- [ ] `go vet ./...` clean (no output)
- [ ] Zero behavior changes in user-facing output

### Must Have
- Every commit passes tests + vet
- Existing test patterns preserved (stdlib testing, no new deps)
- All dead code removed (SuccessMsg, LoadingMsg, RefreshMsg, 3 error aliases, FileSelector, DiffViewer)
- Size formatting deduplicated to single canonical function
- Generic `updateView[T]` helper in place

### Must NOT Have (Guardrails)
- NO new dependencies or external test libraries
- NO changes to user-visible output format, error messages, or workflow
- NO modification to encryption/crypto code
- NO changes to files not explicitly listed in the CQ item being implemented
- NO `IDEATION_CODE_QUALITY.md` committed in any CQ commit
- NO sub-packages created during settings.go split (all stays in `views` package)
- NO theme switching infrastructure added in CQ-012 (just add Styles field)
- NO acceptance criteria requiring manual TUI interaction or "user visually confirms"

---

## Verification Strategy

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
> ALL tasks verified by running `make test`, `go vet`, and `make build`. No manual testing.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: Tests-after (existing tests as regression safety net, no TDD)
- **Framework**: Go stdlib `testing` package

### Baseline Capture (Task 0)
Before any CQ work begins, capture:
```bash
make test 2>&1 | grep "^ok" > /tmp/dotkeeper-test-baseline.txt
wc -l < /tmp/dotkeeper-test-baseline.txt
# Record this count — must match after all 14 commits
```

### Per-Commit Verification
After EVERY commit:
```bash
make test 2>&1 | tail -5       # Assert: all PASS, zero FAIL
go vet ./... 2>&1               # Assert: no output (clean)
```

### Final Verification (After All 14 Commits)
```bash
make test 2>&1 | grep "^ok" | wc -l    # Assert: matches baseline count
make build 2>&1                          # Assert: Build complete
```

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 0 (Setup — MUST be first):
└── Task 0: Create branch, stash working tree, capture baseline

Wave 1 (Phase 1 Quick Wins — parallel-safe):
├── Task 1: CQ-001 (formatSize dedup)
├── Task 2: CQ-005 (dead message types)
├── Task 3: CQ-007 (dead error aliases)    [depends: none within wave]
├── Task 4: CQ-010 (maxPasswordAttempts)
└── Task 5: CQ-011 (TrimSuffix)

Wave 2 (Phase 2 Structural):
├── Task 6: CQ-004 (updateView generic)     [no deps]
├── Task 7: CQ-003 (restore render methods) [no deps]
├── Task 8: CQ-008 (togglePathDisabled)     [no deps]
├── Task 9: CQ-014 (logHistory helper)      [no deps]
└── Task 10: CQ-013 (setup preset render)   [no deps]

Wave 3 (Phase 3 Larger Refactors — sequential):
├── Task 11: CQ-002 (settings split)        [depends: Task 8]
├── Task 12: CQ-006 (wire/delete components)[depends: Task 7]
└── Task 13: CQ-012 (Styles in ProgramCtx)  [no deps]

Wave 4 (Final):
└── Task 14: Final verification & cleanup
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 0 | None | ALL | None (must be first) |
| 1 | 0 | None | 2, 3, 4, 5 |
| 2 | 0 | None | 1, 3, 4, 5 |
| 3 | 0 | None | 1, 2, 4, 5 |
| 4 | 0 | None | 1, 2, 3, 5 |
| 5 | 0 | None | 1, 2, 3, 4 |
| 6 | Wave 1 | None | 7, 8, 9, 10 |
| 7 | Wave 1 | 12 | 6, 8, 9, 10 |
| 8 | Wave 1 | 11 | 6, 7, 9, 10 |
| 9 | Wave 1 | None | 6, 7, 8, 10 |
| 10 | Wave 1 | None | 6, 7, 8, 9 |
| 11 | 8 | None | 12, 13 |
| 12 | 7 | None | 11, 13 |
| 13 | Wave 2 | None | 11, 12 |
| 14 | ALL | None | None (must be last) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Agents |
|------|-------|-------------------|
| 0 | Setup | `delegate_task(category="quick", load_skills=["git-master"])` |
| 1 | 1-5 | `delegate_task(category="quick", load_skills=[])` — parallel |
| 2 | 6-10 | `delegate_task(category="unspecified-low", load_skills=["bubbletea"])` |
| 3 | 11-12 | `delegate_task(category="unspecified-high", load_skills=["bubbletea"])` |
| 3 | 13 | `delegate_task(category="quick", load_skills=["bubbletea"])` |
| 4 | 14 | `delegate_task(category="quick", load_skills=["git-master"])` |

---

## TODOs

- [ ] 0. Setup: Create Branch, Stash Working Tree, Capture Baseline

  **What to do**:
  - Stash or commit the current working tree changes (modified `IDEATION_CODE_QUALITY.md`, deleted docs, coverage files) with a housekeeping commit
  - Create and checkout branch `refactor/code-quality` from `master`
  - Run `make test` and capture baseline: `make test 2>&1 | grep "^ok" > /tmp/dotkeeper-test-baseline.txt`
  - Record the number of passing packages: `wc -l < /tmp/dotkeeper-test-baseline.txt`

  **Must NOT do**:
  - Do NOT modify any source code in this step
  - Do NOT delete any files that aren't already shown as deleted in `git status`

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]
    - `git-master`: Branch creation and stash/commit operations

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential (must be first)
  - **Blocks**: All other tasks
  - **Blocked By**: None

  **References**:
  - `Makefile` — `test` target runs `go test -v -race -coverprofile=coverage.out ./...`

  **Acceptance Criteria**:
  - [ ] Working tree clean (`git status` shows nothing)
  - [ ] On branch `refactor/code-quality`
  - [ ] `/tmp/dotkeeper-test-baseline.txt` exists with passing package lines
  - [ ] Baseline package count recorded

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Branch created and clean working tree
    Tool: Bash
    Preconditions: On master with modified files
    Steps:
      1. git status → shows modified/deleted files
      2. git add -A && git commit -m "chore: clean working tree before refactoring"
      3. git checkout -b refactor/code-quality
      4. git status → Assert: "nothing to commit, working tree clean"
      5. git branch --show-current → Assert: "refactor/code-quality"
    Expected Result: Clean branch ready for work
    Evidence: git status output

  Scenario: Test baseline captured
    Tool: Bash
    Preconditions: On refactor/code-quality branch
    Steps:
      1. make test 2>&1 | grep "^ok" > /tmp/dotkeeper-test-baseline.txt
      2. cat /tmp/dotkeeper-test-baseline.txt → Assert: multiple "ok" lines
      3. wc -l < /tmp/dotkeeper-test-baseline.txt → Assert: > 0
      4. make test 2>&1 | grep FAIL → Assert: no output (zero failures)
    Expected Result: Baseline captured, all tests passing
    Evidence: /tmp/dotkeeper-test-baseline.txt content
  ```

  **Commit**: YES
  - Message: `chore: clean working tree before code quality refactoring`
  - Files: `IDEATION_CODE_QUALITY.md`, deleted files, coverage artifacts
  - Pre-commit: `make test`

---

- [ ] 1. CQ-001: Deduplicate Size Formatting (4 impls → 1)

  **What to do**:
  - **First**: Upgrade `pathutil.FormatSize()` in `internal/pathutil/scanner.go` to use the loop-based algorithm that handles up to EB (matching the behavior of the duplicates). Keep the same function signature `func FormatSize(bytes int64) string`.
  - Add test cases for TB+ sizes in `internal/pathutil/scanner_test.go` (the existing `TestFormatSize` tests)
  - Run `go test -v -run TestFormatSize ./internal/pathutil/` → PASS
  - Replace `formatSize()` calls in `internal/cli/list.go` (lines 152, 155) with `pathutil.FormatSize()`
  - Replace `formatSize()` call in `internal/cli/history.go` (line 83) with `pathutil.FormatSize()`
  - Replace `formatBytes()` calls in `internal/tui/views/logs.go` (line 44) and `internal/tui/views/setup.go` (lines 436, 465) with `pathutil.FormatSize()`
  - Delete the private `formatSize()` function from `internal/cli/list.go` (lines 169-183)
  - Delete the private `formatBytes()` function from `internal/tui/views/logs.go` (lines 54-66)
  - Add `"github.com/diogo/dotkeeper/internal/pathutil"` import to `list.go` and `history.go` (if not already imported)
  - The `views/logs.go` and `views/setup.go` files are in the `views` package which may already import `pathutil` — verify
  - Remove unused imports if any remain after deleting the private functions
  - Update existing tests in `internal/cli/quickwins_test.go` (lines 177-198) that test the private `formatSize` — these must either be removed (since the function no longer exists) or converted to test `pathutil.FormatSize()` (already covered in scanner_test.go)
  - Update existing tests in `internal/tui/views/logs_test.go` (lines 174-196) that test the private `formatBytes` — same treatment

  **Must NOT do**:
  - Do NOT change the function signature of `FormatSize`
  - Do NOT change the output format (e.g., "1.5 GB" must stay "1.5 GB")
  - Do NOT touch any file not listed above

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3, 4, 5)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/pathutil/scanner.go:104-121` — Current `FormatSize()` implementation (switch-based, up to GB only)
  - `internal/cli/list.go:169-183` — Loop-based algorithm to adopt (handles up to EB via `"KMGTPE"[exp]`)
  - `internal/tui/views/logs.go:54-66` — Identical loop-based algorithm (same as list.go)

  **Caller References** (all sites to update):
  - `internal/cli/list.go:152,155` — Table column formatting for backup sizes
  - `internal/cli/history.go:83` — History entry size display
  - `internal/tui/views/logs.go:44` — Log entry size display
  - `internal/tui/views/setup.go:436,465` — Setup wizard file/folder size display

  **Test References**:
  - `internal/pathutil/scanner_test.go:88-106` — Existing `TestFormatSize` (5 cases up to GB)
  - `internal/cli/quickwins_test.go:177-198` — Tests for private `formatSize()` (6 cases up to TB)
  - `internal/tui/views/logs_test.go:174-196` — Tests for private `formatBytes()` (6 cases up to TB)

  **Already using pathutil.FormatSize** (no change needed):
  - `internal/tui/views/dashboard.go:128`
  - `internal/tui/views/settings.go:629`

  **Acceptance Criteria**:
  - [ ] `pathutil.FormatSize()` upgraded with loop-based algorithm handling up to EB
  - [ ] New test cases for TB and PB added and passing
  - [ ] Private `formatSize()` deleted from `cli/list.go`
  - [ ] Private `formatBytes()` deleted from `views/logs.go`
  - [ ] All 6 caller sites updated to use `pathutil.FormatSize()`
  - [ ] Tests for deleted private functions removed or redirected
  - [ ] `go test -v -run TestFormatSize ./internal/pathutil/` → PASS
  - [ ] `make test` → all PASS
  - [ ] `go vet ./...` → clean

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: FormatSize handles TB+ sizes correctly
    Tool: Bash
    Preconditions: FormatSize upgraded
    Steps:
      1. go test -v -run TestFormatSize ./internal/pathutil/ 2>&1
      2. Assert: PASS
      3. Assert: Test cases include TB/PB values
    Expected Result: All format tests pass including new TB+ cases
    Evidence: Test output captured

  Scenario: No duplicate formatSize functions remain
    Tool: Bash
    Preconditions: Duplicates deleted
    Steps:
      1. grep -rn "func formatSize" internal/ → Assert: no output
      2. grep -rn "func formatBytes" internal/ → Assert: no output
      3. grep -rn "pathutil.FormatSize" internal/ → Assert: 6+ matches
    Expected Result: Only canonical FormatSize exists
    Evidence: grep output captured

  Scenario: Full test suite passes
    Tool: Bash
    Steps:
      1. make test 2>&1 | tail -5
      2. Assert: "PASS" for all packages
      3. go vet ./... 2>&1 → Assert: no output
    Expected Result: Zero regressions
    Evidence: Test output
  ```

  **Commit**: YES
  - Message: `refactor: CQ-001 deduplicate size formatting to pathutil.FormatSize`
  - Files: `internal/pathutil/scanner.go`, `internal/pathutil/scanner_test.go`, `internal/cli/list.go`, `internal/cli/history.go`, `internal/cli/quickwins_test.go`, `internal/tui/views/logs.go`, `internal/tui/views/logs_test.go`, `internal/tui/views/setup.go`
  - Pre-commit: `make test`

---

- [ ] 2. CQ-005: Delete Dead Message Types

  **What to do**:
  - Delete `SuccessMsg`, `LoadingMsg`, `RefreshMsg` type definitions from `internal/tui/views/messages.go` (lines 10-29)
  - Keep `ErrorMsg` and its methods (`Error()`, `Unwrap()`)
  - The remaining file will have ~15 lines — this is fine as a focused types file

  **Must NOT do**:
  - Do NOT delete `ErrorMsg` or its methods
  - Do NOT touch any other file

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3, 4, 5)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:
  - `internal/tui/views/messages.go:10-29` — Three dead type definitions to delete

  **Acceptance Criteria**:
  - [ ] `SuccessMsg`, `LoadingMsg`, `RefreshMsg` deleted from `messages.go`
  - [ ] `ErrorMsg` + methods still present
  - [ ] `make test` → all PASS
  - [ ] `go vet ./...` → clean

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Dead types removed, ErrorMsg preserved
    Tool: Bash
    Steps:
      1. grep -n "SuccessMsg\|LoadingMsg\|RefreshMsg" internal/tui/views/messages.go → Assert: no output
      2. grep -n "ErrorMsg" internal/tui/views/messages.go → Assert: matches found
      3. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: Dead code removed, compilation intact
    Evidence: grep and test output
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-005 remove dead message types SuccessMsg, LoadingMsg, RefreshMsg`
  - Files: `internal/tui/views/messages.go`
  - Pre-commit: `make test`

---

- [ ] 3. CQ-007: Delete Dead Error Type Aliases (3 of 5)

  **What to do**:
  - Delete `backupDeleteErrorMsg` alias from `internal/tui/views/backuplist.go` (line 31)
  - Delete `diffErrorMsg` alias from `internal/tui/views/restore.go` (line 72)
  - Delete `restoreErrorMsg` alias from `internal/tui/views/restore.go` (line 79)
  - **CRITICAL**: Replace all construction sites that use these aliases with `ErrorMsg{...}` directly:
    - `backuplist.go` — find all `backupDeleteErrorMsg{...}` constructions and replace with `ErrorMsg{Source: "backup-delete", Err: ...}`
    - `restore.go` — find all `diffErrorMsg{...}` constructions and replace with `ErrorMsg{Source: "restore-diff", Err: ...}`
    - `restore.go` — find all `restoreErrorMsg{...}` constructions and replace with `ErrorMsg{Source: "restore", Err: ...}`
  - Keep `BackupErrorMsg` and `passwordInvalidMsg` (they ARE used in switch statements)

  **Must NOT do**:
  - Do NOT delete `BackupErrorMsg` (used in `backuplist.go:163` switch)
  - Do NOT delete `passwordInvalidMsg` (used in `restore.go:434` switch)
  - Do NOT change the `Source` field values — they must match what the switch handlers expect

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 4, 5)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Type Alias Definitions** (to delete):
  - `internal/tui/views/backuplist.go:31` — `type backupDeleteErrorMsg = ErrorMsg`
  - `internal/tui/views/restore.go:72` — `type diffErrorMsg = ErrorMsg`
  - `internal/tui/views/restore.go:79` — `type restoreErrorMsg = ErrorMsg`

  **Construction Sites** (to update):
  - `internal/tui/views/backuplist.go` — grep for `backupDeleteErrorMsg{` to find instantiations
  - `internal/tui/views/restore.go` — grep for `diffErrorMsg{` and `restoreErrorMsg{` to find instantiations

  **Switch Handlers** (DO NOT TOUCH):
  - `internal/tui/views/backuplist.go:163` — `case BackupErrorMsg:` — routes by Source field
  - `internal/tui/views/restore.go:434` — `case passwordInvalidMsg:` — routes by Source field

  **Acceptance Criteria**:
  - [ ] 3 alias definitions deleted
  - [ ] All construction sites updated to use `ErrorMsg{...}` directly
  - [ ] `BackupErrorMsg` and `passwordInvalidMsg` aliases preserved
  - [ ] `make test` → all PASS
  - [ ] `go vet ./...` → clean

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Dead aliases removed, live ones preserved
    Tool: Bash
    Steps:
      1. grep -n "backupDeleteErrorMsg\|diffErrorMsg\|restoreErrorMsg" internal/tui/views/*.go → Assert: no matches (definitions AND usages gone)
      2. grep -n "BackupErrorMsg\|passwordInvalidMsg" internal/tui/views/*.go → Assert: matches exist (preserved)
      3. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: Dead aliases removed, routing preserved
    Evidence: grep and test output
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-007 remove dead error type aliases and inline constructors`
  - Files: `internal/tui/views/backuplist.go`, `internal/tui/views/restore.go`
  - Pre-commit: `make test`

---

- [ ] 4. CQ-010: Extract maxPasswordAttempts Constant

  **What to do**:
  - Add `const maxPasswordAttempts = 3` near the top of `internal/tui/views/restore.go` (near the other constants)
  - Replace `m.passwordAttempts >= 3` with `m.passwordAttempts >= maxPasswordAttempts` (line 438)
  - Replace `"attempt %d/3"` with `"attempt %d/%d", ..., maxPasswordAttempts` in the error format string

  **Must NOT do**:
  - Do NOT change the actual limit value (keep it at 3)
  - Do NOT change error message wording beyond the constant substitution

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 5)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:
  - `internal/tui/views/restore.go:438` — Magic number `3` in condition and format string

  **Acceptance Criteria**:
  - [ ] `maxPasswordAttempts` constant defined
  - [ ] No magic `3` remains in password attempt logic
  - [ ] `make test` → all PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Magic number replaced with constant
    Tool: Bash
    Steps:
      1. grep -n "maxPasswordAttempts" internal/tui/views/restore.go → Assert: constant definition + usage
      2. grep -n "Attempts >= 3\|attempt %d/3" internal/tui/views/restore.go → Assert: no matches
      3. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: Named constant in use
    Evidence: grep output
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-010 extract maxPasswordAttempts constant`
  - Files: `internal/tui/views/restore.go`
  - Pre-commit: `make test`

---

- [ ] 5. CQ-011: Replace Manual String Trimming with stdlib

  **What to do**:
  - In `internal/cli/backup.go` (lines 113-117), replace:
    ```go
    password := string(data)
    if len(password) > 0 && password[len(password)-1] == '\n' {
        password = password[:len(password)-1]
    }
    ```
    With:
    ```go
    password := strings.TrimSuffix(string(data), "\n")
    ```
  - Ensure `strings` is imported (may already be)

  **Must NOT do**:
  - Do NOT use `strings.TrimRight` (strips ALL trailing newlines, behavior change)
  - Do NOT use `strings.TrimSpace` (strips more than just newlines)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 4)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:
  - `internal/cli/backup.go:113-117` — Manual newline stripping code

  **Acceptance Criteria**:
  - [ ] Manual if-block replaced with `strings.TrimSuffix`
  - [ ] `make test` → all PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Manual trimming replaced with stdlib
    Tool: Bash
    Steps:
      1. grep -n "TrimSuffix" internal/cli/backup.go → Assert: match found
      2. grep -n "password\[len(password)-1\]" internal/cli/backup.go → Assert: no match
      3. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: stdlib function in use
    Evidence: grep output
  ```

  **Commit**: YES
  - Message: `refactor(cli): CQ-011 use strings.TrimSuffix for password newline trimming`
  - Files: `internal/cli/backup.go`
  - Pre-commit: `make test`

---

- [ ] 6. CQ-004: Create Generic updateView Helper (11 repetitions → helper)

  **What to do**:
  - Add a generic helper function in `internal/tui/update.go`:
    ```go
    func updateView[T tea.Model](view T, msg tea.Msg) (T, tea.Cmd) {
        model, cmd := view.Update(msg)
        if v, ok := model.(T); ok {
            return v, cmd
        }
        return view, cmd
    }
    ```
  - Replace ALL 11 type-assertion-after-Update patterns:
    - 5 in `propagateWindowSize()` (lines 31-59): dashboard, backupList, restore, settings, logs
    - 6 in `Update()` (lines 111-116 for setup, lines 249-285 for dashboard/backupList/restore/settings/logs)
  - Each 4-5 line block becomes 2 lines:
    ```go
    m.dashboard, cmd = updateView(m.dashboard, viewMsg)
    cmds = append(cmds, cmd)
    ```
  - If Go type inference fails for `updateView(m.dashboard, viewMsg)`, use explicit type params: `updateView[views.DashboardModel](m.dashboard, viewMsg)`

  **Must NOT do**:
  - Do NOT change function signatures of `propagateWindowSize()` or `Update()`
  - Do NOT change the behavior of any view Update call
  - Do NOT modify test files — `update_test.go` tests the outer Model.Update(), not sub-view updates

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: [`bubbletea`]
    - `bubbletea`: BubbleTea framework patterns, Model-Update-View cycle

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 8, 9, 10)
  - **Blocks**: None
  - **Blocked By**: Wave 1 complete

  **References**:

  **Pattern Reference** (code to replace):
  - `internal/tui/update.go:31-59` — 5 identical blocks in `propagateWindowSize()`
  - `internal/tui/update.go:111-116` — Setup model Update in `Update()` function
  - `internal/tui/update.go:249-285` — 5 identical blocks in view-routing switch

  **Test Reference**:
  - `internal/tui/update_test.go` — 583 lines, tests propagateWindowSize and Update thoroughly

  **Acceptance Criteria**:
  - [ ] `updateView[T tea.Model]` helper defined in `update.go`
  - [ ] All 11 type-assertion blocks replaced with 2-line calls
  - [ ] `go test -v -run TestPropagateWindowSize ./internal/tui/` → PASS
  - [ ] `go test -v -run TestUpdate ./internal/tui/` → PASS
  - [ ] `make test` → all PASS
  - [ ] `go vet ./...` → clean

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Generic helper works with type inference
    Tool: Bash
    Steps:
      1. grep -c "updateView(" internal/tui/update.go → Assert: 11 (matches replacement count)
      2. grep -c "if.*ok := .*\.\(views\." internal/tui/update.go → Assert: 0 (old pattern gone)
      3. go test -v -run "TestPropagateWindowSize|TestUpdate" ./internal/tui/ 2>&1 → Assert: PASS
      4. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: All boilerplate eliminated, tests pass
    Evidence: grep counts and test output
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-004 add generic updateView helper, eliminate 11 type assertion blocks`
  - Files: `internal/tui/update.go`
  - Pre-commit: `make test`

---

- [ ] 7. CQ-003: Extract Restore View Render Methods

  **What to do**:
  - In `internal/tui/views/restore.go`, refactor the `View()` method:
  - Convert the 6 sequential `if m.phase == phaseXxx { ... return ... }` blocks into a switch statement
  - Extract each phase's rendering into a dedicated private method:
    - `func (m RestoreModel) renderBackupList() string`
    - `func (m RestoreModel) renderPassword() string`
    - `func (m RestoreModel) renderFileSelect() string`
    - `func (m RestoreModel) renderRestoring() string`
    - `func (m RestoreModel) renderDiffPreview() string`
    - `func (m RestoreModel) renderResults() string`
  - The `View()` method becomes a simple switch dispatcher (~15 lines)
  - Move the render methods to the same file (or keep in restore.go — no file split needed for this task)

  **Must NOT do**:
  - Do NOT change the rendered output of any phase
  - Do NOT modify the Update() method or phase transitions
  - Do NOT create new files (this is method extraction within restore.go)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: [`bubbletea`]
    - `bubbletea`: BubbleTea View() pattern, rendering conventions

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 8, 9, 10)
  - **Blocks**: Task 12 (CQ-006 depends on this)
  - **Blocked By**: Wave 1 complete

  **References**:
  - `internal/tui/views/restore.go:531-655` — Current View() with 6 if-blocks
  - `internal/tui/views/restore_test.go` — Existing tests for restore view

  **Acceptance Criteria**:
  - [ ] `View()` method uses switch statement with 6 cases
  - [ ] 6 `renderXxx()` methods extracted
  - [ ] No `if m.phase ==` patterns remain in View()
  - [ ] `go test -v ./internal/tui/views/ -run TestRestore` → PASS
  - [ ] `make test` → all PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: View uses switch dispatcher
    Tool: Bash
    Steps:
      1. grep -A2 "func (m RestoreModel) View" internal/tui/views/restore.go → Assert: contains "switch m.phase"
      2. grep -c "func (m RestoreModel) render" internal/tui/views/restore.go → Assert: 6
      3. grep -c "if m.phase ==" internal/tui/views/restore.go → Assert: 0 (in View method)
      4. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: Clean switch + 6 render methods
    Evidence: grep output and test results
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-003 extract restore View() into 6 phase-specific render methods`
  - Files: `internal/tui/views/restore.go`
  - Pre-commit: `make test`

---

- [ ] 8. CQ-008: Extract togglePathDisabled Method

  **What to do**:
  - In `internal/tui/views/settings.go`, extract the space-key handler logic (lines 374-393) into a dedicated method:
    ```go
    func (m *SettingsModel) togglePathDisabled(lt pathListType, path string) {
        disabled := m.disabledPathsForType(lt)
        for i, d := range disabled {
            if d == path {
                m.setDisabledPathsForType(lt, append(disabled[:i], disabled[i+1:]...))
                return
            }
        }
        m.setDisabledPathsForType(lt, append(disabled, path))
    }
    ```
  - Replace the inline logic in the `case " ":` handler with a call to `m.togglePathDisabled(lt, path)`

  **Must NOT do**:
  - Do NOT change the toggle behavior
  - Do NOT modify any other key handlers

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 7, 9, 10)
  - **Blocks**: Task 11 (CQ-002 depends on this — the extracted method moves cleanly to `settings_paths.go`)
  - **Blocked By**: Wave 1 complete

  **References**:
  - `internal/tui/views/settings.go:374-393` — Inline toggle logic in space-key handler
  - `internal/tui/views/settings.go` — `disabledPathsForType()` and `setDisabledPathsForType()` methods (already exist)
  - `internal/tui/views/settings_test.go` — Existing tests

  **Acceptance Criteria**:
  - [ ] `togglePathDisabled()` method exists
  - [ ] Inline logic replaced with method call
  - [ ] Nesting in space-key handler reduced
  - [ ] `make test` → all PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Method extracted, nesting reduced
    Tool: Bash
    Steps:
      1. grep -n "togglePathDisabled" internal/tui/views/settings.go → Assert: method definition + call site
      2. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: Cleaner code structure
    Evidence: grep output
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-008 extract togglePathDisabled method to reduce nesting`
  - Files: `internal/tui/views/settings.go`
  - Pre-commit: `make test`

---

- [ ] 9. CQ-014: Extract CLI History Logging Helper

  **What to do**:
  - Create a `logHistory()` helper function in `internal/cli/` (in a shared file like `helpers.go` or at the top of `backup.go`):
    ```go
    func logHistory(store *history.Store, storeErr error, entry history.HistoryEntry) {
        if storeErr != nil {
            fmt.Fprintf(os.Stderr, "Warning: history unavailable: %v\n", storeErr)
            return
        }
        if err := store.Append(entry); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: failed to log history: %v\n", err)
        }
    }
    ```
  - Replace all 5 inline logging patterns:
    - `internal/cli/backup.go` (lines 74-79, 96-100) — 2 occurrences
    - `internal/cli/restore.go` (lines 91-97, 115-121, 128-132) — 3 occurrences
  - Check the exact signature of `history.HistoryEntry` — it may need to be `history.Entry` or similar. Verify before implementing.

  **Must NOT do**:
  - Do NOT change the warning messages or stderr output
  - Do NOT add error handling beyond what exists (best-effort logging)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 7, 8, 10)
  - **Blocks**: None
  - **Blocked By**: Wave 1 complete

  **References**:
  - `internal/cli/backup.go:74-79,96-100` — 2 inline logging patterns
  - `internal/cli/restore.go:91-97,115-121,128-132` — 3 inline logging patterns
  - `internal/history/` — History store types and `Append()` method

  **Acceptance Criteria**:
  - [ ] `logHistory()` helper function defined
  - [ ] All 5 inline patterns replaced with helper calls
  - [ ] `make test` → all PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Helper extracts all duplicated patterns
    Tool: Bash
    Steps:
      1. grep -rn "logHistory(" internal/cli/ → Assert: 5+ matches (1 definition + 5 calls)
      2. grep -rn "Warning: failed to log history" internal/cli/ → Assert: only inside logHistory function
      3. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: DRY history logging
    Evidence: grep output
  ```

  **Commit**: YES
  - Message: `refactor(cli): CQ-014 extract logHistory helper for repeated best-effort logging pattern`
  - Files: `internal/cli/backup.go`, `internal/cli/restore.go` (and `helpers.go` if created)
  - Pre-commit: `make test`

---

- [ ] 10. CQ-013: Extract Shared Preset Rendering in Setup View

  **What to do**:
  - In `internal/tui/views/setup.go`, extract the nearly identical rendering code for `StepPresetFiles` and `StepPresetFolders` (lines 418-473) into a shared function:
    ```go
    func renderPresetList(presets []pathutil.DotfilePreset, cursor int, selected map[int]bool, st styles.Styles) string {
        var s strings.Builder
        for i, p := range presets {
            // shared cursor/checkbox rendering logic
        }
        return s.String()
    }
    ```
  - Replace both rendering blocks with calls to `renderPresetList()`
  - Verify the exact parameters needed by reading the current code (cursor, selection state, style)

  **Must NOT do**:
  - Do NOT change the visual output of the preset lists
  - Do NOT modify the setup wizard flow or step transitions

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`bubbletea`]
    - `bubbletea`: BubbleTea rendering patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 7, 8, 9)
  - **Blocks**: None
  - **Blocked By**: Wave 1 complete

  **References**:
  - `internal/tui/views/setup.go:418-444` — StepPresetFiles rendering
  - `internal/tui/views/setup.go:446-473` — StepPresetFolders rendering (nearly identical)

  **Acceptance Criteria**:
  - [ ] `renderPresetList()` function defined
  - [ ] Both preset views use the shared function
  - [ ] `make test` → all PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Shared rendering function extracted
    Tool: Bash
    Steps:
      1. grep -n "renderPresetList" internal/tui/views/setup.go → Assert: 1 definition + 2 call sites
      2. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: DRY preset rendering
    Evidence: grep output
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-013 extract shared renderPresetList for setup wizard`
  - Files: `internal/tui/views/setup.go`
  - Pre-commit: `make test`

---

- [ ] 11. CQ-002: Split Settings View into 4 Files

  **What to do**:
  - Split `internal/tui/views/settings.go` (857 lines) into 4 files, all in the `views` package:

  | New File | Contents | ~Lines |
  |----------|----------|--------|
  | `settings.go` | Model struct, types, constants, `NewSettings()`, `Update()`, `Init()`, `Refresh()` | ~200 |
  | `settings_editing.go` | `handleEditingFieldInput()`, `handleEditingSubItemInput()`, `startEditingField()`, `startEditingSubItem()`, `saveFieldValue()` | ~200 |
  | `settings_paths.go` | `handleBrowsingPathsInput()`, path type helpers, `refreshPathList()`, `togglePathDisabled()` (from CQ-008), `disabledPathsForType()`, `setDisabledPathsForType()` | ~200 |
  | `settings_view.go` | `View()`, `HelpBindings()`, `StatusHelpText()`, rendering helpers | ~150 |

  - All types remain in the `views` package — no sub-packages
  - All unexported methods/types remain accessible within the package
  - Verify `settings_test.go` still compiles and passes after the split

  **Must NOT do**:
  - Do NOT create sub-packages (e.g., `views/settings/`)
  - Do NOT change any method signatures or types
  - Do NOT rename any functions
  - Do NOT modify `settings_test.go` unless compilation requires it

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`bubbletea`]
    - `bubbletea`: BubbleTea model structure, method organization

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 13)
  - **Blocks**: None
  - **Blocked By**: Task 8 (CQ-008 — togglePathDisabled must be extracted first)

  **References**:

  **Source File** (to split):
  - `internal/tui/views/settings.go` — 857 lines, 6 states, all methods on `SettingsModel`

  **Test File** (must stay working):
  - `internal/tui/views/settings_test.go` — 8 test functions

  **Pattern Reference** (existing file splits in codebase):
  - `internal/tui/views/restore.go` + `restore_fileselect.go` + `restore_diff.go` — Example of split pattern (main file + component files)

  **Acceptance Criteria**:
  - [ ] `settings.go` reduced to ~200 lines (core model + Update)
  - [ ] `settings_editing.go` created with editing handlers
  - [ ] `settings_paths.go` created with path browsing handlers
  - [ ] `settings_view.go` created with View() and rendering helpers
  - [ ] All 4 files in `views` package (no sub-packages)
  - [ ] `settings_test.go` unchanged and passing
  - [ ] `make test` → all PASS
  - [ ] `go vet ./...` → clean

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Settings split into 4 files, all compiling
    Tool: Bash
    Steps:
      1. ls internal/tui/views/settings*.go → Assert: settings.go, settings_editing.go, settings_paths.go, settings_view.go
      2. wc -l internal/tui/views/settings.go → Assert: < 250 lines
      3. grep "^package views" internal/tui/views/settings_editing.go → Assert: "package views"
      4. go test -v ./internal/tui/views/ -run TestSettings 2>&1 → Assert: PASS
      5. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: Clean 4-file split, all tests passing
    Evidence: ls, wc, grep, test output
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-002 split settings.go into 4 focused files`
  - Files: `internal/tui/views/settings.go`, `internal/tui/views/settings_editing.go`, `internal/tui/views/settings_paths.go`, `internal/tui/views/settings_view.go`
  - Pre-commit: `make test`

---

- [ ] 12. CQ-006: Wire FileSelector/DiffViewer into Restore (or Delete)

  **What to do**:
  - **Attempt 1 (preferred)**: Wire existing `FileSelector` and `DiffViewer` components into `RestoreModel`:
    - Replace inline `selectedFiles map[string]bool` + related methods with `FileSelector` composition
    - Replace inline `viewport` + `currentDiff` + `diffFile` with `DiffViewer` composition
    - Update `Update()` to delegate to `fileSelector.Update()` and `diffViewer.Update()` for relevant phases
    - Update render methods (from CQ-003) to call `fileSelector.View()` and `diffViewer.View()`
  - **Fallback (if >30 min)**: Delete both dead component files entirely:
    - Delete `internal/tui/views/restore_fileselect.go`
    - Delete `internal/tui/views/restore_diff.go`
  - Use the **fallback** if the components' API doesn't match the inline code's behavior 1:1

  **Must NOT do**:
  - Do NOT spend more than 30 minutes on wiring attempt — fall back to deletion
  - Do NOT change the restore workflow or user-facing behavior
  - Do NOT create new component files — use or delete the existing ones

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`bubbletea`]
    - `bubbletea`: BubbleTea component composition, Model delegation patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 11, 13)
  - **Blocks**: None
  - **Blocked By**: Task 7 (CQ-003 — render methods make wiring easier)

  **References**:

  **Inline Code** (to replace or keep):
  - `internal/tui/views/restore.go` — `selectedFiles`, `updateFileListSelection()`, `countSelectedFiles()`, `getSelectedFilePaths()`
  - `internal/tui/views/restore.go` — `viewport`, `currentDiff`, `diffFile` fields

  **Component Files** (to wire or delete):
  - `internal/tui/views/restore_fileselect.go` (147 lines) — `FileSelector` with `selected`, `SelectedCount()`, `SelectedFiles()`
  - `internal/tui/views/restore_diff.go` (150 lines) — `DiffViewer` with viewport management

  **Test Reference**:
  - `internal/tui/views/restore_test.go` — Tests for restore workflow phases

  **Acceptance Criteria**:
  - [ ] Either: Components wired in and inline duplicates removed, OR both component files deleted
  - [ ] Zero dead component code remaining (no unused `FileSelector` or `DiffViewer`)
  - [ ] Restore workflow unchanged (same phases, same behavior)
  - [ ] `go test -v ./internal/tui/views/ -run TestRestore` → PASS
  - [ ] `make test` → all PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: No dead component code remains
    Tool: Bash
    Steps:
      1. IF wired: grep -rn "FileSelector\|DiffViewer" internal/tui/views/restore.go → Assert: composition fields exist
      2. IF deleted: ls internal/tui/views/restore_fileselect.go 2>&1 → Assert: "No such file"
      3. go test -v ./internal/tui/views/ -run TestRestore 2>&1 → Assert: PASS
      4. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: Dead code eliminated one way or another
    Evidence: ls/grep and test output
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-006 wire FileSelector/DiffViewer into restore` OR `refactor(tui): CQ-006 remove unused FileSelector and DiffViewer components`
  - Files: `internal/tui/views/restore.go`, `internal/tui/views/restore_fileselect.go`, `internal/tui/views/restore_diff.go`
  - Pre-commit: `make test`

---

- [ ] 13. CQ-012: Add Styles to ProgramContext

  **What to do**:
  - Add a `Styles` field to the `ProgramContext` struct (wherever it's defined — likely `internal/tui/model.go` or `internal/tui/context.go`)
  - Populate it in the constructor: `Styles: styles.DefaultStyles()`
  - In all 6 view constructors, store the styles from context (or access via `m.ctx.Styles`)
  - Replace all `styles.DefaultStyles()` calls in View() methods with `m.ctx.Styles` (or a local field)
  - Files to update:
    - `internal/tui/views/settings.go` (line 733)
    - `internal/tui/views/restore.go` (line 534)
    - `internal/tui/views/backuplist.go` (line 256)
    - `internal/tui/views/dashboard.go` (line 100)
    - `internal/tui/views/logs.go` (line 182)
    - `internal/tui/views/setup.go` (line 389)
    - `internal/tui/views/helpers.go` (line 156)
    - `internal/tui/view.go` (line 38)
    - `internal/tui/help.go` (line 69)

  **Must NOT do**:
  - Do NOT add theme switching infrastructure
  - Do NOT modify `styles.DefaultStyles()` function itself
  - Do NOT change the actual style values

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: [`bubbletea`]
    - `bubbletea`: BubbleTea model context, shared state patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 11, 12)
  - **Blocks**: None
  - **Blocked By**: Wave 2 complete

  **References**:

  **ProgramContext Definition**:
  - `internal/tui/model.go` or similar — Where `ProgramContext` struct is defined
  - `internal/tui/styles/styles.go` — `DefaultStyles()` function and `Styles` type

  **All Call Sites** (9 total):
  - `internal/tui/views/settings.go:733`
  - `internal/tui/views/restore.go:534`
  - `internal/tui/views/backuplist.go:256`
  - `internal/tui/views/dashboard.go:100`
  - `internal/tui/views/logs.go:182`
  - `internal/tui/views/setup.go:389`
  - `internal/tui/views/helpers.go:156`
  - `internal/tui/view.go:38`
  - `internal/tui/help.go:69`

  **Acceptance Criteria**:
  - [ ] `ProgramContext` has `Styles` field
  - [ ] `DefaultStyles()` called once in constructor
  - [ ] All 9 View() call sites use cached styles
  - [ ] `make test` → all PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Styles cached in ProgramContext
    Tool: Bash
    Steps:
      1. grep "Styles" internal/tui/model.go → Assert: field in ProgramContext
      2. grep -rn "DefaultStyles()" internal/tui/ internal/tui/views/ | wc -l → Assert: 1-2 (only in constructor/init)
      3. make test 2>&1 | grep FAIL → Assert: no output
    Expected Result: Single point of styles initialization
    Evidence: grep output
  ```

  **Commit**: YES
  - Message: `refactor(tui): CQ-012 cache DefaultStyles in ProgramContext, eliminate per-render calls`
  - Files: `internal/tui/model.go`, `internal/tui/view.go`, `internal/tui/help.go`, `internal/tui/views/dashboard.go`, `internal/tui/views/backuplist.go`, `internal/tui/views/restore.go`, `internal/tui/views/settings.go`, `internal/tui/views/logs.go`, `internal/tui/views/setup.go`, `internal/tui/views/helpers.go`
  - Pre-commit: `make test`

---

- [ ] 14. Final Verification & Cleanup

  **What to do**:
  - Run full test suite: `make test`
  - Run vet: `go vet ./...`
  - Run build: `make build`
  - Compare passing packages against baseline: `make test 2>&1 | grep "^ok" | wc -l` must match `/tmp/dotkeeper-test-baseline.txt`
  - Verify all 14 commits are present: `git log --oneline refactor/code-quality ^master`
  - Verify no untracked files: `git status`

  **Must NOT do**:
  - Do NOT push to remote (user decides when to push/merge)
  - Do NOT squash commits

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential (must be last)
  - **Blocks**: None
  - **Blocked By**: ALL previous tasks

  **References**: None (verification only)

  **Acceptance Criteria**:
  - [ ] `make test` → all PASS, package count matches baseline
  - [ ] `go vet ./...` → clean
  - [ ] `make build` → success
  - [ ] 14 CQ commits present on branch (+ 1 setup commit = 15 total)
  - [ ] Working tree clean

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Full verification passes
    Tool: Bash
    Steps:
      1. make test 2>&1 | grep "^ok" | wc -l → Assert: matches baseline
      2. make test 2>&1 | grep FAIL → Assert: no output
      3. go vet ./... 2>&1 → Assert: no output
      4. make build 2>&1 → Assert: success
      5. git log --oneline refactor/code-quality ^master | wc -l → Assert: 15
      6. git status → Assert: "working tree clean"
    Expected Result: All verification passes
    Evidence: Command outputs

  Scenario: No dead code remains
    Tool: Bash
    Steps:
      1. grep -rn "func formatSize\|func formatBytes" internal/ → Assert: no output
      2. grep -rn "SuccessMsg\|LoadingMsg\|RefreshMsg" internal/tui/views/messages.go → Assert: no output
      3. grep -rn "backupDeleteErrorMsg\|diffErrorMsg\|restoreErrorMsg" internal/ → Assert: no output
    Expected Result: All dead code eliminated
    Evidence: grep output
  ```

  **Commit**: NO (verification only)

---

## Commit Strategy

| After Task | Message | Key Files | Verification |
|------------|---------|-----------|-------------|
| 0 | `chore: clean working tree before code quality refactoring` | modified/deleted docs | `make test` |
| 1 | `refactor: CQ-001 deduplicate size formatting to pathutil.FormatSize` | scanner.go, list.go, history.go, logs.go, setup.go | `make test` |
| 2 | `refactor(tui): CQ-005 remove dead message types` | messages.go | `make test` |
| 3 | `refactor(tui): CQ-007 remove dead error type aliases and inline constructors` | backuplist.go, restore.go | `make test` |
| 4 | `refactor(tui): CQ-010 extract maxPasswordAttempts constant` | restore.go | `make test` |
| 5 | `refactor(cli): CQ-011 use strings.TrimSuffix for password newline trimming` | backup.go | `make test` |
| 6 | `refactor(tui): CQ-004 add generic updateView helper` | update.go | `make test` |
| 7 | `refactor(tui): CQ-003 extract restore View() into 6 render methods` | restore.go | `make test` |
| 8 | `refactor(tui): CQ-008 extract togglePathDisabled method` | settings.go | `make test` |
| 9 | `refactor(cli): CQ-014 extract logHistory helper` | backup.go, restore.go | `make test` |
| 10 | `refactor(tui): CQ-013 extract shared renderPresetList` | setup.go | `make test` |
| 11 | `refactor(tui): CQ-002 split settings.go into 4 focused files` | settings*.go | `make test` |
| 12 | `refactor(tui): CQ-006 wire/remove FileSelector and DiffViewer` | restore*.go | `make test` |
| 13 | `refactor(tui): CQ-012 cache DefaultStyles in ProgramContext` | model.go, views/*.go | `make test` |

---

## Success Criteria

### Verification Commands
```bash
make test 2>&1 | grep "^ok" | wc -l   # Expected: matches baseline count
make test 2>&1 | grep FAIL             # Expected: no output
go vet ./...                           # Expected: no output
make build                             # Expected: success
git log --oneline refactor/code-quality ^master | wc -l  # Expected: 15
```

### Final Checklist
- [ ] All 14 CQ items implemented
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass with race detector
- [ ] Build succeeds
- [ ] Zero behavior changes
- [ ] ~400 lines of dead code removed
- [ ] ~35 lines of duplication eliminated
- [ ] settings.go split into 4 files (each < 250 lines)
- [ ] Generic updateView[T] helper in place
- [ ] FormatSize handles TB+ sizes
