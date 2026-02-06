package components

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/pathutil"
)

// CompletionResultMsg carries filesystem completion results
// Exported so consumers (Settings/Setup views) can route it
type CompletionResultMsg struct {
	Candidates []CompletionCandidate
	Prefix     string
}

type CompletionCandidate struct {
	Path  string
	IsDir bool
}

type PathCompleter struct {
	Input             textinput.Model
	candidates        []CompletionCandidate
	showCandidates    bool
	selectedCandidate int
}

func NewPathCompleter() PathCompleter {
	ti := textinput.New()
	return PathCompleter{Input: ti}
}

func (p PathCompleter) Update(msg tea.Msg) (PathCompleter, tea.Cmd) {
	switch msg := msg.(type) {
	case CompletionResultMsg:
		if len(msg.Candidates) == 1 {
			// Single match — auto-complete
			completed := msg.Candidates[0].Path
			if msg.Candidates[0].IsDir {
				completed += "/"
			}
			p.Input.SetValue(completed)
			p.Input.CursorEnd()
			p.candidates = nil
			p.showCandidates = false
		} else if len(msg.Candidates) > 0 {
			// Multiple matches — complete common prefix, show candidates
			common := commonPrefix(msg.Candidates)
			if common != "" {
				p.Input.SetValue(common)
				p.Input.CursorEnd()
			}
			p.candidates = msg.Candidates
			if len(p.candidates) > 10 {
				p.candidates = p.candidates[:10]
			}
			p.showCandidates = true
			p.selectedCandidate = 0
		}
		return p, nil

	case tea.KeyMsg:
		if msg.String() == "tab" && !p.showCandidates {
			// Trigger completion
			return p, p.completeCmd()
		}
		if msg.String() == "tab" && p.showCandidates {
			// Cycle through candidates
			if len(p.candidates) > 0 {
				p.selectedCandidate = (p.selectedCandidate + 1) % len(p.candidates)
				c := p.candidates[p.selectedCandidate]
				val := c.Path
				if c.IsDir {
					val += "/"
				}
				p.Input.SetValue(val)
				p.Input.CursorEnd()
			}
			return p, nil
		}
		// Any other key hides candidates
		if p.showCandidates && msg.String() != "tab" {
			p.showCandidates = false
			p.candidates = nil
		}
	}

	var cmd tea.Cmd
	p.Input, cmd = p.Input.Update(msg)
	return p, cmd
}

func (p PathCompleter) View() string {
	var b strings.Builder
	b.WriteString(p.Input.View())
	if p.showCandidates && len(p.candidates) > 0 {
		b.WriteString("\n")
		for i, c := range p.candidates {
			prefix := "  "
			if i == p.selectedCandidate {
				prefix = "> "
			}
			typeIndicator := "[F]"
			if c.IsDir {
				typeIndicator = "[D]"
			}
			b.WriteString(prefix + typeIndicator + " " + filepath.Base(c.Path) + "\n")
		}
	}
	return b.String()
}

func (p PathCompleter) completeCmd() tea.Cmd {
	value := p.Input.Value()
	return func() tea.Msg {
		expanded := pathutil.ExpandHome(value)
		dir := filepath.Dir(expanded)
		base := filepath.Base(expanded)

		// If value ends with /, list directory contents
		if strings.HasSuffix(value, "/") || strings.HasSuffix(expanded, "/") {
			dir = expanded
			base = ""
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return CompletionResultMsg{}
		}

		var candidates []CompletionCandidate
		for _, entry := range entries {
			name := entry.Name()
			if base == "" || strings.HasPrefix(name, base) {
				fullPath := filepath.Join(dir, name)
				// Convert back to ~ notation if applicable
				home, _ := os.UserHomeDir()
				displayPath := fullPath
				if home != "" && strings.HasPrefix(fullPath, home) {
					displayPath = "~" + fullPath[len(home):]
				}
				candidates = append(candidates, CompletionCandidate{
					Path:  displayPath,
					IsDir: entry.IsDir(),
				})
			}
		}
		return CompletionResultMsg{Candidates: candidates, Prefix: value}
	}
}

func commonPrefix(candidates []CompletionCandidate) string {
	if len(candidates) == 0 {
		return ""
	}
	prefix := candidates[0].Path
	for _, c := range candidates[1:] {
		for !strings.HasPrefix(c.Path, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}
