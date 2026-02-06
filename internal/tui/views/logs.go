package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
)

// LogsModel represents the logs view
type LogsModel struct {
	config *config.Config
	store  *history.Store
	list   list.Model
	width  int
	height int
	filter string // "all", "backup", "restore"
	err    string
}

// logItem adapts history.HistoryEntry to list.Item
type logItem struct {
	entry history.HistoryEntry
}

func (i logItem) Title() string {
	status := "✓"
	if i.entry.Status == "error" {
		status = "✗"
	}
	// Local time "2006-01-02 15:04"
	ts := i.entry.Timestamp.Local().Format("2006-01-02 15:04")
	return fmt.Sprintf("%s | %s | %s", ts, i.entry.Operation, status)
}

func (i logItem) Description() string {
	size := formatBytes(i.entry.TotalSize)
	// Duration
	duration := time.Duration(i.entry.DurationMs) * time.Millisecond
	return fmt.Sprintf("%d files, %s, %s", i.entry.FileCount, size, duration.String())
}

func (i logItem) FilterValue() string {
	return i.entry.Operation
}

// formatBytes converts bytes to human readable string
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// NewLogs creates a new logs model
// Accepts optional store for backward compatibility with existing model.go
// until Task 4 updates the call site.
func NewLogs(cfg *config.Config, stores ...*history.Store) LogsModel {
	var store *history.Store
	if len(stores) > 0 {
		store = stores[0]
	}

	l := list.New([]list.Item{}, NewListDelegate(), 0, 0)
	l.SetShowTitle(false) // We render our own title
	l.SetShowHelp(false)

	return LogsModel{
		config: cfg,
		store:  store,
		list:   l,
		filter: "all",
	}
}

// Init initializes the logs view
func (m LogsModel) Init() tea.Cmd {
	return m.LoadHistory()
}

// LoadHistory loads history from the store. Exported for cross-view refresh.
func (m LogsModel) LoadHistory() tea.Cmd {
	if m.store == nil {
		return nil
	}

	return func() tea.Msg {
		var entries []history.HistoryEntry
		var err error

		if m.filter == "all" {
			entries, err = m.store.Read(100)
		} else {
			entries, err = m.store.ReadByType(m.filter, 100)
		}

		if err != nil {
			return logsErrorMsg{err}
		}
		return logsLoadedMsg(entries)
	}
}

type logsLoadedMsg []history.HistoryEntry
type logsErrorMsg struct{ err error }

// Update handles messages
func (m LogsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust list size. Reserve space for chrome
		m.list.SetSize(msg.Width, msg.Height-ViewChromeHeight)

	case logsLoadedMsg:
		items := make([]list.Item, len(msg))
		for i, entry := range msg {
			items[i] = logItem{entry: entry}
		}
		m.list.SetItems(items)
		if len(items) == 0 {
			m.err = ""
		}
		return m, nil

	case logsErrorMsg:
		m.err = msg.err.Error()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "f":
			// Cycle filter
			switch m.filter {
			case "all":
				m.filter = "backup"
			case "backup":
				m.filter = "restore"
			case "restore":
				m.filter = "all"
			default:
				m.filter = "all"
			}
			m.list.ResetSelected()
			return m, m.LoadHistory()

		case "r":
			return m, m.LoadHistory()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the logs view
func (m LogsModel) View() string {
	var s strings.Builder
	styles := DefaultStyles()

	// Title
	title := "Operation History"
	switch m.filter {
	case "all":
		title += " [all]"
	case "backup":
		title += " [backup]"
	case "restore":
		title += " [restore]"
	}
	s.WriteString(styles.Title.Render(title) + "\n\n")

	// Content
	if m.err != "" {
		s.WriteString("\n")
	} else if len(m.list.Items()) == 0 {
		s.WriteString("No operations recorded yet. Run a backup to get started!\n")
	} else {
		s.WriteString(m.list.View())
	}

	helpText := fmt.Sprintf("f: filter (%s) | r: refresh | ↑/↓: navigate", m.filter)
	s.WriteString(RenderStatusBar(m.width, "", m.err, helpText))

	return s.String()
}

func (m LogsModel) HelpBindings() []HelpEntry {
	return []HelpEntry{
		{"f", "filter"},
		{"r", "refresh"},
		{"↑/↓", "navigate"},
	}
}
