# Path Validation on Input

## TL;DR

> **Quick Summary**: Add real-time path validation when user inputs files/folders in setup wizard and settings view
> 
> **Deliverables**:
> - Validation functions in helpers.go
> - Updated setup.go with validation on file/folder input
> - Updated settings.go with validation on file/folder input
> - Clear error messages shown to user
> 
> **Estimated Effort**: Quick
> **Parallel Execution**: NO - sequential
> **Critical Path**: helpers.go → setup.go → settings.go

---

## Context

### Original Request
User wants validation of directories/files when inputting paths in the TUI, so invalid paths are caught immediately rather than failing silently during backup.

### Current State
- Paths are accepted without validation
- Errors only surface during backup ("no files to backup")
- `~` expansion already implemented in collector.go

---

## Work Objectives

### Core Objective
Validate file/folder paths immediately when user enters them, showing clear error messages for invalid paths.

### Must Have
- Validate path exists
- Validate path is readable
- For files: ensure it's a file (not directory)
- For folders: ensure it's a directory (not file)
- Expand `~` before validation
- Show error message if validation fails
- Allow user to correct input

### Must NOT Have
- Don't block valid paths
- Don't require re-entering entire path on error

---

## TODOs

- [x] 1. Add validation functions to helpers.go

  **What to do**:
  Add these functions to `internal/tui/views/helpers.go`:

  ```go
  // Add to imports
  import (
      "fmt"
      "os"
      "path/filepath"
      "strings"
  )

  // PathValidationResult holds the result of path validation
  type PathValidationResult struct {
      Valid        bool
      Exists       bool
      IsDir        bool
      IsFile       bool
      Readable     bool
      Error        string
      ExpandedPath string
  }

  // ValidatePath checks if a path exists and is accessible
  func ValidatePath(path string) PathValidationResult {
      result := PathValidationResult{
          ExpandedPath: expandHome(strings.TrimSpace(path)),
      }

      if result.ExpandedPath == "" {
          result.Error = "path cannot be empty"
          return result
      }

      info, err := os.Stat(result.ExpandedPath)
      if err != nil {
          if os.IsNotExist(err) {
              result.Error = fmt.Sprintf("path does not exist: %s", result.ExpandedPath)
          } else if os.IsPermission(err) {
              result.Error = fmt.Sprintf("permission denied: %s", result.ExpandedPath)
          } else {
              result.Error = fmt.Sprintf("cannot access path: %v", err)
          }
          return result
      }

      result.Exists = true
      result.IsDir = info.IsDir()
      result.IsFile = info.Mode().IsRegular()

      // Check if readable
      if result.IsFile {
          f, err := os.Open(result.ExpandedPath)
          if err != nil {
              result.Error = fmt.Sprintf("file not readable: %s", result.ExpandedPath)
              return result
          }
          f.Close()
          result.Readable = true
      } else if result.IsDir {
          _, err := os.ReadDir(result.ExpandedPath)
          if err != nil {
              result.Error = fmt.Sprintf("directory not readable: %s", result.ExpandedPath)
              return result
          }
          result.Readable = true
      }

      result.Valid = true
      return result
  }

  // ValidateFilePath validates a path intended to be a file
  func ValidateFilePath(path string) (string, error) {
      result := ValidatePath(path)

      if !result.Exists {
          return "", fmt.Errorf("%s", result.Error)
      }

      if result.IsDir {
          return "", fmt.Errorf("path is a directory, not a file: %s", result.ExpandedPath)
      }

      if !result.Readable {
          return "", fmt.Errorf("%s", result.Error)
      }

      return result.ExpandedPath, nil
  }

  // ValidateFolderPath validates a path intended to be a folder
  func ValidateFolderPath(path string) (string, error) {
      result := ValidatePath(path)

      if !result.Exists {
          return "", fmt.Errorf("%s", result.Error)
      }

      if !result.IsDir {
          return "", fmt.Errorf("path is a file, not a directory: %s", result.ExpandedPath)
      }

      if !result.Readable {
          return "", fmt.Errorf("%s", result.Error)
      }

      return result.ExpandedPath, nil
  }
  ```

  **Acceptance Criteria**:
  - [ ] ValidatePath function exists
  - [ ] ValidateFilePath function exists  
  - [ ] ValidateFolderPath function exists
  - [ ] Functions expand ~ in paths

  **Commit**: YES
  - Message: `feat(tui): add path validation helpers`
  - Files: `internal/tui/views/helpers.go`

