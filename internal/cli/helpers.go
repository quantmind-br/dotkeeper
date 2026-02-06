package cli

import (
	"fmt"
	"os"

	"github.com/diogo/dotkeeper/internal/history"
)

// logHistory is a helper for repeated best-effort logging pattern.
func logHistory(store *history.Store, storeErr error, entry history.HistoryEntry) {
	if storeErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to log history: %v\n", storeErr)
		return
	}
	if err := store.Append(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to log history: %v\n", err)
	}
}
