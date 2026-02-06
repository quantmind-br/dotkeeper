package views

import (
	"regexp"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

// executeBatchCmd executes a batch command and returns the first non-spinner message.
// Batch commands return tea.BatchMsg which contains multiple commands.
// This helper recursively executes commands until it finds a domain message.
func executeBatchCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if msg == nil {
		return nil
	}
	// If it's a batch, execute the first command in the batch
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil {
				innerMsg := executeBatchCmd(t, c)
				// Skip spinner ticks, return the first domain message
				if _, isSpinnerTick := innerMsg.(spinner.TickMsg); !isSpinnerTick && innerMsg != nil {
					return innerMsg
				}
			}
		}
		return nil
	}
	return msg
}
