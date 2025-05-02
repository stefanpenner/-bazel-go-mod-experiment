package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	goModContent := "module example.com/test"
	goModFile := filepath.Join(tmpDir, "go.mod")
	require.NoError(t, os.WriteFile(goModFile, []byte(goModContent), 0644))

	srcContent := "package test"
	srcFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(srcFile, []byte(srcContent), 0644))

	stampContent := "VOLATILE_VERSION v1.0.0"
	statusFile := filepath.Join(tmpDir, "stamp.txt")
	require.NoError(t, os.WriteFile(statusFile, []byte(stampContent), 0644))

	tests := []struct {
		name        string
		cfg         Config
		wantErr     bool
		wantFiles   []string
		wantContent map[string]string
	}{
		{
			name: "basic zip creation",
			cfg: Config{
				Output:             filepath.Join(tmpDir, "out.zip"),
				ModulePath:         "example.com/test",
				GoMod:              goModFile,
				SrcFiles:           []string{srcFile},
				VolatileStatusFile: statusFile,
				StripPrefix:        tmpDir,
			},
			wantFiles: []string{
				"example.com/test@v1.0.0/go.mod",
				"example.com/test@v1.0.0/test.go",
			},
			wantContent: map[string]string{
				"example.com/test@v1.0.0/go.mod":  goModContent,
				"example.com/test@v1.0.0/test.go": srcContent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.cfg)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify zip contents
			r, err := zip.OpenReader(tt.cfg.Output)
			require.NoError(t, err)
			defer r.Close()

			// Check file names
			var foundFiles []string
			for _, f := range r.File {
				foundFiles = append(foundFiles, f.Name)
			}
			assert.ElementsMatch(t, tt.wantFiles, foundFiles, "zip file contents don't match")

			// Check file contents
			for name, wantContent := range tt.wantContent {
				found := false
				for _, f := range r.File {
					if f.Name == name {
						found = true
						rc, err := f.Open()
						require.NoError(t, err)
						content, err := io.ReadAll(rc)
						rc.Close()
						require.NoError(t, err)
						gotContent := string(content)
						if gotContent != wantContent {
							t.Errorf("content mismatch for %s:\nwant:\n%s\ngot:\n%s", name, wantContent, gotContent)
						}
						break
					}
				}
				assert.True(t, found, "file %s not found in zip", name)
			}
		})
	}
}
