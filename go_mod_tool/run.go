package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func run(cfg Config) error {
	if cfg.OutputDir == "" {
		return fmt.Errorf("output directory is required")
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
	moduleDir := filepath.Join(cfg.OutputDir, cfg.ModulePath+"@"+version)

	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		return fmt.Errorf("failed to create module directory %s: %w", moduleDir, err)
	}

	// TODO: now look at the go.mod, and update versions based on the version set in the status file

	if err := copyFile(cfg.GoMod, filepath.Join(moduleDir, "go.mod")); err != nil {
		return fmt.Errorf("failed to copy go.mod: %w", err)
	}

	for _, src := range cfg.SrcFiles {
		relPath := stripPathPrefix(src, cfg.StripPrefix)
		destPath := filepath.Join(moduleDir, relPath)
		if err := copyFile(src, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", src, err)
		}
	}

	return nil
}
