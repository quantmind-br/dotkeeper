# UI/UX Improvements Analysis Report

## Executive Summary

Analyzed the dotkeeper TUI (BubbleTea) and CLI interfaces across **22 Go files** (~4,500 lines). Found **19 issues** spanning usability, visual polish, state handling, and interaction quality. The application has a solid foundation with consistent styling and good keyboard navigation, but several gaps affect user confidence and experience — particularly around feedback during async operations, visual hierarchy, and edge-case state handling.

**Breakdown:** 5 high priority, 8 medium priority, 6 low priority.

---

## Issues Found

### High Priority

---

#### UIUX-001: Dashboard lacks visual hierarchy and feels flat

**Category:** visual | usability

**Affected Components:**
- `internal/tui/views/dashboard.go` (View method, lines 49-71)

**Current State:**
The dashboard renders plain text with no visual structure — just lines of text. "Quick Actions" are rendered as plain strings with no visual distinction. The dashboard is the first thing users see and sets the tone for the entire application.

```go
// Current (lines 49-71)
s += styles.Title.Render("Dashboard") + "\n\n"
s += fmt.Sprintf("Last Backup: %s\n", m.lastBackup.Format("2006-01-02 15:04"))
s += fmt.Sprintf("Files Tracked: %d\n\n", m.fileCount)
s += "Quick Actions:\n"
s += "  [b] Backup now\n"
s += "  [r] Restore\n"
s += "  [s] Settings\n"
```

**Proposed Change:**
- Add a bordered status card for backup info using lipgloss
- Style quick action keys with the accent color (`#7D56F4`)
- Add backup health indicator (time since last backup: green/yellow/red)
- Show backup directory size if available
- Add visual separator between sections

```go
// Proposed
// Status card with border
statusCard := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#7D56F4")).
    Padding(0, 2).
    Width(m.width - 8)

var statusContent strings.Builder
// Show "Last Backup" with color based on recency
if !m.lastBackup.IsZero() {
    age := time.Since(m.lastBackup)
    ageColor := "#04B575" // green: <24h
    if age > 72*time.Hour { ageColor = "#FF5555" } // red: >3d
    else if age > 24*time.Hour { ageColor = "#FFAA00" } // yellow: >1d
    ageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ageColor))
    statusContent.WriteString(fmt.Sprintf("Last Backup:   %s (%s ago)\n",
        ageStyle.Render(m.lastBackup.Format("2006-01-02 15:04")),
        humanizeDuration(age)))
} else {
    statusContent.WriteString(styles.Error.Render("Last Backup:   Never") + "\n")
}
statusContent.WriteString(fmt.Sprintf("Files Tracked: %d\n", m.fileCount))

s += statusCard.Render(statusContent.String()) + "\n\n"

// Quick actions with styled keys
keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
s += "Quick Actions:\n"
s += fmt.Sprintf("  %s Backup now\n", keyStyle.Render("[b]"))
s += fmt.Sprintf("  %s Restore\n", keyStyle.Render("[r]"))
s += fmt.Sprintf("  %s Settings\n", keyStyle.Render("[s]"))
```

**User Benefit:**
First impression is dramatically improved. Users immediately understand backup health status via color coding and can visually scan quick actions more easily.

**Estimated Effort:** small

---

#### UIUX-002: No spinner or progress indicator during backup creation in TUI

**Category:** performance | usability

**Affected Components:**
- `internal/tui/views/backuplist.go` (lines 192-194, 143, 250-256)

**Current State:**
When creating a backup, the status message changes to "Creating backup..." as plain text. For large backups this can take several seconds with no visual feedback that work is happening. Users may think the app is frozen.

```go
// Current (line 193)
m.backupStatus = "Creating backup..."
```

The "Creating backup..." text is rendered via `styles.Success` (green) which is semantically wrong — it's not a success state yet.

**Proposed Change:**
Add a BubbleTea spinner component during async operations. Use `styles.Value` or a dedicated "in-progress" style instead of `Success` for status messages.

```go
// Add to BackupListModel struct
spinner spinner.Model

// In NewBackupList
sp := spinner.New()
sp.Spinner = spinner.Dot
sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

// In View() during creation
if m.creatingBackup && m.backupStatus != "" {
    s.WriteString(m.spinner.View() + " " + m.backupStatus + "\n")
}

// In Update, tick the spinner
case spinner.TickMsg:
    m.spinner, cmd = m.spinner.Update(msg)
    return m, cmd
```

