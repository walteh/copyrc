package main

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessFile_IgnoreFiles(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		ignoreFiles []string
		shouldSkip  bool
	}{
		{
			name:        "test_ignore_exact_match",
			file:        "README.md",
			ignoreFiles: []string{"README.md"},
			shouldSkip:  true,
		},
		{
			name:        "test_ignore_glob_pattern",
			file:        "test.yaml",
			ignoreFiles: []string{"*.yaml"},
			shouldSkip:  true,
		},
		{
			name:        "test_ignore_multiple_patterns",
			file:        "test.yaml",
			ignoreFiles: []string{"*.md", "*.yaml", "*.json"},
			shouldSkip:  true,
		},
		{
			name:        "test_no_ignore_match",
			file:        "main.go",
			ignoreFiles: []string{"*.md", "*.yaml", "*.json"},
			shouldSkip:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock provider
			mock := NewMockProvider(t)
			mock.AddFile(tt.file, []byte("test content"))

			// Create config with ignore patterns
			cfg := &Config{
				ProviderArgs: ProviderArgs{
					Repo: "github.com/test/repo",
					Ref:  "main",
					Path: ".",
				},
				DestPath: t.TempDir(),
				CopyArgs: &ConfigCopyArgs{
					IgnoreFiles: tt.ignoreFiles,
				},
			}

			// Create status file
			status := &StatusFile{
				CoppiedFiles:   make(map[string]StatusEntry),
				GeneratedFiles: make(map[string]GeneratedFileEntry),
			}

			// Setup logger in context
			logger := NewDiscardDebugLogger(os.Stdout)
			ctx := NewLoggerInContext(context.Background(), logger)

			var mu sync.Mutex
			err := processFile(ctx, mock, cfg, tt.file, "test-hash", status, &mu, cfg.DestPath)

			if tt.shouldSkip {
				// If file should be ignored, we expect no error and no file in status
				require.NoError(t, err, "should not return error when ignoring file")
				assert.Empty(t, status.CoppiedFiles, "should not have any copied files in status")
			} else {
				// If file should not be ignored, we expect the file to be processed
				require.NoError(t, err, "should not return error when processing file")
				assert.NotEmpty(t, status.CoppiedFiles, "should have copied files in status")
			}
		})
	}
}
