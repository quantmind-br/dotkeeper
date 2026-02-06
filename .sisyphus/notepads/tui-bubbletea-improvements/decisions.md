# Decisions — TUI BubbleTea Improvements


- Kept ProgramContext concrete type in views package and exposed tui.ProgramContext as alias; this preserves constructor ergonomics without package cycles.
- Setup view now accepts ProgramContext and initializes ctx.Config when nil, preserving first-run setup behavior.