**User Benefit:**
Users see visual confirmation that work is in progress, reducing anxiety about frozen states. The animated spinner provides continuous feedback.

**Estimated Effort:** small

---

#### UIUX-003: Restore view phase 3 has no progress feedback — just static text

**Category:** performance | usability

**Affected Components:**
- `internal/tui/views/restore.go` (lines 522-526)

**Current State:**
Phase 3 (restoring) renders a static "Restoring..." title and a status message with no animation or progress indication:

```go
// Phase 3: Restoring (lines 522-526)
if m.phase == 3 {
    s.WriteString(styles.Title.Render("Restoring...") + "\n\n")
    s.WriteString(m.restoreStatus)
    return s.String()
}
```

**Proposed Change:**
- Add a spinner component (same pattern as UIUX-002)
- Show the count of selected files being restored
- Include a "this may take a moment" hint for user expectations

```go
// Proposed
if m.phase == 3 {
    s.WriteString(styles.Title.Render("Restoring") + "\n\n")
    s.WriteString(m.spinner.View() + " " + m.restoreStatus + "\n\n")
    s.WriteString(styles.Hint.Render("This may take a moment..."))
    return s.String()
}
```

**User Benefit:**
Users aren't left staring at static text during a potentially long operation. Clear feedback that work is happening.

**Estimated Effort:** small

---

#### UIUX-004: CLI backup command has no progress feedback for long operations

**Category:** performance | usability

**Affected Components:**
- `internal/cli/backup.go` (lines 64-78)

**Current State:**
The CLI prints "Starting backup..." and then goes silent until completion or failure. For large backups (many files, large folders), there's no indication of progress.

```go
// Current
fmt.Println("Starting backup...")
result, err := backup.Backup(cfg, password)
```

**Proposed Change:**
Introduce a simple progress indicator — either a spinner (using `briandowns/spinner` or similar) or periodic status dots. At minimum, print file count feedback.

```go
// Option A: Simple dots approach (no new dependency)
fmt.Print("Starting backup")
// ... in backup.Backup, accept a progress callback
// progress(fmt.Sprintf("  Collecting files... (%d found)", count))
// progress(fmt.Sprintf("  Archiving..."))
// progress(fmt.Sprintf("  Encrypting..."))

// Option B: Phase-based feedback
fmt.Println("Starting backup...")
fmt.Printf("  Collecting %d files and %d folders...\n", len(cfg.Files), len(cfg.Folders))
result, err := backup.Backup(cfg, password)
```

**User Benefit:**
Users running CLI backups (especially via terminal) get feedback that the tool is working, not hung.

**Estimated Effort:** medium (requires adding callback or phase reporting to `backup.Backup`)

---

#### UIUX-005: Setup wizard doesn't transition to Dashboard after completion

**Category:** usability

**Affected Components:**
- `internal/tui/views/setup.go` (lines 280-289)
- `internal/tui/update.go` (lines 77-98)

**Current State:**
After setup completes, the StepComplete view says "Press Ctrl+C to exit". The user must exit and re-launch the TUI. But the codebase has `SetupCompleteMsg` handling in `update.go` that initializes all views and switches to Dashboard — yet StepComplete never emits this message. The setup wizard stores the config but doesn't signal completion to the framework.

```go
// setup.go StepComplete (line 288)
s.WriteString(subtitleStyle.Render("Press Ctrl+C to exit"))

// update.go has handler for SetupCompleteMsg (line 79-92) — but it's never triggered
```

**Proposed Change:**
After successful save in StepConfirm, emit `SetupCompleteMsg` instead of showing a static completion screen, or show the success briefly then auto-transition:

```go
// Option A: Emit message immediately after save
case StepConfirm:
    m.config.Files = m.addedFiles
    m.config.Folders = m.addedFolders
    if err := m.config.Save(); err != nil {
        m.err = err
        return m, nil
    }
    return m, func() tea.Msg { return SetupCompleteMsg{Config: m.config} }

// Option B: Show success then transition on any key
case StepComplete:
    s.WriteString(successStyle.Render("Setup Complete!") + "\n\n")
    s.WriteString("Press any key to continue to Dashboard...")
// And handle keypress in Update to emit SetupCompleteMsg
```

