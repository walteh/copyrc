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

package operation_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/gen/mockery"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/operation"
	"github.com/walteh/copyrc/pkg/status"
)

// Helper function to check if a file should be ignored
func shouldIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}

func TestProcessFile(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		files    map[string]string
		setup    func(t *testing.T, tmpDir string)
		wantErr  bool
		validate func(t *testing.T, tmpDir string)
	}{
		{
			name: "process_file_with_replacements",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "test/repo",
					Ref:  "main",
					Path: "src",
				},
				Destination: "dst",
				Copy: &config.CopyArgs{
					Replacements: []config.Replacement{
						{Old: "old", New: "new"},
					},
				},
			},
			files: map[string]string{
				"file.txt": "old content here and old there",
			},
			validate: func(t *testing.T, tmpDir string) {
				content, err := os.ReadFile(filepath.Join(tmpDir, "dst", "file.txt"))
				require.NoError(t, err, "reading processed file should succeed")
				assert.Equal(t, "new content here and new there", string(content), "all occurrences should be replaced")
			},
		},
		{
			name: "process_file_specific_replacement",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "test/repo",
					Ref:  "main",
					Path: "src",
				},
				Destination: "dst",
				Copy: &config.CopyArgs{
					Replacements: []config.Replacement{
						{Old: "old", New: "new", File: stringPtr("other.txt")},
						{Old: "test", New: "replaced", File: stringPtr("file.txt")},
					},
				},
			},
			files: map[string]string{
				"file.txt":  "test content here and test there",
				"other.txt": "old content here and old there",
			},
			validate: func(t *testing.T, tmpDir string) {
				// Check file.txt - should be replaced
				content1, err := os.ReadFile(filepath.Join(tmpDir, "dst", "file.txt"))
				require.NoError(t, err, "reading file.txt should succeed")
				assert.Equal(t, "replaced content here and replaced there", string(content1), "replacements should only apply to specified file")

				// Check other.txt - should be replaced with its specific replacement
				content2, err := os.ReadFile(filepath.Join(tmpDir, "dst", "other.txt"))
				require.NoError(t, err, "reading other.txt should succeed")
				assert.Equal(t, "new content here and new there", string(content2), "replacements should only apply to specified file")
			},
		},
		{
			name: "ignore_file_pattern",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "test/repo",
					Ref:  "main",
					Path: "src",
				},
				Destination: "dst",
				Copy: &config.CopyArgs{
					IgnorePatterns: []string{"*.ignore", "tmp/*"},
				},
			},
			files: map[string]string{
				"test.ignore": "should not be copied",
				"tmp/file":    "should not be copied",
				"file.txt":    "should be copied",
			},
			validate: func(t *testing.T, tmpDir string) {
				// Verify ignored files don't exist
				_, err1 := os.Stat(filepath.Join(tmpDir, "dst", "test.ignore"))
				assert.True(t, os.IsNotExist(err1), "ignored file should not exist: test.ignore")

				_, err2 := os.Stat(filepath.Join(tmpDir, "dst", "tmp/file"))
				assert.True(t, os.IsNotExist(err2), "ignored file should not exist: tmp/file")

				// Verify non-ignored file exists and is correct
				content, err := os.ReadFile(filepath.Join(tmpDir, "dst", "file.txt"))
				require.NoError(t, err, "reading non-ignored file should succeed")
				assert.Equal(t, "should be copied", string(content), "non-ignored file should be copied correctly")
			},
		},
		{
			name: "provider_error",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "test/repo",
					Ref:  "main",
					Path: "src",
				},
				Destination: "dst",
			},
			files: map[string]string{
				"error.txt": "error content",
			},
			setup: func(t *testing.T, tmpDir string) {
				// Create a file that will cause permission error
				err := os.MkdirAll(filepath.Join(tmpDir, "dst"), 0444)
				require.NoError(t, err, "creating read-only directory")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir
			tmpDir := t.TempDir()

			// Run setup if provided
			if tt.setup != nil {
				tt.setup(t, tmpDir)
			}

			// Update config with temp dir
			tt.cfg.Destination = filepath.Join(tmpDir, tt.cfg.Destination)

			// Create destination dir and any parent directories
			err := os.MkdirAll(filepath.Dir(tt.cfg.Destination), 0755)
			require.NoError(t, err, "creating destination directory")

			// Set up mock provider
			p := mockery.NewMockProvider_provider(t)

			// Get list of non-ignored files
			files := make([]string, 0, len(tt.files))
			for path := range tt.files {
				if tt.cfg.Copy == nil || len(tt.cfg.Copy.IgnorePatterns) == 0 || !shouldIgnore(path, tt.cfg.Copy.IgnorePatterns) {
					files = append(files, path)
				}
			}

			// Mock ListFiles
			p.EXPECT().ListFiles(mock.Anything, tt.cfg.Provider).Return(files, nil).Once()

			// Mock GetCommitHash
			p.EXPECT().GetCommitHash(mock.Anything, tt.cfg.Provider).Return("test-hash", nil).Once()

			// Mock GetPermalink for each file
			for _, file := range files {
				p.EXPECT().GetPermalink(mock.Anything, tt.cfg.Provider, "test-hash", file).
					Return("https://test.com/"+file, nil).Once()
			}

			// Mock GetFile for each file
			for _, file := range files {
				p.EXPECT().GetFile(mock.Anything, tt.cfg.Provider, file).
					RunAndReturn(func(ctx context.Context, args config.ProviderArgs, path string) (io.ReadCloser, error) {
						if strings.Contains(path, "error") {
							return nil, errors.New("simulated provider error")
						}
						return io.NopCloser(strings.NewReader(tt.files[path])), nil
					}).Once()
			}

			// Create logger with test writer
			logger := zerolog.New(zerolog.NewTestWriter(t))

			// Create status manager
			statusMgr := status.NewManager(tt.cfg.Destination, status.NewDefaultFileFormatter())

			// Create operation
			op := &operation.BaseOperation{
				Config:    tt.cfg,
				Provider:  p,
				StatusMgr: statusMgr,
				Logger:    &logger,
			}

			// Execute operation
			err = op.Copy(context.Background())
			if tt.wantErr {
				assert.Error(t, err, "Execute should return error")
				return
			}
			assert.NoError(t, err, "Execute should succeed")

			// Run validation
			if tt.validate != nil {
				tt.validate(t, tmpDir)
			}
		})
	}
}

