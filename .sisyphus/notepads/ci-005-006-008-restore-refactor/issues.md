
## Issues Discovered (2026-02-05)

### Diff Algorithm Infinite Loop
**Location:** `internal/restore/diff.go:computeHunks()`

**Symptom:** Test hangs indefinitely when comparing very short strings (single words without newlines)

**Trigger:** 
- Original content: "original"
- Modified content: "modified"
- Causes infinite loop in hunk computation

**Workaround:** Use multi-line content with newlines
- Works: "line1\nline2\nline3\n"
- Fails: "original"

**Root Cause:** The LCS algorithm and hunk computation don't handle edge cases with very short strings properly. The loop at line 108 (`for i < len(a) || j < len(b)`) doesn't make progress in certain conditions.

**Impact:** Low - real dotfiles are typically multi-line files, so this edge case is unlikely in production.

**Recommendation:** Add bounds checking and progress verification in the diff algorithm to prevent infinite loops.
