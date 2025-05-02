package main

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestAddFileToZip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testContent := "test content"
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("successful file addition", func(t *testing.T) {
		// Create a buffer to hold the zip file
		buf := new(bytes.Buffer)
		zw := zip.NewWriter(buf)

		// Add the file to the zip
		if err := addFileToZip(zw, testFile, "test.txt"); err != nil {
			t.Fatalf("addFileToZip failed: %v", err)
		}

		// Close the zip writer to flush the data
		if err := zw.Close(); err != nil {
			t.Fatalf("failed to close zip writer: %v", err)
		}

		// Read the zip file
		zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			t.Fatalf("failed to read zip: %v", err)
		}

		// Verify the zip contains our file
		if len(zr.File) != 1 {
			t.Fatalf("expected 1 file in zip, got %d", len(zr.File))
		}

		// Verify the file content
		file := zr.File[0]
		if file.Name != "test.txt" {
			t.Errorf("expected file name 'test.txt', got %q", file.Name)
		}

		rc, err := file.Open()
		if err != nil {
			t.Fatalf("failed to open zip file: %v", err)
		}
		defer rc.Close()

		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("failed to read zip file content: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("expected content %q, got %q", testContent, string(content))
		}
	})

	t.Run("nonexistent source file", func(t *testing.T) {
		buf := new(bytes.Buffer)
		zw := zip.NewWriter(buf)
		defer zw.Close()

		err := addFileToZip(zw, filepath.Join(tmpDir, "nonexistent.txt"), "test.txt")
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})
}
