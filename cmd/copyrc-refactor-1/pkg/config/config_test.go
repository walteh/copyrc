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

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		wantErr     bool
		errContains string
		check       func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid_config",
			config: `
provider:
  repo: github.com/walteh/copyrc
  ref: main
  path: pkg/provider
destination: /tmp/copyrc
copy:
  replacements:
    - old: foo
      new: bar
    - old: baz
      new: qux
      file: specific.go
  ignore_patterns:
    - "*.tmp"
    - "*.log"
async: true
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "github.com/walteh/copyrc", cfg.Provider.Repo, "repo should match")
				assert.Equal(t, "main", cfg.Provider.Ref, "ref should match")
				assert.Equal(t, "pkg/provider", cfg.Provider.Path, "path should match")
				assert.Equal(t, "/tmp/copyrc", cfg.Destination, "destination should match")
				assert.NotNil(t, cfg.Copy, "copy should not be nil")
				assert.Len(t, cfg.Copy.Replacements, 2, "should have 2 replacements")
				assert.Equal(t, "foo", cfg.Copy.Replacements[0].Old, "first replacement old should match")
				assert.Equal(t, "bar", cfg.Copy.Replacements[0].New, "first replacement new should match")
				assert.Nil(t, cfg.Copy.Replacements[0].File, "first replacement file should be nil")
				assert.Equal(t, "baz", cfg.Copy.Replacements[1].Old, "second replacement old should match")
				assert.Equal(t, "qux", cfg.Copy.Replacements[1].New, "second replacement new should match")
				assert.NotNil(t, cfg.Copy.Replacements[1].File, "second replacement file should not be nil")
				assert.Equal(t, "specific.go", *cfg.Copy.Replacements[1].File, "second replacement file should match")
				assert.Len(t, cfg.Copy.IgnorePatterns, 2, "should have 2 ignore patterns")
				assert.Equal(t, "*.tmp", cfg.Copy.IgnorePatterns[0], "first ignore pattern should match")
				assert.Equal(t, "*.log", cfg.Copy.IgnorePatterns[1], "second ignore pattern should match")
				assert.True(t, cfg.Async, "async should be true")
			},
		},
		{
			name: "minimal_config",
			config: `
provider:
  repo: github.com/walteh/copyrc
  path: pkg/provider
destination: /tmp/copyrc
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "github.com/walteh/copyrc", cfg.Provider.Repo, "repo should match")
				assert.Equal(t, "main", cfg.Provider.Ref, "ref should have default value")
				assert.Equal(t, "pkg/provider", cfg.Provider.Path, "path should match")
				assert.Equal(t, "/tmp/copyrc", cfg.Destination, "destination should match")
				assert.Nil(t, cfg.Copy, "copy should be nil")
				assert.False(t, cfg.Async, "async should be false")
			},
		},
		{
			name: "missing_required_repo",
			config: `
provider:
  path: pkg/provider
destination: /tmp/copyrc
`,
			wantErr:     true,
			errContains: "provider.repo is required",
		},
		{
			name: "missing_required_path",
			config: `
provider:
  repo: github.com/walteh/copyrc
destination: /tmp/copyrc
`,
			wantErr:     true,
			errContains: "provider.path is required",
		},
		{
			name: "missing_required_destination",
			config: `
provider:
  repo: github.com/walteh/copyrc
  path: pkg/provider
`,
			wantErr:     true,
			errContains: "destination is required",
		},
	}

	ctx := zerolog.New(os.Stderr).WithContext(context.Background())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.config), 0644)
			require.NoError(t, err, "writing config file should succeed")

			// Load config
			cfg, err := Load(ctx, configPath)
			if tt.wantErr {
				require.Error(t, err, "Load should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "Load should succeed")
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			name: "full_config",
			cfg: &Config{
				Provider: ProviderArgs{
					Repo: "github.com/walteh/copyrc",
					Ref:  "main",
					Path: "pkg/provider",
				},
				Destination: "/tmp/copyrc",
			},
			want: "github.com/walteh/copyrc@main:pkg/provider -> /tmp/copyrc",
		},
		{
			name: "minimal_config",
			cfg: &Config{
				Provider: ProviderArgs{
					Repo: "github.com/walteh/copyrc",
					Path: "pkg/provider",
				},
				Destination: "/tmp/copyrc",
			},
			want: "github.com/walteh/copyrc@main:pkg/provider -> /tmp/copyrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.String()
			assert.Equal(t, tt.want, got, "String() should match")
		})
	}
}
