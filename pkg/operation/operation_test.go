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

package operation

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/pkg/config"
)

// MockProvider mocks the provider interface
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) ListFiles(ctx context.Context, args config.ProviderArgs) ([]string, error) {
	result := m.Called(ctx, args)
	return result.Get(0).([]string), result.Error(1)
}

func (m *MockProvider) GetFile(ctx context.Context, args config.ProviderArgs, path string) (io.ReadCloser, error) {
	result := m.Called(ctx, args, path)
	return result.Get(0).(io.ReadCloser), result.Error(1)
}

func (m *MockProvider) GetCommitHash(ctx context.Context, args config.ProviderArgs) (string, error) {
	result := m.Called(ctx, args)
	return result.String(0), result.Error(1)
}

func (m *MockProvider) GetPermalink(ctx context.Context, args config.ProviderArgs, commitHash string, file string) (string, error) {
	result := m.Called(ctx, args, commitHash, file)
	return result.String(0), result.Error(1)
}

func (m *MockProvider) GetSourceInfo(ctx context.Context, args config.ProviderArgs, commitHash string) (string, error) {
	result := m.Called(ctx, args, commitHash)
	return result.String(0), result.Error(1)
}

func (m *MockProvider) GetArchiveURL(ctx context.Context, args config.ProviderArgs) (string, error) {
	result := m.Called(ctx, args)
	return result.String(0), result.Error(1)
}

func TestProcessFile(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		files    map[string]string
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
				"file.txt": "old content",
			},
			validate: func(t *testing.T, tmpDir string) {
				content, err := os.ReadFile(filepath.Join(tmpDir, "dst", "file.txt"))
				require.NoError(t, err, "reading file should succeed")
				assert.Equal(t, "new content", string(content), "content should be replaced")
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
				"file.txt": "test content",
			},
			validate: func(t *testing.T, tmpDir string) {
				content, err := os.ReadFile(filepath.Join(tmpDir, "dst", "file.txt"))
				require.NoError(t, err, "reading file should succeed")
				assert.Equal(t, "replaced content", string(content), "content should be replaced")
			},
		},
		{
			name: "ignore_file",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "test/repo",
					Ref:  "main",
					Path: "src",
				},
				Destination: "dst",
				Copy: &config.CopyArgs{
					IgnorePatterns: []string{"ignored.txt"},
				},
			},
			files: map[string]string{
				"ignored.txt": "should not be copied",
			},
			validate: func(t *testing.T, tmpDir string) {
				_, err := os.Stat(filepath.Join(tmpDir, "dst", "ignored.txt"))
				assert.True(t, os.IsNotExist(err), "ignored file should not exist")
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
			p := &MockProvider{}
			for path, content := range tt.files {
				// Only set up expectations for non-ignored files
				if tt.cfg.Copy == nil || len(tt.cfg.Copy.IgnorePatterns) == 0 || !shouldIgnore(path, tt.cfg.Copy.IgnorePatterns) {
					// Create a new reader for each call to avoid EOF issues
					p.On("GetFile", mock.Anything, tt.cfg.Provider, path).Return(
						func(ctx context.Context, args config.ProviderArgs, path string) io.ReadCloser {
							return io.NopCloser(strings.NewReader(tt.files[path]))
						},
						func(ctx context.Context, args config.ProviderArgs, path string) error {
							return nil
						},
					)
				}
			}

			// Create logger
			var buf bytes.Buffer
			logger := zerolog.New(&buf)

			// Create manager
			mgr := New(tt.cfg, p, &logger)

			// Process each file
			for path := range tt.files {
				err := mgr.ProcessFile(context.Background(), path)
				if tt.wantErr {
					assert.Error(t, err, "ProcessFile should return error")
					return
				}
				assert.NoError(t, err, "ProcessFile should succeed")
			}

			// Run validation
			if tt.validate != nil {
				tt.validate(t, tmpDir)
			}

			// Verify all mock expectations
			p.AssertExpectations(t)
		})
	}
}

// Helper function to check if a file should be ignored
func shouldIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
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
			p := &MockProvider{}
			files := make([]string, 0, len(tt.files))
			for path, content := range tt.files {
				files = append(files, path)
				// Create a new reader for each call to avoid EOF issues
				p.On("GetFile", mock.Anything, tt.cfg.Provider, path).Return(
					func(ctx context.Context, args config.ProviderArgs, path string) io.ReadCloser {
						return io.NopCloser(strings.NewReader(tt.files[path]))
					},
					func(ctx context.Context, args config.ProviderArgs, path string) error {
						return nil
					},
				)
			}
			p.On("ListFiles", mock.Anything, tt.cfg.Provider).Return(files, nil).Once()

			// Create logger
			var buf bytes.Buffer
			logger := zerolog.New(&buf)

			// Create manager
			mgr := New(tt.cfg, p, &logger)

			// Process files
			err = mgr.ProcessFiles(context.Background(), files)
			if tt.wantErr {
				assert.Error(t, err, "ProcessFiles should return error")
				return
			}
			assert.NoError(t, err, "ProcessFiles should succeed")

			// Run validation
			if tt.validate != nil {
				tt.validate(t, tmpDir)
			}

			// Verify all mock expectations
			p.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