func TestProcessFiles(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		files    map[string]string
		wantErr  bool
		validate func(t *testing.T, tmpDir string)
	}{
		{
			name: "process_multiple_files",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "test/repo",
					Ref:  "main",
					Path: "src",
				},
				Destination: "dst",
				Copy: &config.CopyArgs{
					Replacements: []config.Replacement{
						{Old: "old", New: "new"},
					},
				},
			},
			files: map[string]string{
				"file1.txt": "old content 1",
				"file2.txt": "old content 2",
			},
			validate: func(t *testing.T, tmpDir string) {
				content1, err := os.ReadFile(filepath.Join(tmpDir, "dst", "file1.txt"))
				require.NoError(t, err, "reading file1 should succeed")
				assert.Equal(t, "new content 1", string(content1), "content1 should be replaced")

				content2, err := os.ReadFile(filepath.Join(tmpDir, "dst", "file2.txt"))
				require.NoError(t, err, "reading file2 should succeed")
				assert.Equal(t, "new content 2", string(content2), "content2 should be replaced")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir
			tmpDir := t.TempDir()

			// Update config with temp dir
			tt.cfg.Destination = filepath.Join(tmpDir, tt.cfg.Destination)

			// Create destination dir
			err := os.MkdirAll(tt.cfg.Destination, 0755)
			require.NoError(t, err, "creating destination directory")

			// Set up mock provider
			p := mockery.NewMockProvider_provider(t)
			files := make([]string, 0, len(tt.files))
			for path := range tt.files {
				files = append(files, path)
			}

			// Mock ListFiles
			p.EXPECT().ListFiles(mock.Anything, tt.cfg.Provider).Return(files, nil).Once()

			// Mock GetCommitHash
			p.EXPECT().GetCommitHash(mock.Anything, tt.cfg.Provider).Return("test-hash", nil).Once()

			// Mock GetPermalink for each file
			for _, file := range files {
				p.EXPECT().GetPermalink(mock.Anything, tt.cfg.Provider, "test-hash", file).
					Return("https://test.com/"+file, nil).Once()
			}

			// Mock GetFile for each file
			for _, file := range files {
				p.EXPECT().GetFile(mock.Anything, tt.cfg.Provider, file).
					RunAndReturn(func(ctx context.Context, args config.ProviderArgs, path string) (io.ReadCloser, error) {
						return io.NopCloser(strings.NewReader(tt.files[path])), nil
					}).Once()
			}

			// Create logger with test writer
			logger := zerolog.New(zerolog.NewTestWriter(t))

			// Create status manager
			statusMgr := status.NewManager(tt.cfg.Destination, status.NewDefaultFileFormatter())

			// Create operation
			op := &operation.BaseOperation{
				Config:    tt.cfg,
				Provider:  p,
				StatusMgr: statusMgr,
				Logger:    &logger,
			}

			// Process files
			err = op.Copy(context.Background())
			if tt.wantErr {
				assert.Error(t, err, "ProcessFiles should return error")
				return
			}
			assert.NoError(t, err, "ProcessFiles should succeed")

			// Run validation
			if tt.validate != nil {
				tt.validate(t, tmpDir)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
