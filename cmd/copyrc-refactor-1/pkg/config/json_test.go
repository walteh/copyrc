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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 🧪 TestJSONParsing tests JSON config parsing
func TestJSONParsing(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		wantErr     bool
		errContains string
		check       func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid_minimal_json",
			config: `{
				"provider": {
					"repo": "github.com/walteh/copyrc",
					"path": "pkg/provider"
				},
				"destination": "/tmp/copyrc"
			}`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "github.com/walteh/copyrc", cfg.Provider.Repo)
				assert.Equal(t, "main", cfg.Provider.Ref) // default value
				assert.Equal(t, "pkg/provider", cfg.Provider.Path)
				assert.Equal(t, "/tmp/copyrc", cfg.Destination)
				assert.Nil(t, cfg.Copy)
			},
		},
		{
			name: "valid_full_json",
			config: `{
				"provider": {
					"repo": "github.com/walteh/copyrc",
					"ref": "develop",
					"path": "pkg/provider"
				},
				"destination": "/tmp/copyrc",
				"copy": {
					"replacements": [
						{
							"old": "foo",
							"new": "bar"
						},
						{
							"old": "baz",
							"new": "qux",
							"file": "specific.go"
						}
					],
					"ignore_patterns": [
						"*.tmp",
						"*.log"
					]
				},
				"go_embed": true,
				"clean": true,
				"status": true,
				"remote_status": true,
				"force": true,
				"async": true
			}`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "github.com/walteh/copyrc", cfg.Provider.Repo)
				assert.Equal(t, "develop", cfg.Provider.Ref)
				assert.Equal(t, "pkg/provider", cfg.Provider.Path)
				assert.Equal(t, "/tmp/copyrc", cfg.Destination)
				require.NotNil(t, cfg.Copy)
				assert.Len(t, cfg.Copy.Replacements, 2)
				assert.Equal(t, "foo", cfg.Copy.Replacements[0].Old)
				assert.Equal(t, "bar", cfg.Copy.Replacements[0].New)
				assert.Nil(t, cfg.Copy.Replacements[0].File)
				assert.Equal(t, "baz", cfg.Copy.Replacements[1].Old)
				assert.Equal(t, "qux", cfg.Copy.Replacements[1].New)
				require.NotNil(t, cfg.Copy.Replacements[1].File)
				assert.Equal(t, "specific.go", *cfg.Copy.Replacements[1].File)
				assert.Equal(t, []string{"*.tmp", "*.log"}, cfg.Copy.IgnorePatterns)
				assert.True(t, cfg.GoEmbed)
				assert.True(t, cfg.Clean)
				assert.True(t, cfg.Status)
				assert.True(t, cfg.RemoteStatus)
				assert.True(t, cfg.Force)
				assert.True(t, cfg.Async)
			},
		},
		{
			name: "invalid_json_syntax",
			config: `{
				"provider": {
					"repo": "github.com/walteh/copyrc",
					"path": "pkg/provider",
				},
				"destination": "/tmp/copyrc"
			}`,
			wantErr:     true,
			errContains: "parsing JSON",
		},
		{
			name:        "empty_json",
			config:      "{}",
			wantErr:     true,
			errContains: "provider.repo is required",
		},
	}

	parser := &JSONParser{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parser.Parse(ctx, []byte(tt.config))
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

// 🧪 TestJSONParserSelection tests JSON parser file detection
func TestJSONParserSelection(t *testing.T) {
	parser := &JSONParser{}

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "json_extension",
			filename: "config.json",
			want:     true,
		},
		{
			name:     "uppercase_extension",
			filename: "config.JSON",
			want:     true,
		},
		{
			name:     "yaml_extension",
			filename: "config.yaml",
			want:     false,
		},
		{
			name:     "no_extension",
			filename: "config",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.CanParse(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}
