package pathutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde with slash",
			input:    "~/foo",
			expected: filepath.Join(home, "foo"),
		},
		{
			name:     "tilde with nested path",
			input:    "~/foo/bar/baz",
			expected: filepath.Join(home, "foo/bar/baz"),
		},
		{
			name:     "tilde alone",
			input:    "~",
			expected: home,
		},
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "path without tilde unchanged",
			input:    "foo/bar",
			expected: "foo/bar",
		},
		{
			name:     "tilde in middle unchanged",
			input:    "foo/~/bar",
			expected: "foo/~/bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandHome(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
