package views

import (
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
)

// ProgramContext carries shared state across all TUI views.
type ProgramContext struct {
	Config *config.Config
	Store  *history.Store
	Width  int
	Height int
}

func NewProgramContext(cfg *config.Config, store *history.Store) *ProgramContext {
	return &ProgramContext{
		Config: cfg,
		Store:  store,
		Width:  0,
		Height: 0,
	}
}

func ensureProgramContext(ctx *ProgramContext) *ProgramContext {
	if ctx == nil {
		return NewProgramContext(nil, nil)
	}
	return ctx
}
