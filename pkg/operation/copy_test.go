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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/gen/mockery"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/operation"
	"github.com/walteh/copyrc/pkg/status"
	"gitlab.com/tozd/go/errors"
)

// 🧪 createTestEnv creates a test environment
func createTestEnv(t *testing.T) (context.Context, *config.Config, *mockery.MockProvider_provider, *status.Manager, *zerolog.Logger) {
	// Create temp dir
	tmpDir := t.TempDir()

	// Create config
	cfg := &config.Config{
		Provider: config.ProviderArgs{
			Repo: "test/repo",
			Ref:  "main",
			Path: "src",
		},
		Destination: filepath.Join(tmpDir, "dst"),
		Copy: &config.CopyArgs{
			IgnorePatterns: []string{"*.ignore"},
		},
	}

	// Create mock provider
	mockProvider := mockery.NewMockProvider_provider(t)

	// Create logger
	logger := zerolog.New(zerolog.NewTestWriter(t))
	ctx := logger.WithContext(context.Background())

	// Create status manager
	statusMgr := status.NewManager(cfg.Destination, status.NewDefaultFileFormatter())

	return ctx, cfg, mockProvider, statusMgr, &logger
}

// 🧪 TestCopyOperation tests the copy operation
func TestCopyOperation(t *testing.T) {
	ctx, cfg, mockProvider, statusMgr, logger := createTestEnv(t)

	// Create test files
	files := []string{
		"test.txt",
		"test.go",
		"test.ignore",
	}

	// Set up mock provider
	mockProvider.EXPECT().ListFiles(ctx, cfg.Provider).Return(files, nil)
	mockProvider.EXPECT().GetCommitHash(ctx, cfg.Provider).Return("abc123", nil)

	// Set up expectations for each file
	for _, file := range files {
		if strings.HasSuffix(file, ".ignore") {
			continue // Ignored files should be skipped
		}

		permalink := "https://test.com/" + file
		mockProvider.EXPECT().GetPermalink(ctx, cfg.Provider, "abc123", file).Return(permalink, nil)
		mockProvider.EXPECT().GetFile(ctx, cfg.Provider, file).
			RunAndReturn(func(ctx context.Context, args config.ProviderArgs, path string) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("test content")), nil
			})
	}

	// Create and run copy operation
	op := operation.NewCopyOperation(operation.Options{
		Config:    cfg,
		Provider:  mockProvider,
		StatusMgr: statusMgr,
		Logger:    logger,
	})
	err := op.Execute(ctx)
	require.NoError(t, err)
}

// 🧪 TestCopyOperationErrors tests error handling in copy operation
func TestCopyOperationErrors(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*testing.T, context.Context, *config.Config, *mockery.MockProvider_provider)
		expectedError string
	}{
		{
			name: "list_files_error",
			setupMocks: func(t *testing.T, ctx context.Context, cfg *config.Config, mockProvider *mockery.MockProvider_provider) {
				mockProvider.EXPECT().ListFiles(ctx, cfg.Provider).Return(nil, errors.New("list error"))
			},
			expectedError: "listing files: list error",
		},
		{
			name: "get_commit_hash_error",
			setupMocks: func(t *testing.T, ctx context.Context, cfg *config.Config, mockProvider *mockery.MockProvider_provider) {
				mockProvider.EXPECT().ListFiles(ctx, cfg.Provider).Return([]string{"test.txt"}, nil)
				mockProvider.EXPECT().GetCommitHash(ctx, cfg.Provider).Return("", errors.New("hash error"))
			},
			expectedError: "getting commit hash: hash error",
		},
		{
			name: "get_permalink_error",
			setupMocks: func(t *testing.T, ctx context.Context, cfg *config.Config, mockProvider *mockery.MockProvider_provider) {
				mockProvider.EXPECT().ListFiles(ctx, cfg.Provider).Return([]string{"test.txt"}, nil)
				mockProvider.EXPECT().GetCommitHash(ctx, cfg.Provider).Return("abc123", nil)
				mockProvider.EXPECT().GetPermalink(ctx, cfg.Provider, "abc123", "test.txt").Return("", errors.New("permalink error"))
			},
			expectedError: "processing file test.txt: getting file content: getting permalink: permalink error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cfg, mockProvider, statusMgr, logger := createTestEnv(t)
			tt.setupMocks(t, ctx, cfg, mockProvider)

			op := operation.NewCopyOperation(operation.Options{
				Config:    cfg,
				Provider:  mockProvider,
				StatusMgr: statusMgr,
				Logger:    logger,
			})
			err := op.Execute(ctx)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// 🧪 TestCopyOperationReplacements tests text replacements in copy operation
func TestCopyOperationReplacements(t *testing.T) {
	ctx, cfg, mockProvider, statusMgr, logger := createTestEnv(t)

	// Add test replacement
	cfg.Copy.Replacements = []config.Replacement{
		{
			Old: "old",
			New: "new",
		},
	}

	// Set up mock provider
	mockProvider.EXPECT().ListFiles(ctx, cfg.Provider).Return([]string{"test.txt"}, nil)
	mockProvider.EXPECT().GetCommitHash(ctx, cfg.Provider).Return("abc123", nil)
	mockProvider.EXPECT().GetPermalink(ctx, cfg.Provider, "abc123", "test.txt").Return("https://test.com/test.txt", nil)
	mockProvider.EXPECT().GetFile(ctx, cfg.Provider, "test.txt").
		RunAndReturn(func(ctx context.Context, args config.ProviderArgs, path string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("this is old text")), nil
		})

	// Create and run copy operation
	op := operation.NewCopyOperation(operation.Options{
		Config:    cfg,
		Provider:  mockProvider,
		StatusMgr: statusMgr,
		Logger:    logger,
	})
	err := op.Execute(ctx)
	require.NoError(t, err)

	// Verify content was replaced
	content, err := os.ReadFile(filepath.Join(cfg.Destination, "test.txt"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "this is new text")
}

// 🚗 mockTransport implements http.RoundTripper for testing
type mockTransport struct {
	response *http.Response
	err      error
}

func (t *mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return t.response, t.err
}
