package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "src.txt")
	dest := filepath.Join(tmpDir, "nested", "dest.txt")

	const content = "hello world"
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write src file: %v", err)
	}

	if err := copyFile(src, dest); err != nil {
		t.Fatalf("copyFile returned error: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}

	if string(data) != content {
		t.Fatalf("expected %q, got %q", content, string(data))
	}
}
