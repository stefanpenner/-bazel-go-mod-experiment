package main

import (
	"archive/zip"
	"io"
	"os"
)

func addFileToZip(zw *zip.Writer, srcPath, zipPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	zipEntry, err := zw.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(zipEntry, srcFile)
	return err
}
