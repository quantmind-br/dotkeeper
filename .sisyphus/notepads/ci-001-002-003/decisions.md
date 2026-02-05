# Decisions

## Session: ses_3d4637b25ffePDzq1Lws6loI7j (2026-02-05)

### CI-001: Schedule Command
- Follow exact ConfigCommand pattern from config.go
- Usage shows: `dotkeeper schedule [enable|disable|status]`
- No flags needed (subcommands handle everything)

### CI-002: Notify Flag
- Default to cfg.Notifications (must load config first)
- Only notify on backup outcomes (success or backup failure)
- Pre-backup errors (config, validation, password) do NOT trigger notifications
