package main

import (
	"testing"
)

func TestStripPathPrefix(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		prefix   string
		expected string
	}{
		{
			name:     "empty prefix",
			path:     "/path/to/file",
			prefix:   "",
			expected: "/path/to/file",
		},
		{
			name:     "prefix not present",
			path:     "/path/to/file",
			prefix:   "/other",
			expected: "/path/to/file",
		},
		{
			name:     "prefix at start",
			path:     "/path/to/file",
			prefix:   "/path",
			expected: "to/file",
		},
		{
			name:     "prefix with trailing slash",
			path:     "/path/to/file",
			prefix:   "/path/",
			expected: "to/file",
		},
		{
			name:     "prefix matches entire path",
			path:     "/path/to/file",
			prefix:   "/path/to/file",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripPathPrefix(tt.path, tt.prefix)
			if got != tt.expected {
				t.Errorf("stripPathPrefix(%q, %q) = %q, want %q", tt.path, tt.prefix, got, tt.expected)
			}
		})
	}
}
