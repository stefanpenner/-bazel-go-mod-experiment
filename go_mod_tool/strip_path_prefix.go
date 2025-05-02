package main

import (
	"path/filepath"
	"strings"
)

// stripPathPrefix removes the given prefix from a path if it exists.
// It also handles removing any leading path separator that might remain.
func stripPathPrefix(path, prefix string) string {
	if prefix == "" {
		return path
	}
	if !strings.HasPrefix(path, prefix) {
		return path
	}
	path = strings.TrimPrefix(path, prefix)
	return strings.TrimPrefix(path, string(filepath.Separator))
}
