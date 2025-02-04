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
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedLogging(t *testing.T) {
	// Create a buffer to capture output

	t.Run("success_message", func(t *testing.T) {
		logger := newTestLogger(t)
		logger.Success("Operation completed")
		assert.Contains(t, logger.CopyOfCurrentConsoleOutputInTest(), "✅")
		assert.Contains(t, logger.CopyOfCurrentConsoleOutputInTest(), "Operation completed")
	})

	t.Run("info_message", func(t *testing.T) {
		logger := newTestLogger(t)
		logger.Info("Processing files")
		assert.Contains(t, logger.CopyOfCurrentConsoleOutputInTest(), "ℹ️")
		assert.Contains(t, logger.CopyOfCurrentConsoleOutputInTest(), "Processing files")
	})

	t.Run("warning_message", func(t *testing.T) {
		logger := newTestLogger(t)
		logger.Warning("Proceed with caution")
		assert.Contains(t, logger.CopyOfCurrentConsoleOutputInTest(), "⚠️")
		assert.Contains(t, logger.CopyOfCurrentConsoleOutputInTest(), "Proceed with caution")
	})

	t.Run("error_message", func(t *testing.T) {
		logger := newTestLogger(t)
		logger.Error("Something went wrong")
		assert.Contains(t, logger.CopyOfCurrentConsoleOutputInTest(), "❌")
		assert.Contains(t, logger.CopyOfCurrentConsoleOutputInTest(), "Something went wrong")
	})
}

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name        string
		repo        string
		ref         string
		path        string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_github_repo",
			repo: "github.com/org/repo",
			ref:  "main",
			path: "path/to/files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewGithubProvider()
			require.NoError(t, err, "unexpected error")
			require.NotNil(t, provider, "provider should not be nil")
		})
	}
}

func TestGithubProvider(t *testing.T) {
	provider, err := NewGithubProvider()
	require.NoError(t, err, "creating provider")

	t.Run("GetSourceInfo", func(t *testing.T) {
		info, err := provider.GetSourceInfo(context.Background(), Source{
			Repo: "github.com/org/repo",
			Ref:  "main",
			Path: "path/to/files",
		}, "abc123")
		require.NoError(t, err, "getting source info")
		assert.Equal(t, "github.com/org/repo@abc123", info)
	})

	t.Run("GetPermalink", func(t *testing.T) {
		link, err := provider.GetPermalink(context.Background(), Source{
			Repo: "github.com/org/repo",
			Ref:  "main",
			Path: "path/to/files",
		}, "abc123", "file.go")
		require.NoError(t, err, "getting permalink")
		assert.Equal(t, "https://raw.githubusercontent.com/org/repo/abc123/file.go", link)
	})
}

func TestNewConfigFromInput(t *testing.T) {
	tests := []struct {
		name        string
		input       Input
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_input",
			input: Input{
				SrcRepo:  "github.com/org/repo",
				SrcRef:   "main",
				SrcPath:  "path/to/files",
				DestPath: "/tmp/dest",
				Replacements: []string{
					"old:new",
					"foo:bar",
				},
				IgnoreFiles: []string{
					"*.tmp",
					"*.bak",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock provider for testing
			mock := NewMockProvider(t)

			cfg, err := NewConfigFromInput(tt.input, mock)
			require.NoError(t, err, "unexpected error")
			require.NotNil(t, cfg, "config should not be nil")
			assert.Equal(t, tt.input.DestPath, cfg.Destination.Path)
			assert.Len(t, cfg.CopyArgs.Replacements, len(tt.input.Replacements))
			assert.Len(t, cfg.CopyArgs.IgnoreFiles, len(tt.input.IgnoreFiles))
		})
	}
}