**User Benefit:**
Eliminates a dead-end in the onboarding flow. Users smoothly transition from setup to the main application without restarting.

**Estimated Effort:** small

---

### Medium Priority

---

#### UIUX-006: Settings "saved successfully" uses error field for success message

**Category:** visual | usability

**Affected Components:**
- `internal/tui/views/settings.go` (lines 184-191, 383-391)

**Current State:**
The `err` field is overloaded to carry both error messages and the success message "Config saved successfully!". The View method uses `strings.Contains(m.err, "success")` to determine styling:

```go
// Line 188-189
m.err = "Config saved successfully!"
// ...
// Lines 385-389
if strings.Contains(m.err, "success") {
    errStyle = styles.Success
} else {
    errStyle = styles.Error
}
```

**Proposed Change:**
Add a dedicated `savedOk bool` field or use separate `statusMsg` and `errMsg` fields:

```go
type SettingsModel struct {
    // ... existing fields
    statusMsg string // success/info messages
    errMsg    string // error messages only
}

// On save:
m.statusMsg = "Config saved successfully!"
m.errMsg = ""

// In View():
if m.statusMsg != "" {
    b.WriteString("\n" + styles.Success.Render(m.statusMsg) + "\n")
}
if m.errMsg != "" {
    b.WriteString("\n" + styles.Error.Render(m.errMsg) + "\n")
}
```

**User Benefit:**
Cleaner code, no string-matching heuristic. Prevents future bugs where an error message accidentally contains "success".

**Estimated Effort:** small

---

#### UIUX-007: Backup list shows size in raw bytes instead of human-readable format

**Category:** visual | usability

**Affected Components:**
- `internal/tui/views/backuplist.go` (line 24)

**Current State:**
The backup item description shows size in raw bytes:

```go
func (i backupItem) Description() string {
    return fmt.Sprintf("%s - %d bytes", i.date, i.size)
}
```

This renders as "2026-02-05 15:04 - 1048576 bytes" which is hard to parse visually.

**Proposed Change:**
Use the `formatBytes` helper that already exists in `logs.go`:

```go
func (i backupItem) Description() string {
    return fmt.Sprintf("%s - %s", i.date, formatBytes(i.size))
}
```

Note: `formatBytes` is defined in `logs.go` — it should be moved to `helpers.go` and exported as a shared utility.

**User Benefit:**
"1.0 MB" is immediately understandable vs "1048576 bytes".

**Estimated Effort:** trivial

---

#### UIUX-008: Restore file selection shows size in raw bytes

**Category:** visual | usability

**Affected Components:**
- `internal/tui/views/restore.go` (lines 83-85)

**Current State:**
Same issue as UIUX-007 but for file items in the restore view:

```go
func (i fileItem) Description() string {
    return fmt.Sprintf("%d bytes", i.size)
}
```

**Proposed Change:**
```go
func (i fileItem) Description() string {
    return formatBytes(i.size)
}
```

**User Benefit:**
Consistent human-readable sizes throughout the application.

**Estimated Effort:** trivial

---

#### UIUX-009: No confirmation or undo hint after destructive backup deletion in TUI

**Category:** usability

**Affected Components:**
- `internal/tui/views/backuplist.go` (lines 161-166)

**Current State:**
After deleting a backup, the status briefly shows "Deleted: backup-name" and immediately refreshes the list. There's no permanent indication of what was deleted and no undo mechanism or warning about irreversibility.

**Proposed Change:**
- Add a timed status message that persists for several seconds
- Include stronger wording in the delete confirmation: "This cannot be undone"
- Consider adding `ctrl+z` to restore from trash (future enhancement)

```go
// In delete confirmation view (line 245-247)
s.WriteString(fmt.Sprintf("Are you sure you want to delete %s?\n", styles.Value.Render(m.deleteTarget)))
s.WriteString(styles.Error.Render("This action cannot be undone.") + "\n\n")
s.WriteString(styles.Help.Render("y: confirm | any other key: cancel"))
```

**User Benefit:**
Users are clearly warned about irreversibility before committing to deletion.

**Estimated Effort:** trivial

---

#### UIUX-010: Tab bar doesn't indicate number-key shortcuts visually

**Category:** usability

**Affected Components:**
- `internal/tui/components/tabbar.go` (lines 48-65)
- `internal/tui/view.go` (line 83)

