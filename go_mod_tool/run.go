package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func run(cfg Config) error {
	// Create output directory
	if err := os.MkdirAll(cfg.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", cfg.Output, err)
	}

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
	moduleDir := filepath.Join(cfg.Output, cfg.ModulePath+"@"+version)

	// Create the module directory
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory %s: %w", moduleDir, err)
	}

	// TODO: now look at the go.mod, and update versions based on the version set in the status file

	// Copy go.mod to the module directory
	goModDest := filepath.Join(moduleDir, "go.mod")
	if err := copyFile(cfg.GoMod, goModDest); err != nil {
		return fmt.Errorf("failed to copy go.mod: %w", err)
	}

	// Copy all source files to the module directory
	for _, src := range cfg.SrcFiles {
		relPath := stripPathPrefix(src, cfg.StripPrefix)
		destPath := filepath.Join(moduleDir, relPath)
		
		if err := copyFile(src, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", src, err)
		}
	}

	return nil
}
