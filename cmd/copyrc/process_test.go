// Copyright 2025 walteh LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
			cfg := &SingleConfig{
				Source: Source{
					Repo: "github.com/test/repo",
					Ref:  "main",
					Path: ".",
				},
				Destination: Destination{
					Path: t.TempDir(),
				},
				CopyArgs: &CopyEntry_Options{
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
			err := processCopy(ctx, mock, cfg.Source, cfg.Destination, cfg.CopyArgs, "test-hash", status, &mu, ProviderFile{
				Path: tt.file,
			})

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

func TestProcessFile_FilePatterns(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		filePatterns []string
		shouldCopy   bool
	}{
		{
			name:         "test_match_exact_pattern",
			file:         "README.md",
			filePatterns: []string{"README.md"},
			shouldCopy:   true,
		},
		{
			name:         "test_match_glob_pattern",
			file:         "src/main.go",
			filePatterns: []string{"**/*.go"},
			shouldCopy:   true,
		},
		{
			name:         "test_match_multiple_patterns",
			file:         "docs/api.md",
			filePatterns: []string{"*.go", "docs/*.md"},
			shouldCopy:   true,
		},
		{
			name:         "test_no_pattern_match",
			file:         "test.yaml",
			filePatterns: []string{"*.go", "*.md"},
			shouldCopy:   false,
		},
		{
			name:         "test_empty_patterns",
			file:         "test.go",
			filePatterns: []string{},
			shouldCopy:   true, // Should copy when no patterns specified
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock provider
			mock := NewMockProvider(t)
			mock.AddFile(tt.file, []byte("test content"))

			// Create config with file patterns
			cfg := &SingleConfig{
				Source: Source{
					Repo: "github.com/test/repo",
					Ref:  "main",
					Path: ".",
				},
				Destination: Destination{
					Path: t.TempDir(),
				},
				CopyArgs: &CopyEntry_Options{
					FilePatterns: tt.filePatterns,
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
			err := processCopy(ctx, mock, cfg.Source, cfg.Destination, cfg.CopyArgs, "test-hash", status, &mu, ProviderFile{
				Path: tt.file,
			})

			if tt.shouldCopy {
				require.NoError(t, err, "should not return error when processing file")
				assert.NotEmpty(t, status.CoppiedFiles, "should have copied files in status")
			} else {
				require.NoError(t, err, "should not return error when skipping file")
				assert.Empty(t, status.CoppiedFiles, "should not have any copied files in status")
			}
		})
	}
}

func TestProcessFile_PatternInteractions(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		filePatterns []string
		ignoreFiles  []string
		shouldCopy   bool
		description  string
	}{
		{
			name:         "test_direct_filename_pattern",
			file:         "specific.go",
			filePatterns: []string{"specific.go"},
			ignoreFiles:  []string{},
			shouldCopy:   true,
			description:  "Direct filename in patterns should match",
		},
		{
			name:         "test_direct_filename_ignore",
			file:         "ignore-me.go",
			filePatterns: []string{"*.go"},
			ignoreFiles:  []string{"ignore-me.go"},
			shouldCopy:   false,
			description:  "Direct filename in ignores should be skipped",
		},
		{
			name:         "test_ignore_takes_precedence",
			file:         "conflict.go",
			filePatterns: []string{"conflict.go", "*.go"},
			ignoreFiles:  []string{"conflict.go"},
			shouldCopy:   false,
			description:  "Ignore patterns should take precedence over include patterns",
		},
		{
			name:         "test_nested_direct_filename",
			file:         "src/internal/special.go",
			filePatterns: []string{"src/internal/special.go"},
			ignoreFiles:  []string{},
			shouldCopy:   true,
			description:  "Direct filename with path should match",
		},
		{
			name:         "test_nested_direct_filename_ignore",
			file:         "src/internal/ignore-me.go",
			filePatterns: []string{"src/internal/*.go"},
			ignoreFiles:  []string{"src/internal/ignore-me.go"},
			shouldCopy:   false,
			description:  "Direct filename with path in ignores should be skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock provider
			mock := NewMockProvider(t)
			mock.AddFile(tt.file, []byte("test content"))

			// Create config with patterns and ignores
			cfg := &SingleConfig{
				Source: Source{
					Repo: "github.com/test/repo",
					Ref:  "main",
					Path: ".",
				},
				Destination: Destination{
					Path: t.TempDir(),
				},
				CopyArgs: &CopyEntry_Options{
					FilePatterns: tt.filePatterns,
					IgnoreFiles:  tt.ignoreFiles,
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
			err := processCopy(ctx, mock, cfg.Source, cfg.Destination, cfg.CopyArgs, "test-hash", status, &mu, ProviderFile{
				Path: tt.file,
			})

			require.NoError(t, err, "should not return error")
			if tt.shouldCopy {
				assert.NotEmpty(t, status.CoppiedFiles, "should have copied files in status: %s", tt.description)
			} else {
				assert.Empty(t, status.CoppiedFiles, "should not have any copied files in status: %s", tt.description)
			}
		})
	}
}

func TestProcessFile_SingleFilePattern(t *testing.T) {
	// Create a mock provider with many files
	mock := NewMockProvider(t)
	for i := 0; i < 100; i++ {
		mock.AddFile(fmt.Sprintf("file%d.go", i), []byte("test content"))
	}
	// Add our target file
	targetFile := "src/internal/special.go"
	mock.AddFile(targetFile, []byte("special content"))

	// Create config to only copy the target file
	cfg := &SingleConfig{
		Source: Source{
			Repo: "github.com/test/repo",
			Ref:  "main",
			Path: ".",
		},
		Destination: Destination{
			Path: t.TempDir(),
		},
		CopyArgs: &CopyEntry_Options{
			FilePatterns: []string{targetFile}, // Only match this exact file
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
	err := processCopy(ctx, mock, cfg.Source, cfg.Destination, cfg.CopyArgs, "test-hash", status, &mu, ProviderFile{
		Path: targetFile,
	})

	require.NoError(t, err, "should not return error")
	assert.NotEmpty(t, status.CoppiedFiles, "should have copied the target file")
	assert.Len(t, status.CoppiedFiles, 1, "should have copied exactly one file")

	// Verify no other files were copied
	for file := range status.CoppiedFiles {
		assert.Equal(t, "src/internal/special.copy.go", file, "only the target file should be copied")
	}
}

func TestProcessDirectory_SingleFilePattern(t *testing.T) {
	// Create a mock provider with many files
	mock := NewMockProvider(t)
	for i := 0; i < 100; i++ {
		mock.AddFile(fmt.Sprintf("file%d.go", i), []byte("test content"))
	}
	// Add our target file
	targetFile := "src/internal/special.go"
	mock.AddFile(targetFile, []byte("special content"))

	// Create config to only copy the target file
	cfg := &SingleConfig{
		Source: Source{
			Repo: "github.com/test/repo",
			Ref:  "main",
			Path: ".",
		},
		Destination: Destination{
			Path: t.TempDir(),
		},
		CopyArgs: &CopyEntry_Options{
			FilePatterns: []string{targetFile}, // Only match this exact file
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
	err := processDirectory(ctx, mock, cfg, "test-hash", status, &mu)

	require.NoError(t, err, "should not return error")
	assert.NotEmpty(t, status.CoppiedFiles, "should have copied the target file")
	assert.Len(t, status.CoppiedFiles, 1, "should have copied exactly one file")

	// Verify no other files were copied
	for file := range status.CoppiedFiles {
		assert.Equal(t, "src/internal/special.copy.go", file, "only the target file should be copied")
	}
}

func TestProcessDirectory_MultipleFilePatterns(t *testing.T) {
	// Create a mock provider with many files
	mock := NewMockProvider(t)

	// Add some noise files
	for i := 0; i < 50; i++ {
		mock.AddFile(fmt.Sprintf("noise/file%d.txt", i), []byte("noise"))
	}

	// Add our target files
	targetFiles := []string{
		"src/main.go",
		"docs/README.md",
		"config/settings.yaml",
	}
	for _, file := range targetFiles {
		mock.AddFile(file, []byte(fmt.Sprintf("content of %s", file)))
	}

	// Create config to match specific patterns
	cfg := &SingleConfig{
		Source: Source{
			Repo: "github.com/test/repo",
			Ref:  "main",
			Path: ".",
		},
		Destination: Destination{
			Path: t.TempDir(),
		},
		CopyArgs: &CopyEntry_Options{
			FilePatterns: []string{
				"src/*.go",      // Match main.go
				"docs/*.md",     // Match README.md
				"config/*.yaml", // Match settings.yaml
			},
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
	err := processDirectory(ctx, mock, cfg, "test-hash", status, &mu)

	require.NoError(t, err, "should not return error")
	assert.Len(t, status.CoppiedFiles, len(targetFiles), "should have copied exactly the target files")

	// Verify each target file was copied
	expectedFiles := map[string]bool{
		"src/main.copy.go":          true,
		"docs/README.copy.md":       true,
		"config/settings.copy.yaml": true,
	}

	for file := range status.CoppiedFiles {
		assert.True(t, expectedFiles[file], "file %s should be in expected files", file)
		delete(expectedFiles, file)
	}
	assert.Empty(t, expectedFiles, "all expected files should have been found")
}

func TestProcessDirectory_RecursiveCopy(t *testing.T) {
	// Create a mock provider with nested directory structure
	mock := NewMockProvider(t)

	// Add files in nested directories
	nestedFiles := []string{
		"src/main.go",
		"src/internal/utils.go",
		"src/internal/deep/helper.go",
		"src/pkg/types.go",
	}
	for _, file := range nestedFiles {
		mock.AddFile(file, []byte(fmt.Sprintf("content of %s", file)))
	}

	// Add some files in root for noise
	mock.AddFile("README.md", []byte("readme content"))
	mock.AddFile("LICENSE", []byte("license content"))

	// Create config with recursive enabled
	cfg := &SingleConfig{
		Source: Source{
			Repo: "github.com/test/repo",
			Ref:  "main",
			Path: ".",
		},
		Destination: Destination{
			Path: t.TempDir(),
		},
		CopyArgs: &CopyEntry_Options{
			FilePatterns: []string{"src/**/*.go"}, // Match all .go files in src recursively
			Recursive:    true,                    // 📁 Enable recursive copying
		},
	}

	// Setup logger in context
	logger := NewDiscardDebugLogger(os.Stdout)
	ctx := NewLoggerInContext(context.Background(), logger)

	err := process(ctx, cfg, mock)

	status, err := loadStatusFile(filepath.Join(cfg.Destination.Path, ".copyrc.lock"))
	require.NoError(t, err, "should not return error")

	require.NoError(t, err, "should not return error")
	assert.Len(t, status.CoppiedFiles, len(nestedFiles), "should have copied all nested .go files")

	// Verify each nested file was copied
	expectedFiles := map[string]bool{
		"src/main.copy.go":                 true,
		"src/internal/utils.copy.go":       true,
		"src/internal/deep/helper.copy.go": true,
		"src/pkg/types.copy.go":            true,
	}

	for file := range status.CoppiedFiles {
		assert.True(t, expectedFiles[file], "file %s should be in expected files", file)
		delete(expectedFiles, file)
	}
	assert.Empty(t, expectedFiles, "all expected files should have been found")
}
