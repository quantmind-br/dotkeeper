package views

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/pathutil"
	"github.com/diogo/dotkeeper/internal/tui/components"
)

// SetupStep represents the current step in the setup wizard
type SetupStep int

const (
	StepWelcome SetupStep = iota
	StepBackupDir
	StepGitRemote
	StepPresetFiles
	StepPresetFolders
	StepAddFiles
	StepAddFolders
	StepConfirm
	StepComplete
)

// SetupCompleteMsg is emitted when setup is complete
type SetupCompleteMsg struct {
	Config *config.Config
}

type presetsDetectedMsg struct {
	files   []pathutil.DotfilePreset
	folders []pathutil.DotfilePreset
}

func detectPresetsCmd(homeDir string) tea.Cmd {
	return func() tea.Msg {
		files, folders := pathutil.DetectDotfiles(homeDir)
		return presetsDetectedMsg{files: files, folders: folders}
	}
}

// SetupModel represents the setup wizard
type SetupModel struct {
	step          SetupStep
	config        *config.Config
	pathCompleter components.PathCompleter
	filePicker    filepicker.Model
	browsing      bool
	presetFiles   []pathutil.DotfilePreset
	presetFolders []pathutil.DotfilePreset
	presetCursor  int
	presetsLoaded bool
	addedFiles    []string
	addedFolders  []string
	width         int
	height        int
	err           error
	validationErr string
}

// NewSetup creates a new setup wizard model
func NewSetup() SetupModel {
	pc := components.NewPathCompleter()
	pc.Input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	pc.Input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	fp := filepicker.New()
	home, _ := os.UserHomeDir()
	if home != "" {
		fp.CurrentDirectory = home
	}
	fp.ShowHidden = true

	return SetupModel{
		step:          StepWelcome,
		config:        &config.Config{},
		pathCompleter: pc,
		filePicker:    fp,
	}
}

