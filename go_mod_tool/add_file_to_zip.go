package main

import (
	"io"
	"os"
	"path/filepath"
)

// copyFile copies a file from srcPath to destPath, creating parent directories if needed
func copyFile(srcPath, destPath string) error {
	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
