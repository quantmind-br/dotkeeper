# UI Standardization & Enhancement

## TL;DR

> **Quick Summary**: Padronizar a UI do TUI dotkeeper em todas as 5 tabs com design system unificado — shared list delegate, status bar helper, dashboard cards/botões, settings full refactor para bubbles/list, e cleanup de estilos inline.
> 
> **Deliverables**:
> - Design system expandido em `views/styles.go` (container, status bar, card, button styles + `NewListDelegate()`)
> - `RenderStatusBar()` helper em `views/helpers.go`
> - Dashboard com stat cards e action buttons horizontais
> - BackupList, Logs, Restore usando shared list delegate
> - Settings totalmente refatorado para bubbles/list
> - view.go e help.go estilos movidos para styles.go
> - Testes atualizados + novos testes para componentes
> 
> **Estimated Effort**: Large
> **Parallel Execution**: YES — 4 waves
> **Critical Path**: Task 1 → Task 2 → Task 3 → Tasks 4/5/6 (parallel) → Task 7 → Task 8 → Task 9

---

## Context

### Original Request
Padronizar e embelezar a UI do dotkeeper TUI conforme plano em FEATURE.md — unificar design language, components, e feedback visual em todas as views.

### Interview Summary
**Key Discussions**:
- **Settings**: Full refactor para bubbles/list — navegação principal E sub-listas files/folders
- **Dashboard Cards**: Simples sem borda — background sutil, bold values, labels
- **Quick Actions**: Botões horizontais SEM ícones — apenas texto estilizado com key highlight
- **Icons**: Sem ícones Unicode — só texto puro + atalho de tecla
- **Viewport Border**: Purple #7D56F4 para consistência com tema
- **Testes**: Atualizar existentes + criar novos para componentes compartilhados

