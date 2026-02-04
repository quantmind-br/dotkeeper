package restore

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
)

// GenerateDiff generates a unified diff between backup content and current file
func GenerateDiff(backupContent []byte, currentPath string) (*DiffResult, error) {
	result := &DiffResult{
		BackupPath:  "(backup)",
		CurrentPath: currentPath,
	}

	// Read current file content
	currentContent, err := os.ReadFile(currentPath)
	if os.IsNotExist(err) {
		// File doesn't exist, show as new file
		result.HasDifference = true
		result.Diff = formatNewFileDiff(currentPath, backupContent)
		return result, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read current file: %w", err)
	}

	// Compare contents
	if bytes.Equal(backupContent, currentContent) {
		result.HasDifference = false
		result.Diff = ""
		return result, nil
	}

	result.HasDifference = true
	result.Diff = unifiedDiff(
		string(currentContent), currentPath,
		string(backupContent), "(backup)",
	)

	return result, nil
}

// GenerateDiffFromFiles generates a unified diff between two file paths
func GenerateDiffFromFiles(backupPath, currentPath string) (*DiffResult, error) {
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	return GenerateDiff(backupContent, currentPath)
}

// formatNewFileDiff formats diff for a completely new file
func formatNewFileDiff(path string, content []byte) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("--- /dev/null\t%s\n", time.Now().Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("+++ %s\t(new from backup)\n", path))

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 {
		buf.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(lines)))
		for _, line := range lines {
			buf.WriteString("+" + line + "\n")
		}
	}

	return buf.String()
}

// unifiedDiff generates a simple unified diff format
func unifiedDiff(aContent, aName, bContent, bName string) string {
	aLines := strings.Split(aContent, "\n")
	bLines := strings.Split(bContent, "\n")

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("--- %s\t(current)\n", aName))
	buf.WriteString(fmt.Sprintf("+++ %s\t(backup)\n", bName))

	// Simple diff: find differences using LCS-based algorithm
	hunks := computeHunks(aLines, bLines)
	for _, hunk := range hunks {
		buf.WriteString(hunk)
	}

	return buf.String()
}

// computeHunks computes diff hunks between two line slices
func computeHunks(a, b []string) []string {
	// Compute LCS matrix
	lcs := computeLCS(a, b)

	// Build diff from LCS
	var hunks []string
	var currentHunk strings.Builder
	var hunkStartA, hunkStartB int
	var hunkLinesA, hunkLinesB int
	inHunk := false
	contextLines := 3

	// Track positions
	i, j := 0, 0
	lcsIdx := 0

	for i < len(a) || j < len(b) {
		// Both lines match LCS
		if i < len(a) && j < len(b) && lcsIdx < len(lcs) && a[i] == lcs[lcsIdx] && b[j] == lcs[lcsIdx] {
			if inHunk {
				// Add context line to current hunk
				currentHunk.WriteString(" " + a[i] + "\n")
				hunkLinesA++
				hunkLinesB++
			}
			i++
			j++
			lcsIdx++
			continue
		}

		// Start new hunk if needed
		if !inHunk {
			inHunk = true
			hunkStartA = i + 1
			hunkStartB = j + 1
			hunkLinesA = 0
			hunkLinesB = 0
			currentHunk.Reset()

			// Add context before (up to 3 lines)
			start := max(0, i-contextLines)
			for k := start; k < i; k++ {
				currentHunk.WriteString(" " + a[k] + "\n")
				hunkLinesA++
				hunkLinesB++
			}
			if start < i {
				hunkStartA = start + 1
				hunkStartB = max(1, j-(i-start)+1)
			}
		}

		// Line only in a (deleted)
		if i < len(a) && (j >= len(b) || (lcsIdx < len(lcs) && a[i] != lcs[lcsIdx])) {
			if j < len(b) && b[j] == a[i] {
				// Lines match, just context
				currentHunk.WriteString(" " + a[i] + "\n")
				hunkLinesA++
				hunkLinesB++
				i++
				j++
			} else if lcsIdx >= len(lcs) || a[i] != lcs[lcsIdx] {
				currentHunk.WriteString("-" + a[i] + "\n")
				hunkLinesA++
				i++
			}
			continue
		}

		// Line only in b (added)
		if j < len(b) && (i >= len(a) || (lcsIdx < len(lcs) && b[j] != lcs[lcsIdx])) {
			currentHunk.WriteString("+" + b[j] + "\n")
			hunkLinesB++
			j++
			continue
		}
	}

	// Close final hunk
	if inHunk && currentHunk.Len() > 0 {
		header := fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", hunkStartA, hunkLinesA, hunkStartB, hunkLinesB)
		hunks = append(hunks, header+currentHunk.String())
	}

	return hunks
}

// computeLCS computes the Longest Common Subsequence of two string slices
func computeLCS(a, b []string) []string {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return nil
	}

	// Create DP table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	// Fill DP table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// Backtrack to find LCS
	lcs := make([]string, dp[m][n])
	i, j, idx := m, n, dp[m][n]-1
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs[idx] = a[i-1]
			i--
			j--
			idx--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}

// IsBinaryFile checks if content appears to be binary
func IsBinaryFile(content []byte) bool {
	// Check for null bytes (common in binary files)
	for _, b := range content[:min(512, len(content))] {
		if b == 0 {
			return true
		}
	}
	return false
}

// FormatDiffStats formats diff statistics
func FormatDiffStats(diff string) (added, removed int) {
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removed++
		}
	}
	return
}