**Current State:**
The tab bar shows "1 Dashboard │ 2 Backups │ ..." but the number keys are rendered in the same style as the label text, making them easy to miss. The global help bar at the bottom says "Tab/1-5: switch views" but the tab bar itself doesn't visually emphasize the number shortcuts.

**Proposed Change:**
Style the number key differently from the label text:

```go
// In TabBar.View(), build tab text with styled number
keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA"))
if i == activeIndex {
    tabText = keyStyle.Render(item.Key) + " " + tb.styles.TabActive.Render(label)
} else {
    tabText = keyStyle.Render(item.Key) + " " + tb.styles.TabInactive.Render(label)
}
```

**User Benefit:**
Number shortcuts are visually discoverable, reducing learning curve for keyboard navigation.

**Estimated Effort:** trivial

---

#### UIUX-011: Restore view password phase doesn't clear error on new attempt

**Category:** usability

**Affected Components:**
- `internal/tui/views/restore.go` (lines 268-281)

**Current State:**
When a user enters a wrong password, the error message "Invalid password (attempt 1/3)" is shown. The message persists while the user types a new password. However, `restoreError` is only cleared after successful validation, not when the user starts typing again. This creates visual noise.

**Proposed Change:**
Clear the error message when the user starts typing in the password field:

```go
// In phase 1 key handling, for non-special keys:
default:
    if m.restoreError != "" {
        m.restoreError = "" // Clear error as user types
    }
    var cmd tea.Cmd
    m.passwordInput, cmd = m.passwordInput.Update(msg)
    return m, cmd
```

**User Benefit:**
Error messages don't linger while the user is actively correcting their input.

**Estimated Effort:** trivial

---

#### UIUX-012: CLI `backup` total size shows raw bytes

**Category:** visual

**Affected Components:**
- `internal/cli/backup.go` (line 83)

**Current State:**
```go
fmt.Printf("  Total size: %d bytes\n", result.TotalSize)
```

**Proposed Change:**
```go
fmt.Printf("  Total size: %s\n", formatSize(result.TotalSize))
```

Note: `formatSize` already exists in `list.go` but is unexported. Export it or move to a shared `cli/format.go`.

**User Benefit:**
Consistent human-readable output matching the `list` command's output format.

**Estimated Effort:** trivial

---

#### UIUX-013: Settings view doesn't auto-clear status/error messages

**Category:** usability

**Affected Components:**
- `internal/tui/views/settings.go` (lines 383-391)

**Current State:**
After saving config or encountering a validation error, the message persists at the bottom of the settings view indefinitely until the next action overwrites it. Old success/error messages linger and can be confusing.

**Proposed Change:**
Clear messages when the user navigates away from settings (on tab switch) or after a timeout using `tea.Tick`:

```go
// Option: Clear on any navigation action
case "up", "down":
    m.err = "" // Clear stale messages on navigation
    // ... existing logic
```

**User Benefit:**
Stale feedback doesn't confuse users about the current state.

**Estimated Effort:** small

---

### Low Priority

---

#### UIUX-014: FileBrowserView is unreachable but still initialized

**Category:** usability (code quality)

**Affected Components:**
- `internal/tui/model.go` (line 49, `fileBrowser` field)
- `internal/tui/view.go` (line 73-75, falls back to dashboard)
- `internal/tui/views/filebrowser.go`

**Current State:**
`FileBrowserView` is defined in `ViewState` but excluded from `tabOrder`. It's eagerly initialized in `NewModel()` but never displayed — the view switch falls through to dashboard. The file browser model initializes a `filepicker.Init()` command on startup, wasting resources.

**Proposed Change:**
Either remove `FileBrowserView` entirely (if not planned for future use) or integrate it into the Settings view for file/folder path selection. If keeping for future use, skip initialization until needed.

**User Benefit:**
Cleaner codebase, no wasted initialization. If integrated into Settings, users get a visual file picker instead of typing paths manually.

**Estimated Effort:** small (removal) / medium (integration)

---

#### UIUX-015: Help overlay doesn't show current view name

**Category:** usability

**Affected Components:**
- `internal/tui/help.go` (lines 37-44)

**Current State:**
The help overlay shows "Global" and "Current View" sections but doesn't identify which view the user is currently on:

```go
content.WriteString(sectionStyle.Render("Current View"))
```

**Proposed Change:**
```go
viewName := "Dashboard" // derive from m.state
content.WriteString(sectionStyle.Render("Current View: " + viewName))
```