---

- [x] 2. Add validation error field to SetupModel

  **What to do**:
  In `internal/tui/views/setup.go`, add a field to show validation errors:

  ```go
  // SetupModel represents the setup wizard
  type SetupModel struct {
      step           SetupStep
      config         *config.Config
      input          textinput.Model
      addedFiles     []string
      addedFolders   []string
      width          int
      height         int
      err            error
      validationErr  string  // ADD THIS LINE
  }
  ```

  **Acceptance Criteria**:
  - [x] SetupModel has validationErr field

  **Commit**: NO (group with task 3)

---

- [x] 3. Update setup.go handleEnter to validate paths

  **What to do**:
  Modify `handleEnter()` in `internal/tui/views/setup.go` to validate file/folder inputs:

  ```go
  case StepAddFiles:
      value := strings.TrimSpace(m.input.Value())
      if value == "" {
          // Empty input means move to next step
          m.validationErr = ""  // Clear any previous error
          m.step = StepAddFolders
          m.resetInput()
          m.input.Focus()
      } else {
          // Validate file path
          expandedPath, err := ValidateFilePath(value)
          if err != nil {
              m.validationErr = err.Error()
              // Don't clear input - let user fix it
          } else {
              m.validationErr = ""
              m.addedFiles = append(m.addedFiles, expandedPath)
              m.input.SetValue("")
          }
      }

  case StepAddFolders:
      value := strings.TrimSpace(m.input.Value())
      if value == "" {
          // Empty input means move to next step
          m.validationErr = ""
          m.step = StepConfirm
          m.resetInput()
      } else {
          // Validate folder path
          expandedPath, err := ValidateFolderPath(value)
          if err != nil {
              m.validationErr = err.Error()
              // Don't clear input - let user fix it
          } else {
              m.validationErr = ""
              m.addedFolders = append(m.addedFolders, expandedPath)
              m.input.SetValue("")
          }
      }
  ```

  **Acceptance Criteria**:
  - [x] File paths validated on input
  - [x] Folder paths validated on input
  - [x] Invalid paths show error, keep input
  - [x] Valid paths are expanded and added

  **Commit**: NO (group with task 4)

---

- [x] 4. Update setup.go View to show validation errors

  **What to do**:
  Add error display in the View() function for StepAddFiles and StepAddFolders:

  ```go
  case StepAddFiles:
      s.WriteString(titleStyle.Render("Step 3: Add Files") + "\n\n")
      s.WriteString("Enter file paths to backup (one per line).\n")
      s.WriteString("Press Enter with empty input to continue.\n\n")

      // Show validation error if any
      if m.validationErr != "" {
          errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
          s.WriteString(errorStyle.Render("✗ " + m.validationErr) + "\n\n")
      }

      if len(m.addedFiles) > 0 {
          s.WriteString(highlightStyle.Render("Added files:") + "\n")
          for _, f := range m.addedFiles {
              s.WriteString("  • " + f + "\n")
          }
          s.WriteString("\n")
      }

      s.WriteString(m.input.View() + "\n\n")
      s.WriteString(subtitleStyle.Render("Press Enter to add/continue, Esc to go back"))

  case StepAddFolders:
      s.WriteString(titleStyle.Render("Step 4: Add Folders") + "\n\n")
      s.WriteString("Enter folder paths to backup (one per line).\n")
      s.WriteString("Press Enter with empty input to continue.\n\n")

      // Show validation error if any
      if m.validationErr != "" {
          errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
          s.WriteString(errorStyle.Render("✗ " + m.validationErr) + "\n\n")
      }

      if len(m.addedFolders) > 0 {
          s.WriteString(highlightStyle.Render("Added folders:") + "\n")
          for _, f := range m.addedFolders {
              s.WriteString("  • " + f + "\n")
          }
          s.WriteString("\n")
      }

      s.WriteString(m.input.View() + "\n\n")
      s.WriteString(subtitleStyle.Render("Press Enter to add/continue, Esc to go back"))
  ```

  **Acceptance Criteria**:
  - [x] Error message shown in red
  - [x] Error clears when valid path entered

  **Commit**: YES
  - Message: `feat(tui): validate file/folder paths in setup wizard`
  - Files: `internal/tui/views/setup.go`

