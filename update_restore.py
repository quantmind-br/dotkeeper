#!/usr/bin/env python3
import re

with open("/home/diogo/dev/dotkeeper/internal/tui/views/restore.go", "r") as f:
    content = f.read()

# 1. Update NewRestore function
content = content.replace(
    """fl := styles.NewMinimalList()

	vp := viewport.New(0, 0)

	return RestoreModel{
		ctx:           ensureProgramContext(ctx),
		backupList:    l,
		passwordInput: ti,
		fileList:      fl,
		selectedFiles: make(map[string]bool),
		viewport:      vp,
		phase:         phaseBackupList,
	}""",
    """fs := NewFileSelector()

	vp := viewport.New(0, 0)

	return RestoreModel{
		ctx:           ensureProgramContext(ctx),
		backupList:    l,
		passwordInput: ti,
		fileSelector:  fs,
		viewport:      vp,
		phase:         phaseBackupList,
	}""",
)

# 2. Remove old helper methods
content = re.sub(
    r"func \(m \*RestoreModel\) updateFileListSelection\(\) \{[^}]+\}\s*", "", content
)

content = re.sub(
    r"func \(m RestoreModel\) countSelectedFiles\(\) int \{[^}]+\}\s*", "", content
)

content = re.sub(
    r"func \(m RestoreModel\) getSelectedFilePaths\(\) \[\]string \{[^}]+\}\s*",
    "",
    content,
)

# 3. Update runRestore to use fileSelector
content = content.replace(
    "SelectedFiles: m.getSelectedFilePaths(),",
    "SelectedFiles: m.fileSelector.SelectedFiles(),",
)

# 4. Update handleFileSelectKey completely
old_handleFileSelectKey = """func (m RestoreModel) handleFileSelectKey(msg tea.KeyMsg) (RestoreModel, tea.Cmd) {
	switch msg.String() {
	case " ":
		if item := m.fileList.SelectedItem(); item != nil {
			fi := item.(fileItem)
			m.selectedFiles[fi.path] = !m.selectedFiles[fi.path]
			m.updateFileListSelection()
		}
	case "a":
		for path := range m.selectedFiles {
			m.selectedFiles[path] = true
		}
		m.updateFileListSelection()
	case "n":
		for path := range m.selectedFiles {
			m.selectedFiles[path] = false
		}
		m.updateFileListSelection()
	case "d":
		if item := m.fileList.SelectedItem(); item != nil {
			fi := item.(fileItem)
			m.restoreStatus = "Loading diff..."
			m.restoreError = ""
			return m, m.loadDiff(fi.path)
		}
	case "enter":
		selectedCount := m.countSelectedFiles()
		if selectedCount == 0 {
			m.restoreError = "Select at least one file"
		} else {
			m.phase = phaseRestoring
			m.restoreStatus = fmt.Sprintf("Restoring %d files...", selectedCount)
			m.restoreError = ""
			return m, m.runRestore()
		}
	case "esc":
		m.phase = phaseBackupList
		m.selectedFiles = make(map[string]bool)
		m.password = ""
		m.restoreError = ""
		m.restoreStatus = ""
		m.passwordInput.SetValue("")
		m.passwordInput.Blur()
	default:
		var cmd tea.Cmd
		m.fileList, cmd = m.fileList.Update(msg)
		return m, cmd
	}
	return m, nil
}"""

new_handleFileSelectKey = """func (m RestoreModel) handleFileSelectKey(msg tea.KeyMsg) (RestoreModel, tea.Cmd) {
	switch msg.String() {
	case " ":
		m.fileSelector.ToggleItem()
	case "a":
		m.fileSelector.SelectAll()
	case "n":
		m.fileSelector.DeselectAll()
	case "d":
		if path := m.fileSelector.SelectedItemPath(); path != "" {
			m.restoreStatus = "Loading diff..."
			m.restoreError = ""
			return m, m.loadDiff(path)
		}
	case "enter":
		selectedCount := m.fileSelector.SelectedCount()
		if selectedCount == 0 {
			m.restoreError = "Select at least one file"
		} else {
			m.phase = phaseRestoring
			m.restoreStatus = fmt.Sprintf("Restoring %d files...", selectedCount)
			m.restoreError = ""
			return m, m.runRestore()
		}
	case "esc":
		m.phase = phaseBackupList
		m.fileSelector.Reset()
		m.password = ""
		m.restoreError = ""
		m.restoreStatus = ""
		m.passwordInput.SetValue("")
		m.passwordInput.Blur()
	default:
		var cmd tea.Cmd
		m.fileSelector, cmd = m.fileSelector.Update(msg)
		return m, cmd
	}
	return m, nil
}"""

content = content.replace(old_handleFileSelectKey, new_handleFileSelectKey)

# 5. Update handleResultsKey
content = content.replace(
    "m.selectedFiles = make(map[string]bool)", "m.fileSelector.Reset()"
)

# 6. Update WindowSizeMsg case
content = content.replace(
    "m.fileList.SetSize(msg.Width, msg.Height)",
    "m.fileSelector.SetSize(msg.Width, msg.Height)",
)

# 7. Update filesLoadedMsg case
old_filesLoadedMsg = """case filesLoadedMsg:
		items := make([]list.Item, len(msg.files))
		m.selectedFiles = make(map[string]bool)
		for i, entry := range msg.files {
			items[i] = fileItem{
				path:     entry.Path,
				size:     int64(len(entry.Content)),
				selected: false,
			}
			m.selectedFiles[entry.Path] = false
		}
		m.fileList.SetItems(items)
		m.restoreStatus = fmt.Sprintf("Loaded %d files", len(msg.files))
		m.restoreError = ""
		return m, nil"""

new_filesLoadedMsg = """case filesLoadedMsg:
		files := make([]FileEntry, len(msg.files))
		for i, entry := range msg.files {
			files[i] = FileEntry{
				Path:    entry.Path,
				Content: entry.Content,
			}
		}
		m.fileSelector.SetFiles(files)
		m.restoreStatus = fmt.Sprintf("Loaded %d files", len(msg.files))
		m.restoreError = ""
		return m, nil"""

content = content.replace(old_filesLoadedMsg, new_filesLoadedMsg)

# 8. Update View for file selection phase
old_view = """selectedCount := m.countSelectedFiles()
		totalCount := len(m.selectedFiles)
		s.WriteString(st.Value.Render(fmt.Sprintf("%d of %d files selected", selectedCount, totalCount)) + "\n\n")

		s.WriteString(m.fileList.View())"""

new_view = """selectedCount := m.fileSelector.SelectedCount()
		totalCount := m.fileSelector.TotalCount()
		s.WriteString(st.Value.Render(fmt.Sprintf("%d of %d files selected", selectedCount, totalCount)) + "\n\n")

		s.WriteString(m.fileSelector.View())"""

content = content.replace(old_view, new_view)

with open("/home/diogo/dev/dotkeeper/internal/tui/views/restore.go", "w") as f:
    f.write(content)

print("Done!")
