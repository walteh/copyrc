package state

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/cmd/copyrc-next/pkg/config"
)

func TestTextModification(t *testing.T) {
	// Setup test directory
	dir := t.TempDir()

	// Create test file
	testFile := filepath.Join(dir, "test.copy.txt")
	err := os.WriteFile(testFile, []byte("Hello World"), 0644)
	require.NoError(t, err, "writing test file")

	// Create RemoteTextFile
	file := &RemoteTextFile{
		LocalPath: testFile,
		IsPatched: false,
	}

	// Test cases
	tests := []struct {
		name        string
		fromText    string
		toText      string
		initialText string
		wantText    string
		wantError   string
	}{
		{
			name:        "simple_replacement",
			fromText:    "World",
			toText:      "Universe",
			initialText: "Hello World",
			wantText:    "Hello Universe",
		},
		{
			name:        "multiple_replacements",
			fromText:    "l",
			toText:      "L",
			initialText: "Hello World",
			wantText:    "HeLLo WorLd",
		},
		{
			name:        "no_match",
			fromText:    "Goodbye",
			toText:      "Hello",
			initialText: "Hello World",
			wantText:    "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write initial content
			err := os.WriteFile(testFile, []byte(tt.initialText), 0644)
			require.NoError(t, err, "writing initial content")

			// Apply modification
			mod := config.TextReplacement{
				FromText: tt.fromText,
				ToText:   tt.toText,
			}
			err = file.ApplyModificationToRawRemoteContent(context.Background(), mod)
			if tt.wantError != "" {
				assert.EqualError(t, err, tt.wantError)
				return
			}
			require.NoError(t, err, "applying modification")

			// Verify modified content
			content, err := os.ReadFile(testFile)
			require.NoError(t, err, "reading modified file")
			assert.Equal(t, tt.wantText, string(content), "modified content")

			// Verify patch file exists and contains original content
			if !file.IsPatched {
				t.Error("file should be marked as patched")
			}

			patchPath := filepath.Join(dir, "test.patch.txt")
			assert.Equal(t, patchPath, file.Patch.PatchPath, "patch path")
			assert.FileExists(t, patchPath, "patch file should exist")

			// Verify we can read original content
			origContent, err := file.RawRemoteContent()
			require.NoError(t, err, "reading original content")
			defer origContent.Close()

			data, err := io.ReadAll(origContent)
			require.NoError(t, err, "reading original content")
			assert.Equal(t, tt.initialText, string(data), "original content")
		})
	}
}

func TestRawPatchContent(t *testing.T) {
	// Setup test directory
	dir := t.TempDir()

	// Create test files
	copyFile := filepath.Join(dir, "test.copy.txt")
	patchFile := filepath.Join(dir, "test.patch.txt")

	err := os.WriteFile(copyFile, []byte("Hello Universe"), 0644)
	require.NoError(t, err, "writing copy file")

	err = os.WriteFile(patchFile, []byte("patch content"), 0644)
	require.NoError(t, err, "writing patch file")

	// Test cases
	tests := []struct {
		name      string
		file      *RemoteTextFile
		wantError string
	}{
		{
			name: "valid_patch",
			file: &RemoteTextFile{
				LocalPath: copyFile,
				IsPatched: true,
				Patch: &Patch{
					PatchPath: patchFile,
				},
			},
		},
		{
			name: "not_patched",
			file: &RemoteTextFile{
				LocalPath: copyFile,
				IsPatched: false,
			},
			wantError: "file is not patched",
		},
		{
			name: "missing_patch_info",
			file: &RemoteTextFile{
				LocalPath: copyFile,
				IsPatched: true,
			},
			wantError: "no patch information available",
		},
		{
			name: "nonexistent_patch",
			file: &RemoteTextFile{
				LocalPath: copyFile,
				IsPatched: true,
				Patch: &Patch{
					PatchPath: filepath.Join(dir, "nonexistent.patch.txt"),
				},
			},
			wantError: "opening patch file:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := tt.file.RawPatchContent()
			if tt.wantError != "" {
				assert.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err, "reading patch content")
			defer content.Close()

			data, err := io.ReadAll(content)
			require.NoError(t, err, "reading patch data")
			assert.Equal(t, "patch content", string(data))
		})
	}
}