---

- [x] 5. Update settings.go to validate paths on save

  **What to do**:
  Modify `saveFieldValue()` in `internal/tui/views/settings.go` to validate:

  ```go
  // saveFieldValue saves the edited field value
  func (m *SettingsModel) saveFieldValue(value string) {
      if m.editingFiles {
          if value == "" {
              return // Don't add empty paths
          }
          // Validate file path
          expandedPath, err := ValidateFilePath(value)
          if err != nil {
              m.err = err.Error()
              return
          }
          if m.fileCursor < len(m.config.Files) {
              m.config.Files[m.fileCursor] = expandedPath
          } else {
              m.config.Files = append(m.config.Files, expandedPath)
          }
          m.err = "" // Clear error on success
      } else if m.editingFolders {
          if value == "" {
              return // Don't add empty paths
          }
          // Validate folder path
          expandedPath, err := ValidateFolderPath(value)
          if err != nil {
              m.err = err.Error()
              return
          }
          if m.folderCursor < len(m.config.Folders) {
              m.config.Folders[m.folderCursor] = expandedPath
          } else {
              m.config.Folders = append(m.config.Folders, expandedPath)
          }
          m.err = "" // Clear error on success
      } else {
          switch m.cursor {
          case 0:
              m.config.BackupDir = expandHome(value)
          case 1:
              m.config.GitRemote = value
          case 4:
              m.config.Schedule = value
          case 5:
              m.config.Notifications = value == "true"
          }
      }
  }
  ```

  **Acceptance Criteria**:
  - [x] Files validated before adding/editing
  - [x] Folders validated before adding/editing
  - [x] Errors shown in err field
  - [x] Paths expanded on save

  **Commit**: YES
  - Message: `feat(tui): validate file/folder paths in settings view`
  - Files: `internal/tui/views/settings.go`

---

- [x] 6. Update settings.go to style error messages

  **What to do**:
  Update the View() function to show errors in red (currently shows in green):

  ```go
  if m.err != "" {
      var errStyle lipgloss.Style
      if strings.Contains(m.err, "success") {
          errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")) // Green
      } else {
          errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")) // Red
      }
      b.WriteString("\n" + errStyle.Render(m.err) + "\n")
  }
  ```

  **Acceptance Criteria**:
  - [x] Success messages in green
  - [x] Error messages in red

  **Commit**: NO (group with task 5)

---

- [x] 7. Build and test

  **What to do**:
  - Run `go build ./...`
  - Run `go test ./...`
  - Test manually:
    1. Run setup wizard
    2. Enter invalid file path → should show error
    3. Enter valid file path → should add to list
    4. Enter invalid folder path → should show error
    5. Enter valid folder path → should add to list

  **Acceptance Criteria**:
  - [x] Build succeeds
  - [x] Tests pass
  - [x] Manual validation works

  **Commit**: NO

---

## Commit Strategy

| After Task | Message | Files |
|------------|---------|-------|
| 1 | `feat(tui): add path validation helpers` | helpers.go |
| 4 | `feat(tui): validate file/folder paths in setup wizard` | setup.go |
| 5 | `feat(tui): validate file/folder paths in settings view` | settings.go |

---

## Success Criteria

- [x] Invalid paths show immediate error
- [x] Valid paths are accepted and expanded
- [x] Files vs folders validation works correctly
- [x] Error messages are clear and actionable

**PLAN COMPLETE** - Finalized 2026-02-05