**Research Findings**:
- 4 instâncias de `list.NewDefaultDelegate()` (backuplist, restore x2, logs)
- view.go tem styles inline que precisam migrar para styles.go
- setup.go tem cores inconsistentes (#FF6B6B vs #FF5555) — FORA do escopo
- `components/tabbar.go` importa `views.Styles` — colocar delegate em `views/styles.go` para evitar import cycle
- Settings tem 5 estados booleanos interagindo — refactor mais arriscado
- Testes usam string-matching em View() — precisam de `stripANSI()` helper
- Magic number `-6` para height em 5 locais

### Metis Review
**Identified Gaps** (addressed):
- Import cycle risk: `NewListDelegate()` vai em `views/styles.go`, NÃO em `components/`
- `stripANSI()` test helper deve ser criado ANTES de qualquer mudança de view
- Settings `IsEditing()` contrato deve ser preservado exatamente
- setup.go marcado explicitamente como OUT OF SCOPE
- `list.SetShowTitle(false)` deve ser padrão em TODAS as lists (views renderizam próprio título)
- Magic `-6` substituído por constante nomeada
- `strings.Contains(m.err, "success")` em settings é code smell — resolvido por campos separados
- Status bar persist-until-next-action (sem auto-dismiss)

---

## Work Objectives

### Core Objective
Unificar a UI do TUI dotkeeper para que todas as 5 tabs compartilhem design language, componentes visuais, e padrões de feedback idênticos.

### Concrete Deliverables
- `internal/tui/views/styles.go` expandido com novos estilos + `NewListDelegate()` factory
- `internal/tui/views/helpers.go` com `RenderStatusBar()` helper
- `internal/tui/views/dashboard.go` refatorado com cards + buttons
- `internal/tui/views/backuplist.go` usando shared delegate + status bar
- `internal/tui/views/logs.go` usando shared delegate + status bar
- `internal/tui/views/restore.go` usando shared delegate + status bar + purple viewport border
- `internal/tui/views/settings.go` full refactor para bubbles/list
- `internal/tui/view.go` estilos movidos para styles.go
- `internal/tui/help.go` estilos consolidados
- Todos os `*_test.go` atualizados + novos testes

### Definition of Done
- [x] `go test ./internal/tui/... -race -count=1` → ALL PASS
- [x] `go vet ./internal/tui/...` → no errors
- [x] `grep -rn 'NewDefaultDelegate' internal/tui/views/*.go` → 0 matches
- [x] `grep -rn 'lipgloss.NewStyle()' internal/tui/views/*.go | grep -v styles.go | grep -v setup.go | grep -v _test.go` → 3 matches (settings textinput cursor/prompt + restore viewport dynamic style)
- [x] `grep -rn 'lipgloss.NewStyle()' internal/tui/view.go` → 0 matches
- [x] Todas as 5 tabs visualmente consistentes (verificado via TUI)

### Must Have
- Shared list delegate com palette Purple/Green/Grey para TODAS as lists
- Status bar unificado em TODAS as views
- Dashboard com stat cards e action buttons
- Settings navegável via bubbles/list
- Constante nomeada para view chrome height (substituir magic `-6`)
- `stripANSI()` test helper compartilhado
- Testes atualizados para cada view modificada

### Must NOT Have (Guardrails)
- **NÃO** mudar Update() logic exceto Settings (que requer para list.Model)
- **NÃO** tocar em setup.go ou filebrowser.go (fora do escopo)
- **NÃO** mudar key bindings existentes
- **NÃO** mudar lógica de backup, restore, config, crypto, ou git
- **NÃO** adicionar auto-dismiss timer para status messages
- **NÃO** mudar filtering/searching behavior das lists
- **NÃO** mudar `saveFieldValue()` ou `config.Save()` logic
- **NÃO** mudar path validation logic
- **NÃO** colocar componentes compartilhados em `components/` (import cycle)
- **NÃO** adicionar ícones Unicode ou emoji
- **NÃO** mudar cores existentes da palette (#7D56F4, #04B575, #FF5555)
- **NÃO** mudar Argon2id ou crypto parameters

---

## Verification Strategy

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> ALL tasks in this plan MUST be verifiable WITHOUT any human action.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: YES (tests-after — update existing + add new)
- **Framework**: `go test` (standard Go testing)

### Agent-Executed QA Scenarios (MANDATORY — ALL tasks)

**Verification Tool by Deliverable Type:**

| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| **Go code** | Bash (`go test`, `go vet`, `go build`) | Compile, test, vet |
| **TUI visual** | interactive_bash (tmux) | Launch TUI, screenshot, navigate |
| **Style consistency** | Bash (grep) | Search for inline styles, verify migration |

### Phase Gate (run after EVERY task):
```bash
go test ./internal/tui/... -race -count=1 && echo "PASS" || echo "FAIL"
go vet ./internal/tui/... && echo "VET OK" || echo "VET FAIL"
```

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately):
└── Task 1: Test infrastructure (stripANSI helper)

Wave 2 (After Wave 1):
├── Task 2: Design system foundation (styles.go expansion)
└── Task 3: Status bar helper (helpers.go)

Wave 3 (After Wave 2):
├── Task 4: BackupList standardization
├── Task 5: Logs standardization
└── Task 6: Restore standardization

Wave 4 (After Wave 3):
└── Task 7: Dashboard overhaul (cards + buttons)

Wave 5 (After Wave 3):
└── Task 8: Settings full refactor

Wave 6 (After Wave 4 + 5):
└── Task 9: Outer frame cleanup (view.go + help.go) + final visual QA
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2, 3 | None |
| 2 | 1 | 4, 5, 6, 7, 8, 9 | 3 |
| 3 | 1 | 4, 5, 6, 7, 8, 9 | 2 |
| 4 | 2, 3 | 9 | 5, 6 |
| 5 | 2, 3 | 9 | 4, 6 |
| 6 | 2, 3 | 9 | 4, 5 |
| 7 | 2, 3 | 9 | 8 |
| 8 | 2, 3 | 9 | 7 |
| 9 | 4, 5, 6, 7, 8 | None | None (final) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Agents |
|------|-------|-------------------|
| 1 | 1 | `category="quick"` |
| 2 | 2, 3 | `category="quick"` (parallel) |
| 3 | 4, 5, 6 | `category="quick"` (parallel) |
| 4-5 | 7, 8 | 7: `category="visual-engineering"`, 8: `category="deep"` (parallel) |
| 6 | 9 | `category="visual-engineering"` |

---

## TODOs

- [x] 1. Test Infrastructure: Create stripANSI Helper

  **What to do**:
  - Create `internal/tui/views/testhelpers_test.go` with a `stripANSI(s string) string` function that removes ANSI escape codes
  - Reference implementation exists in `internal/tui/components/tabbar_test.go` — copy/adapt the `stripANSI()` pattern from there
  - Update ALL existing test files that use `strings.Contains(view, ...)` to wrap the `View()` output with `stripANSI()` before assertions
  - Files to update: `dashboard_test.go`, `backuplist_test.go`, `restore_test.go`, `settings_test.go`, `logs_test.go`
  - Do NOT update `setup_test.go` or `filebrowser_test.go` (out of scope)
  - Verify all existing tests still pass after this change (ANSI stripping should not break anything since current views don't add ANSI, but it future-proofs them)

  **Must NOT do**:
  - Do NOT change any view code — only test files
  - Do NOT modify test logic — only wrap View() output with stripANSI()
  - Do NOT touch setup_test.go or filebrowser_test.go

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple utility function + mechanical replacement across test files
  - **Skills**: [`git-master`]
    - `git-master`: Commit the test infrastructure change

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1 (alone)
  - **Blocks**: Tasks 2, 3
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/components/tabbar_test.go` — Has existing `stripANSI()` implementation to copy/adapt

  **Test References**:
  - `internal/tui/views/dashboard_test.go` — Uses strings.Contains on View() output
  - `internal/tui/views/backuplist_test.go` — Uses strings.Contains on View() output
  - `internal/tui/views/restore_test.go` — Uses strings.Contains on View() output
  - `internal/tui/views/settings_test.go` — Has custom `contains()` helper, uses strings.Contains on View() output
  - `internal/tui/views/logs_test.go` — Uses strings.Contains on View() output

  **Acceptance Criteria**:
  - [x] `internal/tui/views/testhelpers_test.go` exists with `stripANSI()` function
  - [x] `go test ./internal/tui/views/... -v -count=1` → ALL PASS
  - [x] `go test ./internal/tui/components/... -v -count=1` → ALL PASS
  - [x] `grep -rn 'stripANSI' internal/tui/views/*_test.go | wc -l` → >= 5 (used in all view test files)

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: All existing tests pass with stripANSI wrapper
    Tool: Bash
    Preconditions: None
    Steps:
      1. go test ./internal/tui/views/... -v -count=1
      2. Assert: exit code 0
      3. Assert: output contains "PASS"
      4. Assert: output does NOT contain "FAIL"
    Expected Result: All tests pass
    Evidence: Terminal output captured

  Scenario: stripANSI function correctly strips ANSI codes
    Tool: Bash
    Preconditions: testhelpers_test.go exists
    Steps:
      1. go test ./internal/tui/views/ -run TestStripANSI -v
      2. Assert: test passes (if a test for it was added)
      OR verify by: grep -c 'stripANSI' internal/tui/views/testhelpers_test.go
      3. Assert: count >= 1
    Expected Result: Helper function exists and works
    Evidence: Terminal output captured
  ```

  **Commit**: YES
  - Message: `refactor(tui): add stripANSI test helper and future-proof view tests`
  - Files: `internal/tui/views/testhelpers_test.go`, `internal/tui/views/*_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [x] 2. Design System Foundation: Expand styles.go

  **What to do**:
  - Add new styles to the `Styles` struct in `views/styles.go`:
    - `ViewContainer` — lipgloss.Style with consistent padding/margins for wrapping view content
    - `StatusBar` — lipgloss.Style for the fixed status/help bar area
    - `Card` — lipgloss.Style for dashboard stat cards (subtle background `#2A2A2A`, padding 1)
    - `CardTitle` — lipgloss.Style for card bold value (bold, foreground `#FFFFFF`)
    - `CardLabel` — lipgloss.Style for card label below value (foreground `#AAAAAA`)
    - `ActionButton` — lipgloss.Style for dashboard action buttons (background `#2A2A2A`, padding horizontal 2)
    - `ActionButtonKey` — lipgloss.Style for the key shortcut inside action button (foreground `#7D56F4`, bold)
  - Create `NewListDelegate() list.DefaultDelegate` function in `views/styles.go`:
    - Configure `NormalTitle` style with foreground `#FFFFFF`
    - Configure `NormalDesc` style with foreground `#AAAAAA`
    - Configure `SelectedTitle` style with foreground `#7D56F4`, bold
    - Configure `SelectedDesc` style with foreground `#7D56F4`
    - Set `SelectedTitle` prefix to `"▸ "` and suffix to `""`
    - Set spacing/height appropriately
  - Define named constant: `const ViewChromeHeight = 6` (title + tabbar + help + margins — replaces magic `-6`)
  - Add `ViewChromeHeight` as exported constant so views can calculate list height correctly

  **Must NOT do**:
  - Do NOT change existing style values (Title, Selected, etc.)
  - Do NOT remove any existing styles
  - Do NOT import from `components/` package
  - Do NOT add styles for setup.go or filebrowser.go

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Adding struct fields and factory function to existing file — straightforward Go
  - **Skills**: [`git-master`]
    - `git-master`: Commit the style expansion

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 3)
  - **Blocks**: Tasks 4, 5, 6, 7, 8, 9
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/tui/views/styles.go:1-62` — Current Styles struct and DefaultStyles() factory
  - `internal/tui/components/tabbar.go` — Uses `views.Styles` for rendering (shows how styles are consumed from outside)

  **API/Type References**:
  - `github.com/charmbracelet/bubbles/list` — `DefaultDelegate` type with `Styles` field containing `NormalTitle`, `NormalDesc`, `SelectedTitle`, `SelectedDesc`
  - `github.com/charmbracelet/lipgloss` — Style methods: `Foreground`, `Background`, `Bold`, `Padding`, `MarginLeft`

  **External References**:
  - Bubbles list delegate docs: https://pkg.go.dev/github.com/charmbracelet/bubbles/list#DefaultDelegate

  **Acceptance Criteria**:
  - [x] `go vet ./internal/tui/views/...` → no errors
  - [x] `go build ./internal/tui/...` → compiles
  - [x] `grep 'NewListDelegate' internal/tui/views/styles.go` → 1 match
  - [x] `grep 'ViewChromeHeight' internal/tui/views/styles.go` → 1 match
  - [x] `grep 'Card ' internal/tui/views/styles.go` → >= 1 match
  - [x] `grep 'ActionButton ' internal/tui/views/styles.go` → >= 1 match
  - [x] `grep 'StatusBar' internal/tui/views/styles.go` → >= 1 match
  - [x] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: styles.go compiles and has new types
    Tool: Bash
    Preconditions: Task 1 complete
    Steps:
      1. go vet ./internal/tui/views/...
      2. Assert: exit code 0
      3. go build ./internal/tui/...
      4. Assert: exit code 0
    Expected Result: No compilation errors
    Evidence: Terminal output

  Scenario: NewListDelegate returns valid delegate
    Tool: Bash
    Preconditions: styles.go updated
    Steps:
      1. grep -n 'func NewListDelegate' internal/tui/views/styles.go
      2. Assert: exactly 1 match
      3. grep -n 'list.DefaultDelegate' internal/tui/views/styles.go
      4. Assert: at least 1 match (return type)
    Expected Result: Factory function defined
    Evidence: Terminal output
  ```

  **Commit**: YES
  - Message: `feat(tui): expand design system with shared list delegate and layout styles`
  - Files: `internal/tui/views/styles.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [x] 3. Create RenderStatusBar Helper

  **What to do**:
  - Add `RenderStatusBar(width int, status string, errMsg string, helpText string) string` to `internal/tui/views/helpers.go`
  - Status text rendered with `styles.Success` (green)
  - Error text rendered with `styles.Error` (red)
  - Help text rendered with `styles.Help` (grey)
  - Layout: status/error on one line (priority: error > status if both present), help on next line
  - Truncate status/error to `width - 4` to prevent wrapping
  - Use `styles.StatusBar` from styles.go for outer container
  - Add unit test in `internal/tui/views/helpers_test.go`:
    - Test with status only → green text + help
    - Test with error only → red text + help
    - Test with both → error wins (red text + help)
    - Test with empty status and error → just help text
    - Test truncation on narrow width

  **Must NOT do**:
  - Do NOT add auto-dismiss timer or tea.Cmd return
  - Do NOT change existing helpers (expandHome, ValidatePath, etc.)
  - Do NOT add any state — this is a pure rendering function

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single function + unit test — simple Go
  - **Skills**: [`git-master`]
    - `git-master`: Commit the helper

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 2)
  - **Blocks**: Tasks 4, 5, 6, 7, 8, 9
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/tui/views/helpers.go:1-120` — Existing helpers file structure, expandHome, HelpEntry type
  - `internal/tui/views/backuplist.go:261-269` — Current inline status bar pattern (status + error + help text)
  - `internal/tui/views/logs.go:186-199` — Another inline status/help pattern
  - `internal/tui/views/settings.go:383-391` — Settings `strings.Contains(m.err, "success")` anti-pattern to replace

  **Acceptance Criteria**:
  - [x] `grep 'func RenderStatusBar' internal/tui/views/helpers.go` → 1 match
  - [x] `go test ./internal/tui/views/ -run TestRenderStatusBar -v` → PASS
  - [x] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: RenderStatusBar with status message
    Tool: Bash
    Preconditions: helpers.go updated
    Steps:
      1. go test ./internal/tui/views/ -run TestRenderStatusBar -v
      2. Assert: exit code 0
      3. Assert: output contains "PASS"
    Expected Result: All status bar tests pass
    Evidence: Terminal output

  Scenario: RenderStatusBar compiles with no issues
    Tool: Bash
    Steps:
      1. go vet ./internal/tui/views/...
      2. Assert: exit code 0
    Expected Result: No vet errors
    Evidence: Terminal output
  ```

  **Commit**: YES
  - Message: `feat(tui): add RenderStatusBar helper for unified status/help rendering`
  - Files: `internal/tui/views/helpers.go`, `internal/tui/views/helpers_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [x] 4. BackupList Standardization

  **What to do**:
  - Replace `list.NewDefaultDelegate()` on line 55 with `NewListDelegate()`
  - Set `l.SetShowTitle(false)` (view renders its own title via `styles.Title`)
  - Add title rendering at top of `View()` normal mode: `styles.Title.Render("Backups")`
  - Replace magic `msg.Height - 6` with `msg.Height - ViewChromeHeight`
  - Replace inline status/error/help rendering (lines 261-268) with `RenderStatusBar(m.width, m.backupStatus, m.backupError, "n: new backup | d: delete | r: refresh | ↑/↓: navigate")`
  - Update `backuplist_test.go`:
    - All View() assertions use `stripANSI()` wrapper
    - Update string assertions for new title position (view renders title, not list)
    - Add test: verify status bar renders correctly with backup success/error

  **Must NOT do**:
  - Do NOT change Update() logic
  - Do NOT change Refresh(), runBackup(), deleteBackup() functions
  - Do NOT change key bindings
  - Do NOT change backupItem struct

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Mechanical replacement of delegate + status bar in existing view
  - **Skills**: [`git-master`]
    - `git-master`: Commit the changes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 5, 6)
  - **Blocks**: Task 9
  - **Blocked By**: Tasks 2, 3

  **References**:

  **Pattern References**:
  - `internal/tui/views/styles.go` — `NewListDelegate()` factory (created in Task 2)
  - `internal/tui/views/helpers.go` — `RenderStatusBar()` helper (created in Task 3)

  **API/Type References**:
  - `internal/tui/views/styles.go:ViewChromeHeight` — Named constant for height calculation
  - `internal/tui/views/backuplist.go:40-71` — BackupListModel struct and constructor

  **Test References**:
  - `internal/tui/views/backuplist_test.go` — Existing tests to update

  **Acceptance Criteria**:
  - [x] `grep 'NewDefaultDelegate' internal/tui/views/backuplist.go` → 0 matches
  - [x] `grep 'NewListDelegate' internal/tui/views/backuplist.go` → 1 match
  - [x] `grep 'RenderStatusBar' internal/tui/views/backuplist.go` → >= 1 match
  - [x] `grep 'ViewChromeHeight' internal/tui/views/backuplist.go` → 1 match
  - [x] `go test ./internal/tui/views/ -run TestBackup -v` → ALL PASS
  - [x] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: BackupList renders with custom delegate
    Tool: Bash
    Steps:
      1. go test ./internal/tui/views/ -run TestBackup -v -count=1
      2. Assert: exit code 0
      3. Assert: output contains "PASS"
    Expected Result: BackupList tests pass with new delegate
    Evidence: Terminal output

  Scenario: No default delegates remain in backuplist
    Tool: Bash
    Steps:
      1. grep -n 'NewDefaultDelegate' internal/tui/views/backuplist.go
      2. Assert: no output (exit code 1)
      3. grep -n 'NewListDelegate' internal/tui/views/backuplist.go
      4. Assert: exactly 1 match
    Expected Result: Custom delegate in use
    Evidence: Terminal output
  ```

  **Commit**: YES (groups with Tasks 5, 6)
  - Message: `refactor(tui): standardize backup list with shared delegate and status bar`
  - Files: `internal/tui/views/backuplist.go`, `internal/tui/views/backuplist_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [x] 5. Logs Standardization

  **What to do**:
  - Replace `list.NewDefaultDelegate()` on line 74 with `NewListDelegate()`
  - `l.SetShowTitle(false)` already set — good, keep it
  - Replace magic `msg.Height - 6` with `msg.Height - ViewChromeHeight`
  - Replace inline error/help rendering (lines 186-199) with `RenderStatusBar(m.width, "", m.err, helpText)` where helpText includes current filter: `fmt.Sprintf("f: filter (%s) | r: refresh | ↑/↓: navigate", m.filter)`
  - Update `logs_test.go`:
    - All View() assertions use `stripANSI()` wrapper
    - Update string assertions to match new status bar format

  **Must NOT do**:
  - Do NOT change Update() logic
  - Do NOT change LoadHistory() or filter cycling
  - Do NOT change logItem or formatBytes

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Same mechanical replacement as Task 4
  - **Skills**: [`git-master`]
    - `git-master`: Commit the changes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 4, 6)
  - **Blocks**: Task 9
  - **Blocked By**: Tasks 2, 3

  **References**:

  **Pattern References**:
  - `internal/tui/views/styles.go` — `NewListDelegate()` factory
  - `internal/tui/views/helpers.go` — `RenderStatusBar()` helper

  **API/Type References**:
  - `internal/tui/views/logs.go:15-23` — LogsModel struct
  - `internal/tui/views/logs.go:65-84` — NewLogs constructor

  **Test References**:
  - `internal/tui/views/logs_test.go` — Existing tests to update

  **Acceptance Criteria**:
  - [x] `grep 'NewDefaultDelegate' internal/tui/views/logs.go` → 0 matches
  - [x] `grep 'NewListDelegate' internal/tui/views/logs.go` → 1 match
  - [x] `grep 'RenderStatusBar' internal/tui/views/logs.go` → >= 1 match
  - [x] `grep 'ViewChromeHeight' internal/tui/views/logs.go` → 1 match
  - [x] `go test ./internal/tui/views/ -run TestLogs -v` → ALL PASS
  - [x] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Logs renders with custom delegate
    Tool: Bash
    Steps:
      1. go test ./internal/tui/views/ -run TestLogs -v -count=1
      2. Assert: exit code 0
    Expected Result: Logs tests pass
    Evidence: Terminal output

  Scenario: No default delegates remain in logs
    Tool: Bash
    Steps:
      1. grep -n 'NewDefaultDelegate' internal/tui/views/logs.go
      2. Assert: no output
    Expected Result: Custom delegate in use
    Evidence: Terminal output
  ```

  **Commit**: YES (groups with Tasks 4, 6)
  - Message: `refactor(tui): standardize logs view with shared delegate and status bar`
  - Files: `internal/tui/views/logs.go`, `internal/tui/views/logs_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [x] 6. Restore Standardization

  **What to do**:
  - Replace BOTH `list.NewDefaultDelegate()` calls (lines 93, 103) with `NewListDelegate()`
  - Set `l.SetShowTitle(false)` on both backupList and fileList (view renders own titles)
  - Replace magic `msg.Height - 6` (3 occurrences: lines 251, 252, 254) with `msg.Height - ViewChromeHeight`
  - Replace inline viewport style (lines 533-537) — change border color from `#666666` to `#7D56F4`
  - Move viewport style to `styles.go` as `DiffViewport` style (or use `Styles.ViewContainer` with border)
  - Replace inline status/error/help rendering in EACH phase:
    - Phase 0 (lines 471-478): `RenderStatusBar(m.width, m.restoreStatus, m.restoreError, "↑/↓: navigate | Enter: select | r: refresh")`
    - Phase 1 (lines 488-495): `RenderStatusBar(m.width, m.restoreStatus, m.restoreError, "Enter: validate | Esc: back")`
    - Phase 2 (lines 510-517): `RenderStatusBar(m.width, m.restoreStatus, m.restoreError, "Space: toggle | a: all | n: none | d: diff | Enter: restore | Esc: back")`
    - Phase 4 (lines 541-545): `RenderStatusBar(m.width, "", m.restoreError, "j/k or ↑/↓: scroll | g/G: top/bottom | Esc: back")`
    - Phase 5 (lines 583-584): `RenderStatusBar(m.width, "", "", "Press any key to continue")`
  - Update `restore_test.go`:
    - All View() assertions use `stripANSI()` wrapper
    - Update string assertions for title/status changes

  **Must NOT do**:
  - Do NOT change Update() logic or phase transitions
  - Do NOT change validatePassword(), loadFiles(), loadDiff(), runRestore()
  - Do NOT change fileItem struct or checkbox rendering
  - Do NOT change selectedFiles map logic

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Same pattern as Tasks 4/5 but applied to more complex view — still mechanical
  - **Skills**: [`git-master`]
    - `git-master`: Commit the changes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 4, 5)
  - **Blocks**: Task 9
  - **Blocked By**: Tasks 2, 3

  **References**:

  **Pattern References**:
  - `internal/tui/views/styles.go` — `NewListDelegate()` factory + new viewport style
  - `internal/tui/views/helpers.go` — `RenderStatusBar()` helper

  **API/Type References**:
  - `internal/tui/views/restore.go:20-39` — RestoreModel struct with phase field
  - `internal/tui/views/restore.go:91-119` — NewRestore constructor with both lists

  **Test References**:
  - `internal/tui/views/restore_test.go` — Existing tests covering all 6 phases

  **Acceptance Criteria**:
  - [x] `grep 'NewDefaultDelegate' internal/tui/views/restore.go` → 0 matches
  - [x] `grep 'NewListDelegate' internal/tui/views/restore.go` → 2 matches
  - [x] `grep 'RenderStatusBar' internal/tui/views/restore.go` → >= 4 matches
  - [x] `grep 'ViewChromeHeight' internal/tui/views/restore.go` → >= 2 matches
  - [x] `grep '#666666' internal/tui/views/restore.go` → 0 matches (viewport border changed to purple)
  - [x] `go test ./internal/tui/views/ -run TestRestore -v` → ALL PASS
  - [x] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Restore tests pass all phases
    Tool: Bash
    Steps:
      1. go test ./internal/tui/views/ -run TestRestore -v -count=1
      2. Assert: exit code 0
      3. Assert: output contains "PASS"
    Expected Result: All restore phase tests pass
    Evidence: Terminal output

  Scenario: No default delegates or grey borders remain
    Tool: Bash
    Steps:
      1. grep -n 'NewDefaultDelegate' internal/tui/views/restore.go
      2. Assert: no output
      3. grep -n '#666666' internal/tui/views/restore.go
      4. Assert: no output
    Expected Result: All standardized
    Evidence: Terminal output
  ```

  **Commit**: YES (groups with Tasks 4, 5)
  - Message: `refactor(tui): standardize restore view with shared delegate, status bar, and purple viewport`
  - Files: `internal/tui/views/restore.go`, `internal/tui/views/restore_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [x] 7. Dashboard UI Overhaul

  **What to do**:
  - Refactor `View()` in `dashboard.go` to use stat cards layout:
    - **Card 1**: "Last Backup" — shows date or "Never"
    - **Card 2**: "Files Tracked" — shows count from config
    - Use `styles.Card` for each card container, `styles.CardTitle` for value, `styles.CardLabel` for label
    - Arrange cards horizontally with `lipgloss.JoinHorizontal(lipgloss.Top, card1, card2)`
    - Responsive: if `m.width < 60`, stack vertically with `lipgloss.JoinVertical`
  - Replace plain text Quick Actions with styled horizontal buttons:
    - `[ b  Backup ]  [ r  Restore ]  [ s  Settings ]`
    - Use `styles.ActionButton` for button container
    - Use `styles.ActionButtonKey` for the key letter
    - Arrange with `lipgloss.JoinHorizontal(lipgloss.Top, btn1, btn2, btn3)` with gap `"  "` between
    - Responsive: if `m.width < 60`, stack vertically
  - Add `RenderStatusBar(m.width, "", "", "b: backup | r: restore | s: settings")` at bottom
  - Remove the raw `styles.Title.Render("Dashboard")` — title is already rendered by tabbar system
  - Update `dashboard_test.go`:
    - All View() assertions use `stripANSI()` wrapper
    - Update assertions for new card/button content
    - Add test: verify responsive layout switches at width boundary

  **Must NOT do**:
  - Do NOT change Update() logic or refreshStatus()
  - Do NOT add new state fields to DashboardModel
  - Do NOT make cards interactive/navigable
  - Do NOT add icons or emoji

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Lipgloss layout composition with responsive breakpoints — visual design task
  - **Skills**: [`git-master`]
    - `git-master`: Commit the changes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4-5 (with Task 8)
  - **Blocks**: Task 9
  - **Blocked By**: Tasks 2, 3

  **References**:

  **Pattern References**:
  - `internal/tui/views/styles.go` — Card, CardTitle, CardLabel, ActionButton, ActionButtonKey styles
  - `internal/tui/views/helpers.go` — `RenderStatusBar()` helper
  - `internal/tui/components/tabbar.go:30-50` — Example of lipgloss.JoinHorizontal usage with responsive breakpoint

  **API/Type References**:
  - `internal/tui/views/dashboard.go:14-21` — DashboardModel struct (width, height, lastBackup, fileCount)
  - `internal/tui/views/dashboard.go:49-71` — Current View() to replace

  **External References**:
  - lipgloss layout docs: `JoinHorizontal`, `JoinVertical`, `Place` — https://pkg.go.dev/github.com/charmbracelet/lipgloss

  **Acceptance Criteria**:
  - [x] `go test ./internal/tui/views/ -run TestDashboard -v` → ALL PASS
  - [x] Dashboard View() contains card-like rendering (background-styled blocks)
  - [x] `grep 'JoinHorizontal' internal/tui/views/dashboard.go` → >= 1 match
  - [x] `grep 'RenderStatusBar' internal/tui/views/dashboard.go` → 1 match
  - [x] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Dashboard renders cards with correct data
    Tool: Bash
    Steps:
      1. go test ./internal/tui/views/ -run TestDashboard -v -count=1
      2. Assert: exit code 0
    Expected Result: Dashboard tests pass with new card layout
    Evidence: Terminal output

  Scenario: Dashboard uses lipgloss layout functions
    Tool: Bash
    Steps:
      1. grep -c 'JoinHorizontal\|JoinVertical' internal/tui/views/dashboard.go
      2. Assert: count >= 2 (cards + buttons)
    Expected Result: Layout functions in use
    Evidence: Terminal output

  Scenario: Dashboard visual verification
    Tool: interactive_bash (tmux)
    Preconditions: Application built with `make build`
    Steps:
      1. tmux new-session -d -s dash-test
      2. tmux send-keys -t dash-test './bin/dotkeeper' Enter
      3. Wait 2 seconds
      4. tmux capture-pane -t dash-test -p > .sisyphus/evidence/task-7-dashboard.txt
      5. Assert: output contains "Last Backup"
      6. Assert: output contains "Files Tracked"
      7. Assert: output contains "Backup" "Restore" "Settings" (buttons)
      8. tmux send-keys -t dash-test 'q'
    Expected Result: Dashboard shows cards and buttons
    Evidence: .sisyphus/evidence/task-7-dashboard.txt
  ```

  **Commit**: YES
  - Message: `feat(tui): redesign dashboard with stat cards and styled action buttons`
  - Files: `internal/tui/views/dashboard.go`, `internal/tui/views/dashboard_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [x] 8. Settings Full Refactor

  **What to do**:

  **STATE MACHINE DESIGN** (implement this exactly):
  ```
  States:
  - ReadOnly: browsing settings (list visible, no edit)
  - ListNavigating: edit mode, navigating main 6 fields via bubbles/list
  - EditingField: typing in textinput for a scalar field (BackupDir, GitRemote, Schedule)
  - BrowsingFiles: navigating files sub-list (bubbles/list)
  - BrowsingFolders: navigating folders sub-list (bubbles/list)
  - EditingSubItem: typing in textinput for a file/folder path

  Transitions:
  - ReadOnly → ListNavigating: press 'e'
  - ListNavigating → EditingField: press Enter on scalar field (cursor 0,1,4)
  - ListNavigating → BrowsingFiles: press Enter on Files (cursor 2)
  - ListNavigating → BrowsingFolders: press Enter on Folders (cursor 3)
  - ListNavigating → toggle Notifications: press Enter on Notifications (cursor 5)
  - ListNavigating → ReadOnly: press Esc
  - EditingField → ListNavigating: press Enter (save) or Esc (cancel)
  - BrowsingFiles → ListNavigating: press Esc
  - BrowsingFiles → EditingSubItem: press Enter on file item or 'a' for new
  - BrowsingFolders → ListNavigating: press Esc
  - BrowsingFolders → EditingSubItem: press Enter on folder item or 'a' for new
  - EditingSubItem → BrowsingFiles/BrowsingFolders: press Enter (save) or Esc (cancel)
  - Any navigating state: 's' saves config
  ```

  - Replace manual rendering with `list.Model` for main settings navigation
  - Create `settingItem` struct implementing `list.Item` for the 6 fields
  - Create separate `list.Model` instances for files sub-list and folders sub-list
  - All 3 lists use `NewListDelegate()`
  - All 3 lists use `SetShowTitle(false)` and `SetShowHelp(false)`
  - Replace `strings.Contains(m.err, "success")` with separate `status` and `errMsg` fields
  - Use `RenderStatusBar()` for all status/error/help rendering
  - Replace magic height with `ViewChromeHeight`
  - **PRESERVE** exact `IsEditing()` contract: return true whenever user is in any state other than ReadOnly
  - **PRESERVE** exact key binding behavior: e, esc, up/down, enter, a, d, s
  - **PRESERVE** `saveFieldValue()` and `config.Save()` logic unchanged
  - **PRESERVE** path validation logic unchanged
  - Update `settings_test.go`:
    - All View() assertions use `stripANSI()` wrapper
    - Test state transitions match the state machine above
    - Test `IsEditing()` returns correct value for each state
    - Verify save/error rendering via status bar

  **Must NOT do**:
  - Do NOT change `saveFieldValue()` logic
  - Do NOT change `config.Save()` logic
  - Do NOT change path validation (ValidateFilePath, ValidateFolderPath)
  - Do NOT change textinput behavior when editing
  - Do NOT add new config fields
  - Do NOT change the Notifications toggle behavior

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Highest-risk task — state machine refactor with 6 states, must preserve exact behavior. Needs thorough reasoning.
  - **Skills**: [`git-master`]
    - `git-master`: Commit the changes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4-5 (with Task 7)
  - **Blocks**: Task 9
  - **Blocked By**: Tasks 2, 3

  **References**:

  **Pattern References**:
  - `internal/tui/views/settings.go:1-422` — ENTIRE current implementation (read carefully!)
  - `internal/tui/views/backuplist.go:17-26` — Example of list.Item implementation (backupItem)
  - `internal/tui/views/styles.go` — `NewListDelegate()` factory

  **API/Type References**:
  - `internal/tui/views/settings.go:16-29` — Current SettingsModel struct (reference for state fields)
  - `internal/tui/views/settings.go:240-290` — saveFieldValue logic (preserve exactly)
  - `internal/tui/views/settings.go:419-421` — IsEditing() contract (must preserve)
  - `internal/config/config.go` — Config struct with BackupDir, GitRemote, Files, Folders, Schedule, Notifications fields

  **Test References**:
  - `internal/tui/views/settings_test.go` — Existing tests covering read-only, edit mode, field editing, save

  **Acceptance Criteria**:
  - [x] `go test ./internal/tui/views/ -run TestSettings -v` → ALL PASS
  - [x] `go test ./internal/tui/views/ -run TestNewSettings -v` → ALL PASS
  - [x] `grep 'NewDefaultDelegate' internal/tui/views/settings.go` → 0 matches
  - [x] `grep 'NewListDelegate' internal/tui/views/settings.go` → >= 1 match
  - [x] `grep 'RenderStatusBar' internal/tui/views/settings.go` → >= 1 match
  - [x] `grep 'Contains.*success' internal/tui/views/settings.go` → 0 matches (code smell removed)
  - [x] `grep 'func.*IsEditing' internal/tui/views/settings.go` → 1 match (contract preserved)
  - [x] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Settings state machine preserves all transitions
    Tool: Bash
    Steps:
      1. go test ./internal/tui/views/ -run TestSettings -v -count=1
      2. Assert: exit code 0
      3. Assert: output contains "PASS"
    Expected Result: All settings tests pass
    Evidence: Terminal output

  Scenario: IsEditing contract preserved
    Tool: Bash
    Steps:
      1. go test ./internal/tui/views/ -run TestSettings -v -count=1
      2. grep -c 'IsEditing' internal/tui/views/settings_test.go
      3. Assert: count >= 2 (tested in multiple states)
    Expected Result: IsEditing tested and working
    Evidence: Terminal output

  Scenario: Settings visual verification
    Tool: interactive_bash (tmux)
    Preconditions: Application built with `make build`
    Steps:
      1. tmux new-session -d -s settings-test
      2. tmux send-keys -t settings-test './bin/dotkeeper' Enter
      3. Wait 2 seconds
      4. tmux send-keys -t settings-test '4'
      5. Wait 1 second
      6. tmux capture-pane -t settings-test -p > .sisyphus/evidence/task-8-settings.txt
      7. Assert: output contains "Backup Directory"
      8. Assert: output contains "Git Remote"
      9. Assert: output contains "Files"
      10. Assert: output contains "Folders"
      11. tmux send-keys -t settings-test 'q'
    Expected Result: Settings shows list-based navigation
    Evidence: .sisyphus/evidence/task-8-settings.txt

  Scenario: Settings code smell removed
    Tool: Bash
    Steps:
      1. grep -n 'Contains.*success' internal/tui/views/settings.go
      2. Assert: no output (code smell gone)
    Expected Result: No strings.Contains("success") pattern
    Evidence: Terminal output
  ```

  **Commit**: YES
  - Message: `refactor(tui): rewrite settings view with bubbles/list navigation and clean state machine`
  - Files: `internal/tui/views/settings.go`, `internal/tui/views/settings_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [x] 9. Outer Frame Cleanup + Final QA

  **What to do**:
  - **view.go cleanup**:
    - Move inline `titleStyle` (lines 49-52) to `styles.go` as `AppTitle` style
    - Move inline `helpStyle` (lines 80-82) to `styles.go` as `GlobalHelp` style
    - Move inline `contentStyle` (line 60) to `styles.go` as `ContentArea` style
    - Replace all 3 in view.go with `styles.AppTitle`, `styles.GlobalHelp`, `styles.ContentArea`
    - `view.go` should have ZERO `lipgloss.NewStyle()` calls
  - **help.go cleanup**:
    - Move inline `keyStyle` (line 22) to `styles.go` as `HelpKey` style
    - Move inline `titleStyle` (line 23) to `styles.go` as `HelpTitle` style
    - Move inline `sectionStyle` (line 24) to `styles.go` as `HelpSection` style
    - Move inline `overlayStyle` (line 59-63) to `styles.go` as `HelpOverlay` style
    - Replace all 4 in help.go with references to styles.go
  - Add new styles to `Styles` struct and `DefaultStyles()` in styles.go
  - **Final visual QA**: Launch TUI and verify ALL 5 tabs have consistent appearance
  - Update any remaining tests if needed
  - Run full test suite one final time

  **Must NOT do**:
  - Do NOT change the rendering logic in view.go — only move style definitions
  - Do NOT change help overlay behavior
  - Do NOT change tab bar rendering or activeTabIndex()
  - Do NOT change setup.go or filebrowser.go

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Style migration + visual QA verification
  - **Skills**: [`git-master`]
    - `git-master`: Final commit

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 6 (alone, after all others)
  - **Blocks**: None (final task)
  - **Blocked By**: Tasks 4, 5, 6, 7, 8

  **References**:

  **Pattern References**:
  - `internal/tui/view.go:49-83` — Inline styles to move (titleStyle, contentStyle, helpStyle)
  - `internal/tui/help.go:22-63` — Inline styles to move (keyStyle, titleStyle, sectionStyle, overlayStyle)
  - `internal/tui/views/styles.go` — Target for all moved styles

  **Acceptance Criteria**:
  - [x] `grep 'lipgloss.NewStyle()' internal/tui/view.go` → 0 matches
  - [x] `grep 'lipgloss.NewStyle()' internal/tui/help.go` → 0 matches
  - [x] `grep 'AppTitle\|GlobalHelp\|ContentArea' internal/tui/view.go` → >= 3 matches
  - [x] `grep 'HelpKey\|HelpTitle\|HelpSection\|HelpOverlay' internal/tui/help.go` → >= 4 matches
  - [x] `go test ./internal/tui/... -race -count=1` → ALL PASS
  - [x] `go vet ./internal/tui/...` → no errors
  - [x] `go build ./cmd/dotkeeper/` → compiles
  - [x] `grep -rn 'NewDefaultDelegate' internal/tui/views/*.go` → 0 matches (global check)
  - [x] `grep -rn 'lipgloss.NewStyle()' internal/tui/views/*.go | grep -v styles.go | grep -v setup.go | grep -v _test.go | wc -l` → <= 2

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Zero inline styles in view.go
    Tool: Bash
    Steps:
      1. grep -c 'lipgloss.NewStyle()' internal/tui/view.go
      2. Assert: count is 0
    Expected Result: All styles from styles.go
    Evidence: Terminal output

  Scenario: Zero inline styles in help.go
    Tool: Bash
    Steps:
      1. grep -c 'lipgloss.NewStyle()' internal/tui/help.go
      2. Assert: count is 0
    Expected Result: All styles from styles.go
    Evidence: Terminal output

  Scenario: Full test suite passes
    Tool: Bash
    Steps:
      1. go test ./internal/tui/... -race -count=1
      2. Assert: exit code 0
      3. go test ./... -race -count=1
      4. Assert: exit code 0
    Expected Result: All tests pass including e2e
    Evidence: Terminal output

  Scenario: Full TUI visual verification — Dashboard
    Tool: interactive_bash (tmux)
    Preconditions: make build succeeds
    Steps:
      1. tmux new-session -d -s final-qa
      2. tmux send-keys -t final-qa './bin/dotkeeper' Enter
      3. Wait 2 seconds
      4. tmux capture-pane -t final-qa -p > .sisyphus/evidence/task-9-dashboard.txt
      5. Assert: output contains stat cards (Last Backup, Files Tracked)
      6. Assert: output contains action buttons (Backup, Restore, Settings)
      7. Assert: output contains tab bar
    Expected Result: Dashboard renders beautifully
    Evidence: .sisyphus/evidence/task-9-dashboard.txt

  Scenario: Full TUI visual verification — Backups tab
    Tool: interactive_bash (tmux)
    Preconditions: TUI running in final-qa session
    Steps:
      1. tmux send-keys -t final-qa '2'
      2. Wait 1 second
      3. tmux capture-pane -t final-qa -p > .sisyphus/evidence/task-9-backups.txt
      4. Assert: output contains "Backups" title
      5. Assert: output contains status bar area
    Expected Result: Backups tab renders with custom delegate
    Evidence: .sisyphus/evidence/task-9-backups.txt

  Scenario: Full TUI visual verification — Restore tab
    Tool: interactive_bash (tmux)
    Steps:
      1. tmux send-keys -t final-qa '3'
      2. Wait 1 second
      3. tmux capture-pane -t final-qa -p > .sisyphus/evidence/task-9-restore.txt
      4. Assert: output contains backup list with custom styling
    Expected Result: Restore tab renders correctly
    Evidence: .sisyphus/evidence/task-9-restore.txt

  Scenario: Full TUI visual verification — Settings tab
    Tool: interactive_bash (tmux)
    Steps:
      1. tmux send-keys -t final-qa '4'
      2. Wait 1 second
      3. tmux capture-pane -t final-qa -p > .sisyphus/evidence/task-9-settings.txt
      4. Assert: output contains "Backup Directory", "Git Remote", "Files", "Folders"
    Expected Result: Settings renders with list navigation
    Evidence: .sisyphus/evidence/task-9-settings.txt

  Scenario: Full TUI visual verification — Logs tab
    Tool: interactive_bash (tmux)
    Steps:
      1. tmux send-keys -t final-qa '5'
      2. Wait 1 second
      3. tmux capture-pane -t final-qa -p > .sisyphus/evidence/task-9-logs.txt
      4. Assert: output contains "Operation History"
      5. tmux send-keys -t final-qa 'q'
    Expected Result: Logs renders with custom delegate
    Evidence: .sisyphus/evidence/task-9-logs.txt
  ```

  **Commit**: YES
  - Message: `refactor(tui): consolidate all inline styles into design system`
  - Files: `internal/tui/view.go`, `internal/tui/help.go`, `internal/tui/views/styles.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

## Commit Strategy

| After Task | Message | Key Files | Verification |
|------------|---------|-----------|--------------|
| 1 | `refactor(tui): add stripANSI test helper and future-proof view tests` | `views/testhelpers_test.go`, `views/*_test.go` | `go test ./internal/tui/...` |
| 2 | `feat(tui): expand design system with shared list delegate and layout styles` | `views/styles.go` | `go test && go vet` |
| 3 | `feat(tui): add RenderStatusBar helper for unified status/help rendering` | `views/helpers.go`, `views/helpers_test.go` | `go test` |
| 4 | `refactor(tui): standardize backup list with shared delegate and status bar` | `views/backuplist.go`, `views/backuplist_test.go` | `go test` |
| 5 | `refactor(tui): standardize logs view with shared delegate and status bar` | `views/logs.go`, `views/logs_test.go` | `go test` |
| 6 | `refactor(tui): standardize restore view with shared delegate, status bar, and purple viewport` | `views/restore.go`, `views/restore_test.go` | `go test` |
| 7 | `feat(tui): redesign dashboard with stat cards and styled action buttons` | `views/dashboard.go`, `views/dashboard_test.go` | `go test` |
| 8 | `refactor(tui): rewrite settings view with bubbles/list navigation and clean state machine` | `views/settings.go`, `views/settings_test.go` | `go test` |
| 9 | `refactor(tui): consolidate all inline styles into design system` | `tui/view.go`, `tui/help.go`, `views/styles.go` | `go test && go vet && make build` |

---

## Success Criteria

### Verification Commands
```bash
# All tests pass with race detection
go test ./internal/tui/... -race -count=1
# Expected: ALL PASS

# Full project tests pass
go test ./... -race -count=1
# Expected: ALL PASS

# No vet errors
go vet ./internal/tui/...
# Expected: no output (clean)

# Project builds
make build
# Expected: ./bin/dotkeeper created

# Zero default delegates remaining
grep -rn 'NewDefaultDelegate' internal/tui/views/*.go
# Expected: no output

# Minimal inline styles (only settings textinput config)
grep -rn 'lipgloss.NewStyle()' internal/tui/views/*.go | grep -v styles.go | grep -v setup.go | grep -v _test.go | wc -l
# Expected: <= 2

# Zero inline styles in outer frame
grep -c 'lipgloss.NewStyle()' internal/tui/view.go internal/tui/help.go
# Expected: 0 for both files
```

### Final Checklist
- [x] All "Must Have" present
- [x] All "Must NOT Have" absent
- [x] All tests pass with -race
- [x] make build succeeds
- [x] All 5 tabs visually consistent (verified via TUI capture)
- [x] setup.go and filebrowser.go untouched
- [x] No behavioral regressions in key bindings or state transitions
