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
	"github.com/walteh/copyrc/pkg/log"
)

// 🔧 MockProvider is a mock implementation of the provider.Provider interface
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
		name        string
		cfg         *config.Config
		path        string
		content     string
		wantErr     bool
		errContains string
		check       func(t *testing.T, destPath string)
	}{
		{
			name: "process_file_with_replacements",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "github.com/walteh/copyrc",
					Ref:  "main",
					Path: "pkg/provider",
				},
				Destination: t.TempDir(),
				Copy: &config.CopyArgs{
					Replacements: []config.Replacement{
						{Old: "foo", New: "bar"},
						{Old: "baz", New: "qux"},
					},
				},
			},
			path:    "test.txt",
			content: "This is a foo test with baz content",
			check: func(t *testing.T, destPath string) {
				content, err := os.ReadFile(destPath)
				require.NoError(t, err, "reading file should succeed")
				assert.Equal(t, "This is a bar test with qux content", string(content), "content should be replaced")
			},
		},
		{
			name: "process_file_with_specific_replacement",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "github.com/walteh/copyrc",
					Ref:  "main",
					Path: "pkg/provider",
				},
				Destination: t.TempDir(),
				Copy: &config.CopyArgs{
					Replacements: []config.Replacement{
						{Old: "foo", New: "bar"},
						{Old: "baz", New: "qux", File: stringPtr("other.txt")},
					},
				},
			},
			path:    "test.txt",
			content: "This is a foo test with baz content",
			check: func(t *testing.T, destPath string) {
				content, err := os.ReadFile(destPath)
				require.NoError(t, err, "reading file should succeed")
				assert.Equal(t, "This is a bar test with baz content", string(content), "only matching replacements should be applied")
			},
		},
		{
			name: "process_file_ignored",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "github.com/walteh/copyrc",
					Ref:  "main",
					Path: "pkg/provider",
				},
				Destination: t.TempDir(),
				Copy: &config.CopyArgs{
					IgnoreFiles: []string{"*.txt"},
				},
			},
			path:    "test.txt",
			content: "This is a test",
			check: func(t *testing.T, destPath string) {
				_, err := os.Stat(destPath)
				assert.True(t, os.IsNotExist(err), "file should not exist")
			},
		},
	}

	ctx := zerolog.New(os.Stderr).WithContext(context.Background())
	logger := log.New(io.Discard, zerolog.InfoLevel)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock provider
			p := &MockProvider{}
			p.On("GetFile", mock.Anything, tt.cfg.Provider, tt.path).Return(
				io.NopCloser(strings.NewReader(tt.content)),
				nil,
			)

			// Create operation manager
			mgr := New(tt.cfg, p, logger)

			// Process file
			err := mgr.ProcessFile(ctx, tt.path)
			if tt.wantErr {
				require.Error(t, err, "ProcessFile should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "ProcessFile should succeed")
			if tt.check != nil {
				tt.check(t, filepath.Join(tt.cfg.Destination, tt.path))
			}

			p.AssertExpectations(t)
		})
	}
}

func TestProcessFiles(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		files       []string
		contents    map[string]string
		wantErr     bool
		errContains string
		check       func(t *testing.T, destDir string)
	}{
		{
			name: "process_multiple_files",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "github.com/walteh/copyrc",
					Ref:  "main",
					Path: "pkg/provider",
				},
				Destination: t.TempDir(),
				Copy: &config.CopyArgs{
					Replacements: []config.Replacement{
						{Old: "foo", New: "bar"},
					},
				},
			},
			files: []string{"test1.txt", "test2.txt"},
			contents: map[string]string{
				"test1.txt": "This is foo test 1",
				"test2.txt": "This is foo test 2",
			},
			check: func(t *testing.T, destDir string) {
				content1, err := os.ReadFile(filepath.Join(destDir, "test1.txt"))
				require.NoError(t, err, "reading file 1 should succeed")
				assert.Equal(t, "This is bar test 1", string(content1), "content 1 should be replaced")

				content2, err := os.ReadFile(filepath.Join(destDir, "test2.txt"))
				require.NoError(t, err, "reading file 2 should succeed")
				assert.Equal(t, "This is bar test 2", string(content2), "content 2 should be replaced")
			},
		},
		{
			name: "process_files_async",
			cfg: &config.Config{
				Provider: config.ProviderArgs{
					Repo: "github.com/walteh/copyrc",
					Ref:  "main",
					Path: "pkg/provider",
				},
				Destination: t.TempDir(),
				Copy: &config.CopyArgs{
					Replacements: []config.Replacement{
						{Old: "foo", New: "bar"},
					},
				},
				Async: true,
			},
			files: []string{"test1.txt", "test2.txt"},
			contents: map[string]string{
				"test1.txt": "This is foo test 1",
				"test2.txt": "This is foo test 2",
			},
			check: func(t *testing.T, destDir string) {
				content1, err := os.ReadFile(filepath.Join(destDir, "test1.txt"))
				require.NoError(t, err, "reading file 1 should succeed")
				assert.Equal(t, "This is bar test 1", string(content1), "content 1 should be replaced")

				content2, err := os.ReadFile(filepath.Join(destDir, "test2.txt"))
				require.NoError(t, err, "reading file 2 should succeed")
				assert.Equal(t, "This is bar test 2", string(content2), "content 2 should be replaced")
			},
		},
	}

	ctx := zerolog.New(os.Stderr).WithContext(context.Background())
	logger := log.New(io.Discard, zerolog.InfoLevel)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock provider
			p := &MockProvider{}
			for path, content := range tt.contents {
				p.On("GetFile", mock.Anything, tt.cfg.Provider, path).Return(
					io.NopCloser(strings.NewReader(content)),
					nil,
				)
			}

			// Create operation manager
			mgr := New(tt.cfg, p, logger)

			// Process files
			err := mgr.ProcessFiles(ctx, tt.files)
			if tt.wantErr {
				require.Error(t, err, "ProcessFiles should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "ProcessFiles should succeed")
			if tt.check != nil {
				tt.check(t, tt.cfg.Destination)
			}

			p.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
