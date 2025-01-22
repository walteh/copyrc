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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      LogConfig
		setup       func(t *testing.T, dir string)
		check       func(t *testing.T, mgr *Manager, dir string)
		wantErr     bool
		errContains string
	}{
		{
			name: "basic_config",
			config: LogConfig{
				Level:  "info",
				Format: "text",
				File:   "test.log",
			},
			check: func(t *testing.T, mgr *Manager, dir string) {
				cfg := mgr.GetLogConfig()
				require.NotNil(t, cfg, "config should not be nil")
				assert.Equal(t, "info", cfg.Level, "level should match")
				assert.Equal(t, "text", cfg.Format, "format should match")
				assert.Equal(t, filepath.Join(dir, "test.log"), cfg.File, "file should match")

				state := mgr.status.LogState
				require.NotNil(t, state, "state should not be nil")
				assert.NotZero(t, state.LastRotate, "last rotate should be set")
				assert.NotZero(t, state.LastCleanup, "last cleanup should be set")
				assert.Empty(t, state.LogFiles, "log files should be empty")
			},
		},
		{
			name: "rotation_config",
			config: LogConfig{
				Level:      "info",
				Format:     "text",
				File:       "rotate.log",
				Rotation:   true,
				MaxSize:    1,
				MaxAge:     7,
				MaxBackups: 3,
			},
			setup: func(t *testing.T, dir string) {
				// Create a large log file
				path := filepath.Join(dir, "rotate.log")
				data := make([]byte, 2*1024*1024) // 2MB
				require.NoError(t, os.WriteFile(path, data, 0644))
			},
			check: func(t *testing.T, mgr *Manager, dir string) {
				// Trigger rotation
				require.NoError(t, mgr.LogRotate(context.Background()))

				// Check that the original file is gone
				_, err := os.Stat(filepath.Join(dir, "rotate.log"))
				assert.True(t, os.IsNotExist(err), "original file should be rotated")

				// Check that we have a rotated file
				state := mgr.status.LogState
				require.NotNil(t, state, "state should not be nil")
				assert.NotEmpty(t, state.LogFiles, "should have rotated files")
				assert.NotZero(t, state.LastRotate, "last rotate should be set")
			},
		},
		{
			name: "cleanup_old_logs",
			config: LogConfig{
				Level:      "info",
				Format:     "text",
				File:       "cleanup.log",
				Rotation:   true,
				MaxAge:     1,
				MaxBackups: 2,
			},
			setup: func(t *testing.T, dir string) {
				// Create some old log files
				base := filepath.Join(dir, "cleanup.log")
				files := []struct {
					name    string
					age     time.Duration
					content string
				}{
					{".1", 48 * time.Hour, "old"},
					{".2", 24 * time.Hour, "newer"},
					{".3", time.Hour, "newest"},
				}

				for _, f := range files {
					path := base + f.name
					require.NoError(t, os.WriteFile(path, []byte(f.content), 0644))
					modTime := time.Now().Add(-f.age)
					require.NoError(t, os.Chtimes(path, modTime, modTime))
				}
			},
			check: func(t *testing.T, mgr *Manager, dir string) {
				// Trigger cleanup
				require.NoError(t, mgr.cleanupLogs(context.Background()))

				// Check that we only have the expected files
				state := mgr.status.LogState
				require.NotNil(t, state, "state should not be nil")
				assert.Len(t, state.LogFiles, 2, "should have 2 files")
				assert.NotZero(t, state.LastCleanup, "last cleanup should be set")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			dir := t.TempDir()
			tt.config.File = filepath.Join(dir, tt.config.File)

			// Create status manager
			statusPath := filepath.Join(dir, ".copyrc.lock")
			mgr, err := New(statusPath)
			require.NoError(t, err, "creating manager should succeed")

			// Setup test if needed
			if tt.setup != nil {
				tt.setup(t, dir)
			}

			// Update config
			err = mgr.UpdateLogConfig(context.Background(), tt.config)
			if tt.wantErr {
				require.Error(t, err, "UpdateLogConfig should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "UpdateLogConfig should succeed")
			if tt.check != nil {
				tt.check(t, mgr, dir)
			}
		})
	}
}

func TestLogState(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, mgr *Manager)
		check       func(t *testing.T, mgr *Manager)
		wantErr     bool
		errContains string
	}{
		{
			name: "basic_state",
			setup: func(t *testing.T, mgr *Manager) {
				state := &LogState{
					LastRotation: time.Now().Add(-24 * time.Hour),
					LastCleanup:  time.Now().Add(-48 * time.Hour),
					LogFiles: []string{
						"test1.log",
						"test2.log",
					},
				}
				require.NoError(t, mgr.updateLogState(context.Background(), state))
			},
			check: func(t *testing.T, mgr *Manager) {
				state := mgr.getLogState()
				require.NotNil(t, state, "state should not be nil")
				assert.Len(t, state.LogFiles, 2, "should have 2 log files")
				assert.True(t, state.LastRotation.Before(time.Now()), "last rotation should be in the past")
				assert.True(t, state.LastCleanup.Before(time.Now()), "last cleanup should be in the past")
			},
		},
		{
			name: "update_state",
			setup: func(t *testing.T, mgr *Manager) {
				// Set initial state
				state := &LogState{
					LastRotation: time.Now().Add(-24 * time.Hour),
					LastCleanup:  time.Now().Add(-48 * time.Hour),
					LogFiles: []string{
						"old1.log",
						"old2.log",
					},
				}
				require.NoError(t, mgr.updateLogState(context.Background(), state))

				// Update state
				newState := &LogState{
					LastRotation: time.Now(),
					LastCleanup:  time.Now(),
					LogFiles: []string{
						"new1.log",
						"new2.log",
						"new3.log",
					},
				}
				require.NoError(t, mgr.updateLogState(context.Background(), newState))
			},
			check: func(t *testing.T, mgr *Manager) {
				state := mgr.getLogState()
				require.NotNil(t, state, "state should not be nil")
				assert.Len(t, state.LogFiles, 3, "should have 3 log files")
				assert.Contains(t, state.LogFiles, "new1.log", "should contain new file")
				assert.NotContains(t, state.LogFiles, "old1.log", "should not contain old file")
			},
		},
		{
			name: "cleanup_state",
			setup: func(t *testing.T, mgr *Manager) {
				state := &LogState{
					LastRotation: time.Now().Add(-7 * 24 * time.Hour),
					LastCleanup:  time.Now().Add(-7 * 24 * time.Hour),
					LogFiles: []string{
						"old1.log",
						"old2.log",
						"old3.log",
					},
				}
				require.NoError(t, mgr.updateLogState(context.Background(), state))
			},
			check: func(t *testing.T, mgr *Manager) {
				// Trigger cleanup
				require.NoError(t, mgr.cleanupLogs(context.Background()))

				state := mgr.getLogState()
				require.NotNil(t, state, "state should not be nil")
				assert.True(t, state.LastCleanup.After(time.Now().Add(-time.Minute)), "last cleanup should be recent")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			dir := t.TempDir()

			// Create status manager
			statusPath := filepath.Join(dir, ".copyrc.lock")
			mgr, err := New(statusPath)
			require.NoError(t, err, "creating manager should succeed")

			// Setup test if needed
			if tt.setup != nil {
				tt.setup(t, mgr)
			}

			if tt.check != nil {
				tt.check(t, mgr)
			}
		})
	}
}
