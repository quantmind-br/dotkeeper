# Path Validation Implementation Learnings

## Conventions
- Use `expandHome()` before validation
- Store expanded paths in config (not raw ~ paths)
- Show errors in red (#FF6B6B)
- Don't clear input on error - let user fix it
- Clear error on successful validation

## Gotchas
- Empty string should be allowed to move to next step
- Need to check both existence AND readability
- Distinguish between files and directories
- Validation must happen BEFORE appending to list

## Decisions
- Store expanded paths (absolute) in config
- Show validation errors inline in the view
- Keep input field value on error for easy correction

## Implementation Complete (Task 2)

### What Was Done
- Added `validationErr string` field to SetupModel struct
- Modified StepAddFiles handler: validates with ValidateFilePath, stores expanded path on success
- Modified StepAddFolders handler: validates with ValidateFolderPath, stores expanded path on success
- Added error display in View() for both steps (red #FF6B6B with ✗ prefix)
- Clear validationErr when moving to next step (empty input)

### Key Implementation Details
- ValidateFilePath and ValidateFolderPath return (expandedPath string, error)
- expandHome() is called before validation
- Input is NOT cleared on validation error (user can fix it)
- Input IS cleared on successful validation
- Error display appears after instructions, before added items list

### Testing Notes
- Build passes successfully
- No LSP diagnostics errors
- All validation logic integrated into handleEnter()

## Implementation Complete (Task 3 - Settings View Validation)

### What Was Done
- Modified `saveFieldValue()` in settings.go to validate paths before saving
- Files: ValidateFilePath() called before appending/updating, returns expanded path
- Folders: ValidateFolderPath() called before appending/updating, returns expanded path
- BackupDir: expandHome() called before saving (no validation needed for backup dir)
- Empty path check: Return early if value is empty (don't add empty paths)
- Error handling: Set m.err on validation failure, clear on success
- Updated View() error display: Green (#04B575) for success, Red (#FF6B6B) for errors

### Key Implementation Details
- ValidateFilePath and ValidateFolderPath return (expandedPath string, error)
- Validation happens BEFORE saving to config
- Early return on validation error prevents invalid paths from being saved
- m.err is cleared on successful validation
- Error display checks for "success" substring to determine color
- BackupDir uses expandHome() but no validation (user can set any path)

### Testing Notes
- Build passes successfully
- No LSP diagnostics errors
- All three field types (files, folders, backupdir) properly handled
