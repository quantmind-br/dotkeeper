package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/pathutil"
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
	step          SetupStep
	config        *config.Config
	input         textinput.Model
	addedFiles    []string
	addedFolders  []string
	width         int
	height        int
	err           error
	validationErr string
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
		m.config.BackupDir = pathutil.ExpandHome(m.input.Value())
		if m.config.BackupDir == "" {
			m.config.BackupDir = pathutil.ExpandHome("~/.dotfiles")
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
			m.validationErr = ""
			m.step = StepAddFolders
			m.resetInput()
			m.input.Focus()
		} else {
			expandedPath, err := ValidateFilePath(pathutil.ExpandHome(value))
			if err != nil {
				m.validationErr = err.Error()
			} else {
				m.validationErr = ""
				m.addedFiles = append(m.addedFiles, expandedPath)
				m.input.SetValue("")
			}
		}

	case StepAddFolders:
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			m.validationErr = ""
			m.step = StepConfirm
			m.resetInput()
		} else {
			expandedPath, err := ValidateFolderPath(pathutil.ExpandHome(value))
			if err != nil {
				m.validationErr = err.Error()
			} else {
				m.validationErr = ""
				m.addedFolders = append(m.addedFolders, expandedPath)
				m.input.SetValue("")
			}
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

	styles := DefaultStyles()

	var errMsg string
	var statusText string
	var helpText string

	switch m.step {
	case StepWelcome:
		s.WriteString(styles.Title.Render("Welcome to dotkeeper!") + "\n\n")
		s.WriteString("This wizard will help you set up dotkeeper for the first time.\n\n")
		s.WriteString("You'll configure:\n")
		s.WriteString("  • Backup directory\n")
		s.WriteString("  • Git remote repository\n")
		s.WriteString("  • Files and folders to backup\n")
		helpText = "Enter: continue"

	case StepBackupDir:
		s.WriteString(styles.Title.Render("Step 1: Backup Directory") + "\n\n")
		s.WriteString("Where should backups be stored?\n")
		s.WriteString("(Default: ~/.dotfiles)\n\n")
		s.WriteString(m.input.View() + "\n")
		helpText = "Enter: continue | Esc: back"

	case StepGitRemote:
		s.WriteString(styles.Title.Render("Step 2: Git Remote") + "\n\n")
		s.WriteString("Enter your git repository URL (optional):\n\n")
		s.WriteString(m.input.View() + "\n")
		helpText = "Enter: continue | Esc: back"

	case StepAddFiles:
		s.WriteString(styles.Title.Render("Step 3: Add Files") + "\n\n")
		s.WriteString("Enter file paths to backup (one per line).\n")
		s.WriteString("Press Enter with empty input to continue.\n\n")

		if m.validationErr != "" {
			errMsg = "✗ " + m.validationErr
		}

		if len(m.addedFiles) > 0 {
			s.WriteString(styles.Title.Render("Added files:") + "\n")
			for _, f := range m.addedFiles {
				s.WriteString("  • " + f + "\n")
			}
			s.WriteString("\n")
		}

		s.WriteString(m.input.View() + "\n")
		helpText = "Enter: add/continue | Esc: back"

	case StepAddFolders:
		s.WriteString(styles.Title.Render("Step 4: Add Folders") + "\n\n")
		s.WriteString("Enter folder paths to backup (one per line).\n")
		s.WriteString("Press Enter with empty input to continue.\n\n")

		if m.validationErr != "" {
			errMsg = "✗ " + m.validationErr
		}

		if len(m.addedFolders) > 0 {
			s.WriteString(styles.Title.Render("Added folders:") + "\n")
			for _, f := range m.addedFolders {
				s.WriteString("  • " + f + "\n")
			}
			s.WriteString("\n")
		}

		s.WriteString(m.input.View() + "\n")
		helpText = "Enter: add/continue | Esc: back"

	case StepConfirm:
		s.WriteString(styles.Title.Render("Step 5: Confirm Configuration") + "\n\n")
		s.WriteString(styles.Label.Render("Backup Directory: ") + m.config.BackupDir + "\n")
		s.WriteString(styles.Label.Render("Git Remote: ") + m.config.GitRemote + "\n")
		s.WriteString(styles.Label.Render("Files: ") + fmt.Sprintf("%d", len(m.addedFiles)) + "\n")
		s.WriteString(styles.Label.Render("Folders: ") + fmt.Sprintf("%d", len(m.addedFolders)) + "\n\n")

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

		helpText = "Enter: save | Esc: back"

	case StepComplete:
		if m.err != nil {
			s.WriteString(styles.Title.Render("Setup Failed") + "\n\n")
			errMsg = m.err.Error()
			helpText = "Ctrl+C: exit"
		} else {
			s.WriteString(styles.Success.Render("✓ Setup Complete!") + "\n\n")
			s.WriteString("Your dotkeeper configuration has been saved.\n")
			statusText = "Configuration saved"
			helpText = "Ctrl+C: exit"
		}
	}

	s.WriteString(RenderStatusBar(m.width, statusText, errMsg, helpText))

	return s.String()
}
