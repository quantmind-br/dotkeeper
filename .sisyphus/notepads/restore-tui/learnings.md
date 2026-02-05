## RestoreModel Implementation (2026-02-04)

### Key Patterns Established

1. **Struct Fields**: RestoreModel follows BackupListModel pattern with:
   - `backupList list.Model` for backup selection
   - `phase int` for state machine (0-4)
   - `passwordInput textinput.Model` for password entry
   - `fileList list.Model` for file selection
   - `selectedFiles map[string]bool` for tracking selections
   - `viewport viewport.Model` for diff display
   - Status/error fields for user feedback

2. **Async Loading**: 
   - `Init()` returns `m.refreshBackups()`
   - `refreshBackups()` scans backup directory and returns `backupsLoadedMsg`
   - Reuses `backupItem` and `backupsLoadedMsg` from backuplist.go (shared types)

3. **Message Handling**:
   - `WindowSizeMsg`: Updates dimensions and resizes all components
   - `backupsLoadedMsg`: Populates backup list with items

4. **View Rendering**:
   - Phase 0: Shows backup list with help text
   - Other phases: Placeholder for future implementation
   - Status/error messages rendered conditionally

### Testing Notes
- Test verifies phase 0 renders empty backup list correctly
- List title "Backups" is set but not rendered by bubbles list component
- Test checks for help text and "No items" placeholder

### Next Steps
- Phase 1: Password input handling
- Phase 2: File selection from backup contents
- Phase 3: Restore progress display
- Phase 4: Diff preview before restore
