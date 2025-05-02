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

	status, err := parseStatusFile(cfg.VolatileStatusFile)
	if err != nil {
		return fmt.Errorf("failed to parse status file %s: %w", cfg.VolatileStatusFile, err)
	}

	// TODO: rather then just 1 volatile version, we need to support multiple, one per module
	version, has_version := status["VOLATILE_VERSION"]
	// Default VOLATILE_VERSION to __unversioned__ if not set
	if !has_version {
		version = "__unversioned__"
	}
	moduleDir := cfg.ModulePath + "@" + version

	// TODO: now look at the go.mod, and update versions based on the version set in the status file

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
