package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/config"
)

// Keep lipgloss import for cursor/prompt styles (component configuration)

// SettingsModel represents the settings view
type SettingsModel struct {
	config         *config.Config
	width          int
	height         int
	editMode       bool
	cursor         int // 0: BackupDir, 1: GitRemote, 2: Files, 3: Folders, 4: Schedule, 5: Notifications
	editingField   bool
	textInput      textinput.Model
	editingFiles   bool
	editingFolders bool
	fileCursor     int
	folderCursor   int
	err            string
}

// NewSettings creates a new settings model
func NewSettings(cfg *config.Config) SettingsModel {
	ti := textinput.New()
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	return SettingsModel{
		config:    cfg,
		textInput: ti,
	}
}

// Init initializes the settings view
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.editingField {
			return m.handleEditingFieldInput(msg)
		}
		if m.editMode {
			return m.handleEditModeInput(msg)
		}
		return m.handleReadOnlyInput(msg)
	}
	return m, nil
}

// handleReadOnlyInput handles input when not in edit mode
func (m SettingsModel) handleReadOnlyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "e":
		m.editMode = true
		m.cursor = 0
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// handleEditModeInput handles input when in edit mode
func (m SettingsModel) handleEditModeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editMode = false
		m.cursor = 0
		m.fileCursor = 0
		m.folderCursor = 0
		m.editingFiles = false
		m.editingFolders = false
		return m, nil

	case "up":
		if m.editingFiles {
			if m.fileCursor > 0 {
				m.fileCursor--
			}
		} else if m.editingFolders {
			if m.folderCursor > 0 {
				m.folderCursor--
			}
		} else {
			if m.cursor > 0 {
				m.cursor--
			}
		}
		return m, nil

	case "down":
		if m.editingFiles {
			if m.fileCursor < len(m.config.Files)-1 {
				m.fileCursor++
			}
		} else if m.editingFolders {
			if m.folderCursor < len(m.config.Folders)-1 {
				m.folderCursor++
			}
		} else {
			if m.cursor < 5 {
				m.cursor++
			}
		}
		return m, nil

	case "enter":
		if m.editingFiles || m.editingFolders {
			m.startEditingField()
		} else if m.cursor == 2 {
			m.editingFiles = true
			m.editingFolders = false
			if len(m.config.Files) > 0 {
				m.fileCursor = 0
			} else {
				m.fileCursor = 0
				m.startEditingField()
			}
		} else if m.cursor == 3 {
			m.editingFolders = true
			m.editingFiles = false
			if len(m.config.Folders) > 0 {
				m.folderCursor = 0
			} else {
				m.folderCursor = 0
				m.startEditingField()
			}
		} else if m.cursor == 5 {
			m.config.Notifications = !m.config.Notifications
		} else {
			m.startEditingField()
		}
		return m, nil

	case "a":
		if m.cursor == 2 {
			m.editingFiles = true
			m.fileCursor = len(m.config.Files)
			m.startEditingField()
		} else if m.cursor == 3 {
			m.editingFolders = true
			m.folderCursor = len(m.config.Folders)
			m.startEditingField()
		}
		return m, nil

	case "d":
		if m.cursor == 2 && m.editingFiles && m.fileCursor < len(m.config.Files) {
			m.config.Files = append(m.config.Files[:m.fileCursor], m.config.Files[m.fileCursor+1:]...)
			if m.fileCursor > 0 {
				m.fileCursor--
			}
		} else if m.cursor == 3 && m.editingFolders && m.folderCursor < len(m.config.Folders) {
			m.config.Folders = append(m.config.Folders[:m.folderCursor], m.config.Folders[m.folderCursor+1:]...)
			if m.folderCursor > 0 {
				m.folderCursor--
			}
		} else if m.cursor == 2 && len(m.config.Files) > 0 {
			m.editingFiles = true
			m.fileCursor = 0
		} else if m.cursor == 3 && len(m.config.Folders) > 0 {
			m.editingFolders = true
			m.folderCursor = 0
		}
		return m, nil

	case "s":
		if err := m.config.Save(); err != nil {
			m.err = err.Error()
		} else {
			m.err = "Config saved successfully!"
		}
		return m, nil
	}
	return m, nil
}

// handleEditingFieldInput handles input when actively editing a field
func (m SettingsModel) handleEditingFieldInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editingField = false
		m.textInput.Blur()
		m.textInput.SetValue("")
		return m, nil

	case "enter":
		value := strings.TrimSpace(m.textInput.Value())
		m.saveFieldValue(value)
		m.editingField = false
		m.textInput.Blur()
		m.textInput.SetValue("")
		return m, nil

	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

// startEditingField initializes field editing
func (m *SettingsModel) startEditingField() {
	m.editingField = true
	m.textInput.Focus()

	if m.editingFiles && m.fileCursor < len(m.config.Files) {
		m.textInput.SetValue(m.config.Files[m.fileCursor])
	} else if m.editingFolders && m.folderCursor < len(m.config.Folders) {
		m.textInput.SetValue(m.config.Folders[m.folderCursor])
	} else {
		switch m.cursor {
		case 0:
			m.textInput.SetValue(m.config.BackupDir)
		case 1:
			m.textInput.SetValue(m.config.GitRemote)
		case 4:
			m.textInput.SetValue(m.config.Schedule)
		}
	}
}

