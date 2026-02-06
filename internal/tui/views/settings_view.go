package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the settings view
func (m SettingsModel) View() string {
	var b strings.Builder

	st := m.ctx.Styles

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Center,
			"\n",
			m.spinner.View(),
			"\nScanning paths...",
		)
	}

	b.WriteString(st.Title.Render("Settings") + "\n\n")

	switch m.state {
	case stateListNavigating:
		b.WriteString(m.mainList.View())
	case stateEditingField:
		b.WriteString("Editing: " + m.pathCompleter.View() + "\n\n")
		b.WriteString(m.mainList.View())
	case stateBrowsingFiles:
		b.WriteString(st.Subtitle.Render("Files") + "\n")
		b.WriteString(m.filesList.View())
	case stateBrowsingFolders:
		b.WriteString(st.Subtitle.Render("Folders") + "\n")
		b.WriteString(m.foldersList.View())
	case stateEditingSubItem:
		title := "Editing File"
		listView := m.filesList.View()
		if m.subEditParent == stateBrowsingFolders {
			title = "Editing Folder"
			listView = m.foldersList.View()
		}
		b.WriteString(st.Subtitle.Render(title) + "\n")
		b.WriteString("Value: " + m.pathCompleter.View() + "\n\n")
		b.WriteString(listView)
	case stateFilePickerActive:
		title := "Browse Files"
		if m.filePickerParent == stateBrowsingFolders {
			title = "Browse Folders"
		}
		b.WriteString(st.Subtitle.Render(title) + "\n")
		fpView := m.filePicker.View()
		if m.ctx.Width > 0 {
			fpView = lipgloss.NewStyle().MaxWidth(m.ctx.Width - 4).Render(fpView)
		}
		b.WriteString(fpView)
	}

	if m.inspecting && m.inspectInfo != "" {
		b.WriteString("\n")
		b.WriteString(st.Card.Render(m.inspectInfo))
		b.WriteString("\n")
	}

	b.WriteString("\n" + RenderStatusBar(m.ctx.Width, m.status, m.errMsg, "", st))

	return b.String()
}

func (m SettingsModel) HelpBindings() []HelpEntry {
	switch m.state {
	case stateEditingField, stateEditingSubItem:
		return []HelpEntry{{"Enter", "Save field"}, {"Esc", "Cancel edit"}}
	case stateListNavigating:
		return []HelpEntry{{"↑/↓", "Navigate"}, {"Enter", "Edit"}, {"a", "Add (on lists)"}, {"s", "Save"}, {"Esc", "Back"}}
	case stateBrowsingFiles, stateBrowsingFolders:
		return []HelpEntry{{"↑/↓", "Navigate"}, {"Space", "Toggle"}, {"i", "Inspect"}, {"Enter", "Edit"}, {"a", "Type path"}, {"b", "Browse"}, {"d", "Delete"}, {"s", "Save"}, {"Esc", "Back"}}
	case stateFilePickerActive:
		return []HelpEntry{{"Enter", "Select"}, {"Esc", "Cancel"}}
	default:
		return []HelpEntry{{"↑/↓", "Navigate"}, {"Enter", "Edit field"}, {"a", "Add item"}, {"d", "Delete item"}, {"s", "Save config"}}
	}
}

func (m SettingsModel) StatusHelpText() string {
	switch m.state {
	case stateListNavigating:
		return "↑/↓: Navigate | Enter: Edit | a: Add | s: Save | Esc: Exit"
	case stateEditingField:
		return "Enter: Save field | Esc: Cancel"
	case stateBrowsingFiles, stateBrowsingFolders:
		return "↑/↓: Navigate | Enter: Edit | a: Type path | b: Browse | d: Delete | s: Save | Esc: Back"
	case stateEditingSubItem:
		return "Enter: Save item | Esc: Cancel"
	case stateFilePickerActive:
		return "Enter: Select | ↑/↓: Navigate | Esc: Cancel"
	default:
		return ""
	}
}

func (m SettingsModel) IsEditing() bool {
	return m.state != stateListNavigating
}

func (m SettingsModel) IsInputActive() bool {
	return m.IsEditing()
}
