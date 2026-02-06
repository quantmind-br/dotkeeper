package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandHome expands ~ and ~/ prefixes to the user's home directory.
// Returns the path unchanged if it doesn't start with ~ or if home directory lookup fails.
func ExpandHome(p string) string {
	if strings.HasPrefix(p, "~/") || p == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
