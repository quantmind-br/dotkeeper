# Learnings

## Session: ses_3d4637b25ffePDzq1Lws6loI7j (2026-02-05)

### Code Patterns
- ConfigCommand pattern (config.go:14-54): Uses flag.NewFlagSet, fs.Usage, fs.NArg() check, switch on fs.Arg(0)
- Main switch is at main.go:45-61 - add new case after line 53
- PrintHelp format is at main.go:66-88

### API Details
- notify.SendSuccess(backupName string, duration time.Duration) - uses BackupResult fields
- notify.SendError(err error) - takes the error directly
- notify.Send() gracefully handles missing notify-send (returns nil)