func TestProcessFile(t *testing.T) {
	// Setup mock provider with test files
	mock := NewMockProvider(t)
	mock.AddFile("test.go", []byte(`package foo

func Bar() {}`))
	mock.AddFile("other.go", []byte(`package foo

func Other() {}`))

	args := Source{
		Repo: mock.GetFullRepo(),
		Ref:  mock.ref,
		Path: mock.path,
	}

	cfg := &SingleConfig{
		Source:      args,
		Destination: Destination{Path: t.TempDir()},
		CopyArgs: &CopyEntry_Options{
			Replacements: []Replacement{
				{Old: "Bar", New: "Baz"},
			},
			IgnoreFiles: []string{"*.tmp", "*.bak"},
		},
	}

	// Initialize status
	status := &StatusFile{
		CoppiedFiles: make(map[string]StatusEntry),
	}

	t.Run("normal file", func(t *testing.T) {
		ctx := context.Background()
		logger := newTestLogger(t)
		ctx = NewLoggerInContext(ctx, logger)

		// Process the file
		var mu sync.Mutex
		err := processCopy(ctx, mock, cfg.Source, cfg.Destination, cfg.CopyArgs, mock.commitHash, status, &mu, ProviderFile{Path: "test.go"})
		require.NoError(t, err)

		// Verify the output file
		content, err := os.ReadFile(filepath.Join(cfg.Destination.Path, "test.copy.go"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "func Baz()")
		assert.Contains(t, string(content), "generated by copyrc. DO NOT EDIT.")

		// Verify status entry
		entry, ok := status.CoppiedFiles["test.copy.go"]
		require.True(t, ok, "status entry should exist")
		assert.Equal(t, "test.copy.go", entry.File)
		assert.Equal(t, "mock@abc123", entry.Source)
	})

	t.Run("clean destination", func(t *testing.T) {
		ctx := context.Background()
		logger := newTestLogger(t)
		ctx = NewLoggerInContext(ctx, logger)

		// Create test files
		dir := t.TempDir()
		files := []string{
			"test.copy.go",
			"regular.go",
		}

		st := &StatusFile{
			CoppiedFiles:   make(map[string]StatusEntry),
			GeneratedFiles: make(map[string]GeneratedFileEntry),
		}
		for _, f := range files {
			_, err := writeFile(ctx, WriteFileOpts{
				Path:       filepath.Join(dir, f),
				Contents:   []byte("content"),
				StatusFile: st,
			})
			require.NoError(t, err)
		}

		status.CoppiedFiles["test.copy.go"] = StatusEntry{
			File:        "test.copy.go",
			Source:      "mock@abc123",
			Permalink:   "mock://test.go@abc123",
			LastUpdated: time.Now().UTC(),
			Changes:     []string{"test change"},
		}

		// Clean the directory
		err := cleanDestination(ctx, status, dir)
		require.NoError(t, err)

		require.NoFileExists(t, filepath.Join(dir, "test.copy.go"))
		require.FileExists(t, filepath.Join(dir, "regular.go"))
		require.NoFileExists(t, filepath.Join(dir, ".copyrc.lock"))
	})

	t.Run("status check", func(t *testing.T) {
		dir := t.TempDir()
		ctx := context.Background()
		logger := newTestLogger(t)
		ctx = NewLoggerInContext(ctx, logger)

		// Create initial status
		status := &StatusFile{
			CommitHash: mock.commitHash,
			Ref:        mock.ref,
			Args: StatusFileArgs{
				SrcRepo:  args.Repo,
				SrcRef:   args.Ref,
				SrcPath:  args.Path,
				CopyArgs: &CopyEntry_Options{},
			},
			CoppiedFiles: make(map[string]StatusEntry),
		}
		require.NoError(t, writeStatusFile(ctx, status, dir))

		// Test with same commit hash
		cfg := &SingleConfig{
			Source:      args,
			Destination: Destination{Path: dir},
			Flags:       FlagsBlock{RemoteStatus: true},
			CopyArgs:    &CopyEntry_Options{},
		}
		err := process(ctx, cfg, mock)
		require.NoError(t, err)

		// Test with different commit hash
		status.CommitHash = "different"
		require.NoError(t, writeStatusFile(ctx, status, dir))
		err = process(ctx, cfg, mock)
		assert.Error(t, err)
	})

	t.Run("local_status_check", func(t *testing.T) {
		dir := t.TempDir()
		ctx := context.Background()
		logger := newTestLogger(t)
		ctx = NewLoggerInContext(ctx, logger)

		// Create initial status
		status := &StatusFile{
			CoppiedFiles: make(map[string]StatusEntry),
			Args: StatusFileArgs{
				SrcRepo: args.Repo,
				SrcRef:  args.Ref,
				SrcPath: args.Path,
				CopyArgs: &CopyEntry_Options{
					Replacements: []Replacement{
						{Old: "Bar", New: "Baz"},
					},
					IgnoreFiles: []string{"*.tmp", "*.bak"},
				},
			},
		}
		require.NoError(t, writeStatusFile(ctx, status, dir))

		// Test with same arguments
		cfg := &SingleConfig{
			Source:      args,
			Destination: Destination{Path: dir},
			Flags:       FlagsBlock{Status: true},
			CopyArgs: &CopyEntry_Options{
				Replacements: []Replacement{
					{Old: "Bar", New: "Baz"},
				},
				IgnoreFiles: []string{"*.tmp", "*.bak"},
			},
		}
		err := process(ctx, cfg, mock)
		require.NoError(t, err)

		// Test with different arguments
		cfg.CopyArgs.Replacements[0].New = "Different"
		err = process(ctx, cfg, mock)
		assert.Error(t, err, "should error when configuration changes")
	})
}
