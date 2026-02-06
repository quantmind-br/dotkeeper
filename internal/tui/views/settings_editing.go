package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/pathutil"
)

// handleEditingFieldInput handles input when actively editing a field
func (m SettingsModel) handleEditingFieldInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateListNavigating
		m.pathCompleter.Input.Blur()
		m.pathCompleter.Input.SetValue("")
		m.resizeLists()
		return m, nil

	case "enter":
		value := strings.TrimSpace(m.pathCompleter.Input.Value())
		m.saveFieldValue(value)
		m.refreshMainList()
		m.state = stateListNavigating
		m.pathCompleter.Input.Blur()
		m.pathCompleter.Input.SetValue("")
		m.resizeLists()
		return m, nil

	default:
		var cmd tea.Cmd
		m.pathCompleter, cmd = m.pathCompleter.Update(msg)
		return m, cmd
	}
}

func (m SettingsModel) handleEditingSubItemInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = m.subEditParent
		m.pathCompleter.Input.Blur()
		m.pathCompleter.Input.SetValue("")
		m.resizeLists()
		return m, nil

	case "enter":
		value := strings.TrimSpace(m.pathCompleter.Input.Value())
		if pathutil.IsGlobPattern(value) {
			results, err := pathutil.ResolveGlob(value, m.ctx.Config.Exclude)
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			if m.subEditParent == stateBrowsingFiles {
				m.ctx.Config.Files = append(m.ctx.Config.Files, results...)
			} else {
				m.ctx.Config.Folders = append(m.ctx.Config.Folders, results...)
			}
			m.status = fmt.Sprintf("Added %d paths from glob", len(results))
			m.refreshPathList(pathListFiles)
			m.refreshPathList(pathListFolders)
			m.refreshMainList()
			m.state = m.subEditParent
			m.pathCompleter.Input.Blur()
			m.pathCompleter.Input.SetValue("")
			m.resizeLists()
			m.loading = true
			return m, m.scanPathDescs()
		}
		m.saveFieldValue(value)
		m.refreshPathList(pathListFiles)
		m.refreshPathList(pathListFolders)
		m.refreshMainList()
		m.state = m.subEditParent
		m.pathCompleter.Input.Blur()
		m.pathCompleter.Input.SetValue("")
		m.resizeLists()
		m.loading = true
		return m, m.scanPathDescs()

	default:
		var cmd tea.Cmd
		m.pathCompleter, cmd = m.pathCompleter.Update(msg)
		return m, cmd
	}
}

// startEditingField initializes field editing
func (m *SettingsModel) startEditingField() {
	m.pathCompleter.Input.Focus()
	switch m.editingFieldIndex {
	case 0:
		m.pathCompleter.Input.SetValue(m.ctx.Config.BackupDir)
	case 1:
		m.pathCompleter.Input.SetValue(m.ctx.Config.GitRemote)
	case 4:
		m.pathCompleter.Input.SetValue(m.ctx.Config.Schedule)
	default:
		m.pathCompleter.Input.SetValue("")
	}
}

func (m *SettingsModel) startEditingSubItem(parent settingsState, index int, value string) {
	m.subEditParent = parent
	m.subEditIndex = index
	m.state = stateEditingSubItem
	m.pathCompleter.Input.Focus()
	m.pathCompleter.Input.SetValue(value)
	m.resizeLists()
}

// saveFieldValue saves the edited field value
func (m *SettingsModel) saveFieldValue(value string) {
	if m.state == stateEditingSubItem && m.subEditParent == stateBrowsingFiles {
		// Check if empty
		if value == "" {
			return
		}
		// Validate file path
		expandedPath, err := ValidateFilePath(value)
		if err != nil {
			m.errMsg = err.Error()
			m.status = ""
			return
		}
		// Clear error and save expanded path
		m.errMsg = ""
		if m.subEditIndex < len(m.ctx.Config.Files) {
			m.ctx.Config.Files[m.subEditIndex] = expandedPath
		} else {
			m.ctx.Config.Files = append(m.ctx.Config.Files, expandedPath)
		}
	} else if m.state == stateEditingSubItem && m.subEditParent == stateBrowsingFolders {
		// Check if empty
		if value == "" {
			return
		}
		// Validate folder path
		expandedPath, err := ValidateFolderPath(value)
		if err != nil {
			m.errMsg = err.Error()
			m.status = ""
			return
		}
		// Clear error and save expanded path
		m.errMsg = ""
		if m.subEditIndex < len(m.ctx.Config.Folders) {
			m.ctx.Config.Folders[m.subEditIndex] = expandedPath
		} else {
			m.ctx.Config.Folders = append(m.ctx.Config.Folders, expandedPath)
		}
	} else {
		switch m.editingFieldIndex {
		case 0:
			m.ctx.Config.BackupDir = pathutil.ExpandHome(value)
		case 1:
			m.ctx.Config.GitRemote = value
		case 4:
			m.ctx.Config.Schedule = value
		case 5:
			m.ctx.Config.Notifications = value == "true"
		}
	}
}