// Init initializes the setup wizard
func (m SetupModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.browsing {
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)

		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			if m.step == StepAddFiles {
				m.addedFiles = append(m.addedFiles, path)
			} else if m.step == StepAddFolders {
				m.addedFolders = append(m.addedFolders, path)
			}
			m.browsing = false
			return m, nil
		}

		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
			m.browsing = false
			return m, nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case presetsDetectedMsg:
		m.presetFiles = msg.files
		m.presetFolders = msg.folders
		m.presetsLoaded = true
		return m, nil

	case components.CompletionResultMsg:
		var cmd tea.Cmd
		m.pathCompleter, cmd = m.pathCompleter.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "b":
			if m.step == StepAddFiles || m.step == StepAddFolders {
				m.browsing = true
				return m, m.filePicker.Init()
			}

		case "esc":
			// Go back to previous step (except on Welcome and Complete)
			if m.step > StepWelcome && m.step != StepComplete {
				m.step--
				m.resetInput()
				// Reset cursor when going back to preset steps
				if m.step == StepPresetFiles || m.step == StepPresetFolders {
					m.presetCursor = 0
				}
			}
			return m, nil

		case "enter":
			return m.handleEnter()

		case "up", "k":
			if m.step == StepPresetFiles || m.step == StepPresetFolders {
				if m.presetCursor > 0 {
					m.presetCursor--
				}
			}

		case "down", "j":
			if m.step == StepPresetFiles {
				if m.presetCursor < len(m.presetFiles)-1 {
					m.presetCursor++
				}
			} else if m.step == StepPresetFolders {
				if m.presetCursor < len(m.presetFolders)-1 {
					m.presetCursor++
				}
			}

		case " ":
			if m.step == StepPresetFiles && len(m.presetFiles) > 0 {
				m.presetFiles[m.presetCursor].Selected = !m.presetFiles[m.presetCursor].Selected
			} else if m.step == StepPresetFolders && len(m.presetFolders) > 0 {
				m.presetFolders[m.presetCursor].Selected = !m.presetFolders[m.presetCursor].Selected
			}
		}

		// Handle text input for steps that need it
		if m.step == StepBackupDir || m.step == StepGitRemote || m.step == StepAddFiles || m.step == StepAddFolders {
			var cmd tea.Cmd
			m.pathCompleter, cmd = m.pathCompleter.Update(msg)
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
		m.pathCompleter.Input.SetValue("~/.dotfiles")
		m.pathCompleter.Input.Focus()

	case StepBackupDir:
		m.config.BackupDir = pathutil.ExpandHome(m.pathCompleter.Input.Value())
		if m.config.BackupDir == "" {
			m.config.BackupDir = pathutil.ExpandHome("~/.dotfiles")
		}
		m.step = StepGitRemote
		m.resetInput()
		m.pathCompleter.Input.Focus()

	case StepGitRemote:
		m.config.GitRemote = m.pathCompleter.Input.Value()
		m.step = StepPresetFiles
		m.presetCursor = 0

		// Use UserHomeDir to detect dotfiles relative to user home
		home, _ := os.UserHomeDir()
		return m, detectPresetsCmd(home)

	case StepPresetFiles:
		m.step = StepPresetFolders
		m.presetCursor = 0
		return m, nil

	case StepPresetFolders:
		// Process selected presets
		for _, p := range m.presetFiles {
			if p.Selected {
				m.addedFiles = append(m.addedFiles, p.FullPath)
			}
		}
		for _, p := range m.presetFolders {
			if p.Selected {
				m.addedFolders = append(m.addedFolders, p.FullPath)
			}
		}

		m.step = StepAddFiles
		m.resetInput()
		m.pathCompleter.Input.Focus()

	case StepAddFiles:
		value := strings.TrimSpace(m.pathCompleter.Input.Value())
		if value == "" {
			m.validationErr = ""
			m.step = StepAddFolders
			m.resetInput()
			m.pathCompleter.Input.Focus()
		} else if pathutil.IsGlobPattern(value) {
			results, err := pathutil.ResolveGlob(value, nil)
			if err != nil {
				m.validationErr = err.Error()
			} else {
				m.validationErr = fmt.Sprintf("Added %d paths from glob", len(results))
				m.addedFiles = append(m.addedFiles, results...)
				m.pathCompleter.Input.SetValue("")
			}
		} else {
			expandedPath, err := ValidateFilePath(pathutil.ExpandHome(value))
			if err != nil {
				m.validationErr = err.Error()
			} else {
				m.validationErr = ""
				m.addedFiles = append(m.addedFiles, expandedPath)
				m.pathCompleter.Input.SetValue("")
			}
		}

	case StepAddFolders:
		value := strings.TrimSpace(m.pathCompleter.Input.Value())
		if value == "" {
			m.validationErr = ""
			m.step = StepConfirm
			m.resetInput()
		} else if pathutil.IsGlobPattern(value) {
			results, err := pathutil.ResolveGlob(value, nil)
			if err != nil {
				m.validationErr = err.Error()
			} else {
				m.validationErr = fmt.Sprintf("Added %d paths from glob", len(results))
				m.addedFolders = append(m.addedFolders, results...)
				m.pathCompleter.Input.SetValue("")
			}
		} else {
			expandedPath, err := ValidateFolderPath(pathutil.ExpandHome(value))
			if err != nil {
				m.validationErr = err.Error()
			} else {
				m.validationErr = ""
				m.addedFolders = append(m.addedFolders, expandedPath)
				m.pathCompleter.Input.SetValue("")
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
	m.pathCompleter.Input.SetValue("")
	m.pathCompleter.Input.Blur()
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
		s.WriteString(m.pathCompleter.View() + "\n")
		helpText = "Enter: continue | Esc: back"

	case StepGitRemote:
		s.WriteString(styles.Title.Render("Step 2: Git Remote") + "\n\n")
		s.WriteString("Enter your git repository URL (optional):\n\n")
		s.WriteString(m.pathCompleter.View() + "\n")
		helpText = "Enter: continue | Esc: back"

	case StepPresetFiles:
		s.WriteString(styles.Title.Render("Step 3: Select File Presets") + "\n\n")
		if !m.presetsLoaded {
			s.WriteString("Scanning for dotfiles...\n")
		} else if len(m.presetFiles) == 0 {
			s.WriteString("No common dotfiles detected.\n")
		} else {
			s.WriteString("Detected dotfiles on your system:\n\n")
			for i, p := range m.presetFiles {
				cursor := "  "
				if i == m.presetCursor {
					cursor = "> "
				}
				checked := "[ ]"
				if p.Selected {
					checked = "[x]"
				}

				label := fmt.Sprintf("%s %s %s (%s)", cursor, checked, p.Path, formatBytes(p.Size))
				if i == m.presetCursor {
					s.WriteString(styles.Selected.Render(label) + "\n")
				} else {
					s.WriteString(label + "\n")
				}
			}
		}
		helpText = "Space: toggle | Enter: continue | Esc: back"

	case StepPresetFolders:
		s.WriteString(styles.Title.Render("Step 4: Select Folder Presets") + "\n\n")
		if !m.presetsLoaded {
			s.WriteString("Scanning for dotfiles...\n")
		} else if len(m.presetFolders) == 0 {
			s.WriteString("No common config folders detected.\n")
		} else {
			s.WriteString("Detected config folders on your system:\n\n")
			for i, p := range m.presetFolders {
				cursor := "  "
				if i == m.presetCursor {
					cursor = "> "
				}
				checked := "[ ]"
				if p.Selected {
					checked = "[x]"
				}

				fileCount := fmt.Sprintf("%d files", p.FileCount)
				label := fmt.Sprintf("%s %s %s (%s, %s)", cursor, checked, p.Path, fileCount, formatBytes(p.Size))
				if i == m.presetCursor {
					s.WriteString(styles.Selected.Render(label) + "\n")
				} else {
					s.WriteString(label + "\n")
				}
			}
		}
		helpText = "Space: toggle | Enter: continue | Esc: back"

	case StepAddFiles:
		s.WriteString(styles.Title.Render("Step 5: Add Custom Files") + "\n\n")

		if m.browsing {
			s.WriteString("Browse for files:\n\n")
			s.WriteString(m.filePicker.View())
			helpText = "Enter: select | ↑/↓: navigate | Esc: cancel"
		} else {
			s.WriteString("Enter additional file paths to backup (one per line).\n")
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

			s.WriteString(m.pathCompleter.View() + "\n")
			helpText = "Enter: add/continue | b: Browse | Esc: back"
		}

	case StepAddFolders:
		s.WriteString(styles.Title.Render("Step 6: Add Custom Folders") + "\n\n")

		if m.browsing {
			s.WriteString("Browse for folders:\n\n")
			s.WriteString(m.filePicker.View())
			helpText = "Enter: select | ↑/↓: navigate | Esc: cancel"
		} else {
			s.WriteString("Enter additional folder paths to backup (one per line).\n")
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

			s.WriteString(m.pathCompleter.View() + "\n")
			helpText = "Enter: add/continue | b: Browse | Esc: back"
		}

	case StepConfirm:
		s.WriteString(styles.Title.Render("Step 7: Confirm Configuration") + "\n\n")
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
