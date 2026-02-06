package views

import (
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
)

// ProgramContext holds shared state for all TUI views.
// It eliminates prop drilling by providing a single source of truth
// for config, history store, and terminal dimensions.
type ProgramContext struct {
	Config *config.Config
	Store  *history.Store
	Width  int
	Height int
}

// NewProgramContext creates a new ProgramContext with the given config and store.
// This is a convenience constructor for tests and initialization.
func NewProgramContext(cfg *config.Config, store *history.Store) *ProgramContext {
	return &ProgramContext{
		Config: cfg,
		Store:  store,
		Width:  0,
		Height: 0,
	}
}

// ensureProgramContext returns a valid context, creating a minimal one if nil.
// Used by view constructors to handle setup mode where config may be nil.
func ensureProgramContext(ctx *ProgramContext) *ProgramContext {
	if ctx != nil {
		return ctx
	}
	return &ProgramContext{
		Config: nil,
		Store:  nil,
		Width:  0,
		Height: 0,
	}
}
