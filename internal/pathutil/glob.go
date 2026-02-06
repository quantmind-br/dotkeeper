package pathutil

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// MaxGlobResults is the maximum number of paths a glob pattern may resolve to.
const MaxGlobResults = 500

// IsGlobPattern returns true if the string contains glob meta characters.
func IsGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// ResolveGlob resolves a glob pattern to matching paths.
// Supports standard filepath.Glob patterns plus ** via doublestar.
// Results are capped at MaxGlobResults.
// Exclude patterns filter results using filepath.Match on base names.
func ResolveGlob(pattern string, exclude []string) ([]string, error) {
	expanded := ExpandHome(pattern)

	var matches []string
	var err error

	if strings.Contains(expanded, "**") {
		matches, err = doublestar.FilepathGlob(expanded)
	} else {
		matches, err = filepath.Glob(expanded)
	}

	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern: %w", err)
	}

	var filtered []string
	for _, m := range matches {
		baseName := filepath.Base(m)
		excluded := false
		for _, ex := range exclude {
			if matched, _ := filepath.Match(ex, baseName); matched {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, m)
		}
	}

	if len(filtered) > MaxGlobResults {
		return nil, fmt.Errorf("glob matched %d paths, exceeding limit of %d", len(filtered), MaxGlobResults)
	}

	return filtered, nil
}
