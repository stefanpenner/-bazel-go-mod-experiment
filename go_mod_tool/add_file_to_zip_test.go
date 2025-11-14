package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testContent := "test content"
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("successful file copy", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "dest", "test.txt")

		// Copy the file
		if err := copyFile(testFile, destFile); err != nil {
			t.Fatalf("copyFile failed: %v", err)
		}

		// Verify the file was copied
		content, err := os.ReadFile(destFile)
		if err != nil {
			t.Fatalf("failed to read copied file: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("expected content %q, got %q", testContent, string(content))
		}
	})

	t.Run("nonexistent source file", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "dest2", "test.txt")
		err := copyFile(filepath.Join(tmpDir, "nonexistent.txt"), destFile)
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})

	t.Run("copy to nested directory", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "nested", "deep", "test.txt")

		// Copy the file to a nested directory
		if err := copyFile(testFile, destFile); err != nil {
			t.Fatalf("copyFile failed: %v", err)
		}

		// Verify the file was copied
		content, err := os.ReadFile(destFile)
		if err != nil {
			t.Fatalf("failed to read copied file: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("expected content %q, got %q", testContent, string(content))
		}
	})
}
