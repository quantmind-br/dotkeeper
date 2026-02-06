package tui

import (
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

// ProgramContext is the shared TUI state used by all views.
// It aliases the views package context type to avoid package cycles.
type ProgramContext = views.ProgramContext

func NewProgramContext(cfg *config.Config, store *history.Store) *ProgramContext {
	return &views.ProgramContext{
		Config: cfg,
		Store:  store,
		Width:  0,
		Height: 0,
	}
}