// saveFieldValue saves the edited field value
func (m *SettingsModel) saveFieldValue(value string) {
	if m.editingFiles {
		// Check if empty
		if value == "" {
			return
		}
		// Validate file path
		expandedPath, err := ValidateFilePath(value)
		if err != nil {
			m.err = err.Error()
			return
		}
		// Clear error and save expanded path
		m.err = ""
		if m.fileCursor < len(m.config.Files) {
			m.config.Files[m.fileCursor] = expandedPath
		} else {
			m.config.Files = append(m.config.Files, expandedPath)
		}
	} else if m.editingFolders {
		// Check if empty
		if value == "" {
			return
		}
		// Validate folder path
		expandedPath, err := ValidateFolderPath(value)
		if err != nil {
			m.err = err.Error()
			return
		}
		// Clear error and save expanded path
		m.err = ""
		if m.folderCursor < len(m.config.Folders) {
			m.config.Folders[m.folderCursor] = expandedPath
		} else {
			m.config.Folders = append(m.config.Folders, expandedPath)
		}
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

// View renders the settings view
func (m SettingsModel) View() string {
	var b strings.Builder

	styles := DefaultStyles()

	if m.editMode {
		b.WriteString(styles.Title.Render("Settings [EDIT MODE]") + "\n\n")
	} else {
		b.WriteString(styles.Title.Render("Settings") + "\n")
		b.WriteString(styles.Hint.Render("Press 'e' to edit") + "\n\n")
	}

	if m.editingField {
		b.WriteString("Editing: " + m.textInput.View() + "\n\n")
	}

	fields := []string{
		"Backup Directory",
		"Git Remote",
		"Files",
		"Folders",
		"Schedule",
		"Notifications",
	}

	for i, field := range fields {
		isSelected := m.editMode && m.cursor == i && !m.editingFiles && !m.editingFolders

		if isSelected {
			b.WriteString(styles.Selected.Render(field + ": "))
		} else {
			b.WriteString(styles.Label.Render(field + ": "))
		}

		switch i {
		case 0:
			b.WriteString(styles.Value.Render(m.config.BackupDir) + "\n")
		case 1:
			b.WriteString(styles.Value.Render(m.config.GitRemote) + "\n")
		case 2:
			if m.editingFiles {
				b.WriteString("\n")
				for j, f := range m.config.Files {
					if m.fileCursor == j {
						b.WriteString(styles.Selected.Render("  ["+fmt.Sprintf("%d", j)+"] "+f) + "\n")
					} else {
						b.WriteString("  [" + fmt.Sprintf("%d", j) + "] " + f + "\n")
					}
				}
				if m.fileCursor == len(m.config.Files) {
					b.WriteString(styles.Selected.Render("  [+] Add new file") + "\n")
				} else {
					b.WriteString("  [+] Add new file\n")
				}
			} else {
				b.WriteString(styles.Value.Render(fmt.Sprintf("%d files", len(m.config.Files))) + "\n")
			}
		case 3:
			if m.editingFolders {
				b.WriteString("\n")
				for j, f := range m.config.Folders {
					if m.folderCursor == j {
						b.WriteString(styles.Selected.Render("  ["+fmt.Sprintf("%d", j)+"] "+f) + "\n")
					} else {
						b.WriteString("  [" + fmt.Sprintf("%d", j) + "] " + f + "\n")
					}
				}
				if m.folderCursor == len(m.config.Folders) {
					b.WriteString(styles.Selected.Render("  [+] Add new folder") + "\n")
				} else {
					b.WriteString("  [+] Add new folder\n")
				}
			} else {
				b.WriteString(styles.Value.Render(fmt.Sprintf("%d folders", len(m.config.Folders))) + "\n")
			}
		case 4:
			if m.config.Schedule != "" {
				b.WriteString(styles.Value.Render(m.config.Schedule) + "\n")
			} else {
				b.WriteString(styles.Value.Render("Not scheduled") + "\n")
			}
		case 5:
			b.WriteString(styles.Value.Render(fmt.Sprintf("%v", m.config.Notifications)) + "\n")
		}
	}

	if m.editMode {
		b.WriteString("\n" + styles.Hint.Render("↑/↓: Navigate | Enter: Edit | a: Add | d: Delete | s: Save | Esc: Exit") + "\n")
	}

	if m.err != "" {
		var errStyle lipgloss.Style
		if strings.Contains(m.err, "success") {
			errStyle = styles.Success
		} else {
			errStyle = styles.Error
		}
		b.WriteString("\n" + errStyle.Render(m.err) + "\n")
	}

	return b.String()
}

func (m SettingsModel) HelpBindings() []HelpEntry {
	if m.editingField {
		return []HelpEntry{
			{"Enter", "Save field"},
			{"Esc", "Cancel edit"},
		}
	}
	if m.editMode {
		return []HelpEntry{
			{"↑/↓", "Navigate"},
			{"Enter", "Edit field"},
			{"a", "Add item"},
			{"d", "Delete item"},
			{"s", "Save config"},
			{"Esc", "Exit edit"},
		}
	}
	return []HelpEntry{
		{"e", "Edit mode"},
	}
}
