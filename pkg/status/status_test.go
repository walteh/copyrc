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

package status

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/pkg/config"
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

func TestStatusManager(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, path string)
		check       func(t *testing.T, mgr *Manager)
		wantErr     bool
		errContains string
	}{
		{
			name: "new_status_file",
			check: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr.status, "status should not be nil")
				assert.NotNil(t, mgr.status.Files, "files map should not be nil")
				assert.Empty(t, mgr.status.Files, "files map should be empty")
			},
		},
		{
			name: "existing_status_file",
			setup: func(t *testing.T, path string) {
				status := &Status{
					Files: map[string]FileEntry{
						"test.txt": {
							Hash:         "abc123",
							Replacements: map[string]string{"foo": "bar"},
							Source:       "github.com/test/repo",
						},
					},
					CommitHash: "xyz789",
					Config: &config.Config{
						Provider: config.ProviderArgs{
							Repo: "github.com/test/repo",
							Ref:  "main",
						},
					},
				}
				data, err := json.MarshalIndent(status, "", "  ")
				require.NoError(t, err, "marshaling status should succeed")
				err = os.WriteFile(path, data, 0644)
				require.NoError(t, err, "writing status file should succeed")
			},
			check: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr.status, "status should not be nil")
				assert.NotEmpty(t, mgr.status.Files, "files map should not be empty")
				assert.Equal(t, "abc123", mgr.status.Files["test.txt"].Hash, "file hash should match")
				assert.Equal(t, "xyz789", mgr.status.CommitHash, "commit hash should match")
				assert.Equal(t, "github.com/test/repo", mgr.status.Config.Provider.Repo, "repo should match")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			statusPath := filepath.Join(tmpDir, ".copyrc.lock")

			// Setup test if needed
			if tt.setup != nil {
				tt.setup(t, statusPath)
			}

			// Create status manager
			mgr, err := New(statusPath)
			if tt.wantErr {
				require.Error(t, err, "New should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "New should succeed")
			if tt.check != nil {
				tt.check(t, mgr)
			}
		})
	}
}

func TestStatusOperations(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, mgr *Manager)
		operation   func(t *testing.T, mgr *Manager)
		check       func(t *testing.T, mgr *Manager)
		wantErr     bool
		errContains string
	}{
		{
			name: "update_file_entry",
			operation: func(t *testing.T, mgr *Manager) {
				err := mgr.Update(context.Background(), "test.txt", FileEntry{
					Hash:         "abc123",
					Replacements: map[string]string{"foo": "bar"},
					Source:       "github.com/test/repo",
				})
				require.NoError(t, err, "Update should succeed")
			},
			check: func(t *testing.T, mgr *Manager) {
				entry, ok := mgr.Get("test.txt")
				assert.True(t, ok, "file should exist")
				assert.Equal(t, "abc123", entry.Hash, "hash should match")
				assert.Equal(t, "bar", entry.Replacements["foo"], "replacement should match")
			},
		},
		{
			name: "update_commit_hash",
			operation: func(t *testing.T, mgr *Manager) {
				err := mgr.UpdateCommitHash(context.Background(), "xyz789")
				require.NoError(t, err, "UpdateCommitHash should succeed")
			},
			check: func(t *testing.T, mgr *Manager) {
				assert.Equal(t, "xyz789", mgr.GetCommitHash(), "commit hash should match")
			},
		},
		{
			name: "update_config",
			operation: func(t *testing.T, mgr *Manager) {
				err := mgr.UpdateConfig(context.Background(), &config.Config{
					Provider: config.ProviderArgs{
						Repo: "github.com/test/repo",
						Ref:  "main",
					},
				})
				require.NoError(t, err, "UpdateConfig should succeed")
			},
			check: func(t *testing.T, mgr *Manager) {
				cfg := mgr.GetConfig()
				assert.NotNil(t, cfg, "config should not be nil")
				assert.Equal(t, "github.com/test/repo", cfg.Provider.Repo, "repo should match")
			},
		},
		{
			name: "check_status",
			setup: func(t *testing.T, mgr *Manager) {
				err := mgr.UpdateCommitHash(context.Background(), "abc123")
				require.NoError(t, err, "UpdateCommitHash should succeed")
			},
			operation: func(t *testing.T, mgr *Manager) {
				p := &MockProvider{}
				p.On("GetCommitHash", mock.Anything, mock.Anything).Return("xyz789", nil)

				err := mgr.CheckStatus(context.Background(), &config.Config{}, p)
				assert.Error(t, err, "CheckStatus should fail due to hash mismatch")
				assert.Contains(t, err.Error(), "commit hash mismatch", "error should mention hash mismatch")

				p.AssertExpectations(t)
			},
		},
		{
			name: "clean_status",
			setup: func(t *testing.T, mgr *Manager) {
				err := mgr.Update(context.Background(), "test.txt", FileEntry{
					Hash: "abc123",
				})
				require.NoError(t, err, "Update should succeed")
			},
			operation: func(t *testing.T, mgr *Manager) {
				err := mgr.Clean()
				require.NoError(t, err, "Clean should succeed")
			},
			check: func(t *testing.T, mgr *Manager) {
				assert.Empty(t, mgr.List(), "status should be empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			statusPath := filepath.Join(tmpDir, ".copyrc.lock")

			// Create status manager
			mgr, err := New(statusPath)
			require.NoError(t, err, "New should succeed")

			// Setup test if needed
			if tt.setup != nil {
				tt.setup(t, mgr)
			}

			// Perform operation
			if tt.operation != nil {
				tt.operation(t, mgr)
			}

			// Check result
			if tt.check != nil {
				tt.check(t, mgr)
			}
		})
	}
}
