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
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) *Handler
		wantErr     bool
		errContains string
		validate    func(t *testing.T)
	}{
		{
			name: "basic_run",
			setup: func(t *testing.T) *Handler {
				// Create temp config file
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				configContent := `
provider:
  repo: github.com/walteh/copyrc
  ref: main
  path: pkg/provider
destination: ` + filepath.Join(tmpDir, "dest") + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err, "writing config file")

				return &Handler{
					configFile: configPath,
					debug:      true,
				}
			},
			validate: func(t *testing.T) {
				// Add validation if needed
			},
		},
		{
			name: "invalid_config",
			setup: func(t *testing.T) *Handler {
				// Create temp config file with invalid content
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				configContent := `invalid: yaml: :`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err, "writing config file")

				return &Handler{
					configFile: configPath,
				}
			},
			wantErr:     true,
			errContains: "parsing config",
		},
		{
			name: "clean_run",
			setup: func(t *testing.T) *Handler {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				configContent := `
provider:
  repo: github.com/walteh/copyrc
  ref: main
  path: pkg/provider
destination: ` + filepath.Join(tmpDir, "dest") + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err, "writing config file")

				return &Handler{
					configFile: configPath,
					clean:      true,
				}
			},
		},
		{
			name: "status_check",
			setup: func(t *testing.T) *Handler {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				configContent := `
provider:
  repo: github.com/walteh/copyrc
  ref: main
  path: pkg/provider
destination: ` + filepath.Join(tmpDir, "dest") + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err, "writing config file")

				return &Handler{
					configFile: configPath,
					status:     true,
				}
			},
		},
	}

	// Skip if no GitHub token
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("GITHUB_TOKEN not set")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup(t)
			ctx := context.Background()

			// Set up logging
			logger := zerolog.New(zerolog.NewTestWriter(t)).Level(zerolog.DebugLevel)
			ctx = logger.WithContext(ctx)

			err := h.Run(ctx)
			if tt.wantErr {
				require.Error(t, err, "Run should return error")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				}
				return
			}

			require.NoError(t, err, "Run should succeed")
			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()
	require.NotNil(t, cmd, "command should not be nil")
	assert.Equal(t, "copyrc", cmd.Use, "command name should match")
	assert.NotEmpty(t, cmd.Short, "should have short description")
}
