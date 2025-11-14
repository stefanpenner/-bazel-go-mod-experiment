package main

import (
	"os"
	"path/filepath"
	"strings"
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
		name              string
		cfg               Config
		moduleVersionPath string
		wantErr           bool
		wantFiles         []string
		wantContent       map[string]string
	}{
		{
			name:              "basic directory creation",
			moduleVersionPath: "example.com/test@v1.0.0",
			cfg: Config{
				OutputDir:          filepath.Join(tmpDir, "out"),
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

			moduleDir := filepath.Join(tt.cfg.OutputDir, tt.moduleVersionPath)
			for name, wantContent := range tt.wantContent {
				rel := strings.TrimPrefix(name, tt.moduleVersionPath+"/")
				fullPath := filepath.Join(moduleDir, rel)
				data, err := os.ReadFile(fullPath)
				require.NoError(t, err, "failed reading %s", fullPath)
				assert.Equal(t, wantContent, string(data))
			}

			var found []string
			err = filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
				require.NoError(t, err)
				if !info.IsDir() {
					rel, err := filepath.Rel(moduleDir, path)
					require.NoError(t, err)
					found = append(found, tt.moduleVersionPath+"/"+rel)
				}
				return nil
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantFiles, found)
		})
	}
}
