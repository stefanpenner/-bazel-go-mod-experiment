package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

func run(cfg Config) error {
	zipFile, err := os.Create(cfg.Output)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", cfg.Output, err)
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	stampFile, err := parseStampFile(cfg.VolatileStampFile)
	if err != nil {
		return fmt.Errorf("failed to parse stamp file %s: %w", cfg.VolatileStampFile, err)
	}

	moduleDir := cfg.ModulePath + "@" + stampFile["VOLATILE_VERSION"]

	if err := addFileToZip(zw, cfg.GoMod, filepath.Join(moduleDir, "go.mod")); err != nil {
		return fmt.Errorf("failed to add go.mod to zip: %w", err)
	}

	for _, src := range cfg.SrcFiles {
		relPath := stripPathPrefix(src, cfg.StripPrefix)
		zipPath := filepath.Join(moduleDir, relPath)
		if err := addFileToZip(zw, src, zipPath); err != nil {
			return fmt.Errorf("failed to add %s to zip: %w", src, err)
		}
	}

	if err := zw.Flush(); err != nil {
		return fmt.Errorf("failed to flush zip: %w", err)
	}
	return nil
}