Pass the view name to `renderHelpOverlay`.

**User Benefit:**
Users know exactly which view's shortcuts they're looking at.

**Estimated Effort:** trivial

---

#### UIUX-016: Logs filter state not visually prominent

**Category:** visual

**Affected Components:**
- `internal/tui/views/logs.go` (lines 174-183)

**Current State:**
The active filter is shown in the title as "[all]", "[backup]", or "[restore]" — appended to "Operation History". This is subtle and easy to miss.

**Proposed Change:**
Add a styled filter indicator below the title or use pill/badge-style rendering:

```go
// Render filter as styled badges
filters := []string{"all", "backup", "restore"}
var filterBar strings.Builder
for _, f := range filters {
    if f == m.filter {
        filterBar.WriteString(styles.TabActive.Render(" " + f + " ") + " ")
    } else {
        filterBar.WriteString(styles.TabInactive.Render(" " + f + " ") + " ")
    }
}
s.WriteString(filterBar.String() + "\n\n")
```

**User Benefit:**
Active filter is immediately visible. Users can see all available filters at a glance instead of cycling blindly.

**Estimated Effort:** small

---

#### UIUX-017: Notifications setting shows "true"/"false" instead of user-friendly text

**Category:** visual

**Affected Components:**
- `internal/tui/views/settings.go` (line 375)

**Current State:**
```go
b.WriteString(styles.Value.Render(fmt.Sprintf("%v", m.config.Notifications)) + "\n")
```

Renders as "true" or "false" which is developer-oriented.

**Proposed Change:**
```go
notifStatus := "Enabled"
if !m.config.Notifications {
    notifStatus = "Disabled"
}
b.WriteString(styles.Value.Render(notifStatus) + "\n")
```

**User Benefit:**
More natural language for boolean settings.

**Estimated Effort:** trivial

---

#### UIUX-018: No visual branding/logo on startup or dashboard

**Category:** visual

**Affected Components:**
- `internal/tui/view.go` (lines 49-54)

**Current State:**
The app title "dotkeeper - Dotfiles Backup Manager" is rendered as styled text but there's no visual branding that makes the tool feel polished.

**Proposed Change:**
Add a small ASCII art logo or styled banner for the title area:

```go
// Simple styled banner
banner := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#7D56F4")).
    MarginLeft(2)

b.WriteString(banner.Render("dotkeeper") + " ")
b.WriteString(lipgloss.NewStyle().
    Foreground(lipgloss.Color("#666666")).
    Render("Dotfiles Backup Manager"))
```

Or a minimal ASCII logo for the setup wizard welcome screen.

**User Benefit:**
Professional feel, brand recognition, better first impression.

**Estimated Effort:** trivial

---

#### UIUX-019: CLI help text example uses `--backup-id` which doesn't exist

**Category:** usability

**Affected Components:**
- `cmd/dotkeeper/main.go` (line 94)

**Current State:**
```go
help := `...
Examples:
  dotkeeper restore --backup-id <id>
...`
```

But the actual restore command uses a positional argument (`dotkeeper restore <backup-name>`), not `--backup-id`.

**Proposed Change:**
```go
help := `...
Examples:
  dotkeeper restore backup-2026-01-01-120000
...`
```

**User Benefit:**
Help text matches actual CLI interface, preventing user confusion.

**Estimated Effort:** trivial

---

## Summary

| Category | Count |
|----------|-------|
| Usability | 10 |
| Visual Polish | 6 |
| Performance Perception | 3 |

**Total Components Analyzed:** 22 files (TUI views, CLI handlers, shared components, styles, helpers)
**Total Issues Found:** 19

### Quick Wins (trivial effort, high value):
1. UIUX-007: Human-readable sizes in backup list
2. UIUX-008: Human-readable sizes in restore file list
3. UIUX-009: Add "cannot be undone" to delete confirmation
4. UIUX-011: Clear password error on new input
5. UIUX-012: Human-readable sizes in CLI backup output
6. UIUX-017: "Enabled/Disabled" instead of true/false
7. UIUX-019: Fix help text example

### High-Impact Improvements:
1. UIUX-001: Dashboard visual overhaul (status card, color-coded backup age)
2. UIUX-002 + UIUX-003: Spinner components for async operations
3. UIUX-005: Fix setup wizard dead-end (emit SetupCompleteMsg)
