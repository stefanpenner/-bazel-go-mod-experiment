package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStatusFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "valid stamp file",
			content: `BUILD_SCM_HASH abc123
BUILD_SCM_STATUS clean
BUILD_TIMESTAMP 2024-03-20T12:00:00Z`,
			want: map[string]string{
				"BUILD_SCM_HASH":   "abc123",
				"BUILD_SCM_STATUS": "clean",
				"BUILD_TIMESTAMP":  "2024-03-20T12:00:00Z",
			},
			wantErr: false,
		},
		{
			name:    "empty file",
			content: ``,
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "file with empty lines",
			content: `
BUILD_SCM_HASH abc123

BUILD_SCM_STATUS clean
`,
			want: map[string]string{
				"BUILD_SCM_HASH":   "abc123",
				"BUILD_SCM_STATUS": "clean",
			},
			wantErr: false,
		},
		{
			name: "file with VOLATILE_VERSION set",
			content: `BUILD_SCM_HASH abc123
VOLATILE_VERSION v1.2.3
BUILD_SCM_STATUS clean`,
			want: map[string]string{
				"BUILD_SCM_HASH":   "abc123",
				"VOLATILE_VERSION": "v1.2.3",
				"BUILD_SCM_STATUS": "clean",
			},
			wantErr: false,
		},
		{
			name: "file with trailing spaces",
			content: `BUILD_SCM_HASH abc123  
BUILD_SCM_STATUS clean  `,
			want: map[string]string{
				"BUILD_SCM_HASH":   "abc123",
				"BUILD_SCM_STATUS": "clean",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary stamp file
			stampFile := filepath.Join(tmpDir, "stamp.txt")
			err := os.WriteFile(stampFile, []byte(tt.content), 0644)
			require.NoError(t, err, "failed to create stamp file")

			// Parse the stamp file
			got, err := parseStatusFile(stampFile)

			if tt.wantErr {
				assert.Error(t, err, "expected error")
				return
			}

			require.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.want, got, "parsed stamp file contents don't match")
		})
	}

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := parseStatusFile(filepath.Join(tmpDir, "nonexistent.txt"))
		assert.Error(t, err, "expected error for nonexistent file")
	})
}
