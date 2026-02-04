package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/config"
)

// SetupStep represents the current step in the setup wizard
type SetupStep int

const (
	StepWelcome SetupStep = iota
	StepBackupDir
	StepGitRemote
	StepAddFiles
	StepAddFolders
	StepConfirm
	StepComplete
)

// SetupCompleteMsg is emitted when setup is complete
type SetupCompleteMsg struct {
	Config *config.Config
}

// SetupModel represents the setup wizard
type SetupModel struct {
	step         SetupStep
	config       *config.Config
	input        textinput.Model
	addedFiles   []string
	addedFolders []string
	width        int
	height       int
	err          error
}

// NewSetup creates a new setup wizard model
func NewSetup() SetupModel {
	ti := textinput.New()
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	return SetupModel{
		step:   StepWelcome,
		config: &config.Config{},
		input:  ti,
	}
}

// Init initializes the setup wizard
func (m SetupModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Go back to previous step (except on Welcome and Complete)
			if m.step > StepWelcome && m.step != StepComplete {
				m.step--
				m.resetInput()
			}
			return m, nil

		case "enter":
			return m.handleEnter()
		}

		// Handle text input for steps that need it
		if m.step == StepBackupDir || m.step == StepGitRemote || m.step == StepAddFiles || m.step == StepAddFolders {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// handleEnter processes Enter key based on current step
func (m SetupModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case StepWelcome:
		m.step = StepBackupDir
		m.resetInput()
		m.input.SetValue("~/.dotfiles")
		m.input.Focus()

	case StepBackupDir:
		m.config.BackupDir = expandHome(m.input.Value())
		if m.config.BackupDir == "" {
			m.config.BackupDir = expandHome("~/.dotfiles")
		}
		m.step = StepGitRemote
		m.resetInput()
		m.input.Focus()

	case StepGitRemote:
		m.config.GitRemote = m.input.Value()
		m.step = StepAddFiles
		m.resetInput()
		m.input.Focus()

	case StepAddFiles:
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			// Empty input means move to next step
			m.step = StepAddFolders
			m.resetInput()
			m.input.Focus()
		} else {
			// Add file to list
			m.addedFiles = append(m.addedFiles, value)
			m.input.SetValue("")
		}

	case StepAddFolders:
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			// Empty input means move to next step
			m.step = StepConfirm
			m.resetInput()
		} else {
			// Add folder to list
			m.addedFolders = append(m.addedFolders, value)
			m.input.SetValue("")
		}

	case StepConfirm:
		// Save configuration
		m.config.Files = m.addedFiles
		m.config.Folders = m.addedFolders
		if m.config.Notifications == false && len(m.addedFiles) == 0 && len(m.addedFolders) == 0 {
			m.config.Notifications = true // Default to true
		}

		if err := m.config.Save(); err != nil {
			m.err = err
			return m, nil
		}

		m.step = StepComplete
	}

	return m, nil
}

// resetInput clears and resets the text input
func (m *SetupModel) resetInput() {
	m.input.SetValue("")
	m.input.Blur()
}

// View renders the setup wizard
func (m SetupModel) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	switch m.step {
	case StepWelcome:
		s.WriteString(titleStyle.Render("Welcome to dotkeeper!") + "\n\n")
		s.WriteString("This wizard will help you set up dotkeeper for the first time.\n\n")
		s.WriteString("You'll configure:\n")
		s.WriteString("  • Backup directory\n")
		s.WriteString("  • Git remote repository\n")
		s.WriteString("  • Files and folders to backup\n\n")
		s.WriteString(subtitleStyle.Render("Press Enter to continue..."))

	case StepBackupDir:
		s.WriteString(titleStyle.Render("Step 1: Backup Directory") + "\n\n")
		s.WriteString("Where should backups be stored?\n")
		s.WriteString("(Default: ~/.dotfiles)\n\n")
		s.WriteString(m.input.View() + "\n\n")
		s.WriteString(subtitleStyle.Render("Press Enter to continue, Esc to go back"))

	case StepGitRemote:
		s.WriteString(titleStyle.Render("Step 2: Git Remote") + "\n\n")
		s.WriteString("Enter your git repository URL (optional):\n\n")
		s.WriteString(m.input.View() + "\n\n")
		s.WriteString(subtitleStyle.Render("Press Enter to continue, Esc to go back"))

	case StepAddFiles:
		s.WriteString(titleStyle.Render("Step 3: Add Files") + "\n\n")
		s.WriteString("Enter file paths to backup (one per line).\n")
		s.WriteString("Press Enter with empty input to continue.\n\n")

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

		if len(m.addedFolders) > 0 {
			s.WriteString(highlightStyle.Render("Added folders:") + "\n")
			for _, f := range m.addedFolders {
				s.WriteString("  • " + f + "\n")
			}
			s.WriteString("\n")
		}

		s.WriteString(m.input.View() + "\n\n")
		s.WriteString(subtitleStyle.Render("Press Enter to add/continue, Esc to go back"))

	case StepConfirm:
		s.WriteString(titleStyle.Render("Step 5: Confirm Configuration") + "\n\n")
		s.WriteString(highlightStyle.Render("Backup Directory: ") + m.config.BackupDir + "\n")
		s.WriteString(highlightStyle.Render("Git Remote: ") + m.config.GitRemote + "\n")
		s.WriteString(highlightStyle.Render("Files: ") + fmt.Sprintf("%d", len(m.addedFiles)) + "\n")
		s.WriteString(highlightStyle.Render("Folders: ") + fmt.Sprintf("%d", len(m.addedFolders)) + "\n\n")

		if len(m.addedFiles) > 0 {
			s.WriteString("Files:\n")
			for _, f := range m.addedFiles {
				s.WriteString("  • " + f + "\n")
			}
			s.WriteString("\n")
		}

		if len(m.addedFolders) > 0 {
			s.WriteString("Folders:\n")
			for _, f := range m.addedFolders {
				s.WriteString("  • " + f + "\n")
			}
			s.WriteString("\n")
		}

		s.WriteString(subtitleStyle.Render("Press Enter to save, Esc to go back"))

	case StepComplete:
		if m.err != nil {
			s.WriteString(titleStyle.Render("Setup Failed") + "\n\n")
			s.WriteString("Error: " + m.err.Error() + "\n\n")
			s.WriteString(subtitleStyle.Render("Press Ctrl+C to exit"))
		} else {
			s.WriteString(successStyle.Render("✓ Setup Complete!") + "\n\n")
			s.WriteString("Your dotkeeper configuration has been saved.\n\n")
			s.WriteString(subtitleStyle.Render("Press Ctrl+C to exit"))
		}
	}

	return s.String()
}
