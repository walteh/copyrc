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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 🧪 TestParserRegistration tests the parser registration system
func TestParserRegistration(t *testing.T) {
	// Save original parsers
	originalParsers := parsers
	defer func() {
		parsers = originalParsers
	}()

	// Reset parsers
	parsers = nil

	// Create mock parser
	mockParser := &struct {
		Parser
		canParse bool
	}{
		canParse: true,
	}

	// Test registration
	Register(mockParser)
	assert.Len(t, parsers, 1, "should have 1 parser registered")
	assert.Equal(t, mockParser, parsers[0], "registered parser should match")
}

// 🧪 TestParserSelection tests parser selection by file extension
func TestParserSelection(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     Parser
	}{
		{
			name:     "yaml_file",
			filename: "config.yaml",
			want:     &YAMLParser{},
		},
		{
			name:     "yml_file",
			filename: "config.yml",
			want:     &YAMLParser{},
		},
		{
			name:     "hcl_file",
			filename: "config.hcl",
			want:     &HCLParser{},
		},
		{
			name:     "unknown_extension",
			filename: "config.txt",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetParser(tt.filename)
			if tt.want == nil {
				assert.Nil(t, got, "should return nil for unknown extension")
				return
			}
			require.NotNil(t, got, "should return a parser")
			assert.IsType(t, tt.want, got, "should return correct parser type")
		})
	}
}

// 🧪 TestHCLParsing tests HCL config parsing
func TestHCLParsing(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		wantErr     bool
		errContains string
		check       func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid_hcl",
			config: `
copy {
  source {
    repo = "github.com/walteh/copyrc"
    ref  = "main"
    path = "pkg/provider"
  }
  destination {
    path = "/tmp/copyrc"
  }
  options {
    replacements = [
      {
        old = "foo"
        new = "bar"
      }
    ]
    ignore_files = [
      "*.tmp",
      "*.log"
    ]
  }
}
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "github.com/walteh/copyrc", cfg.Provider.Repo)
				assert.Equal(t, "main", cfg.Provider.Ref)
				assert.Equal(t, "pkg/provider", cfg.Provider.Path)
				assert.Equal(t, "/tmp/copyrc", cfg.Destination)
				assert.NotNil(t, cfg.Copy)
				assert.Len(t, cfg.Copy.Replacements, 1)
				assert.Equal(t, "foo", cfg.Copy.Replacements[0].Old)
				assert.Equal(t, "bar", cfg.Copy.Replacements[0].New)
				assert.Nil(t, cfg.Copy.Replacements[0].File)
				assert.Equal(t, []string{"*.tmp", "*.log"}, cfg.Copy.IgnorePatterns)
			},
		},
		{
			name: "invalid_hcl_syntax",
			config: `
copy {
  source {
    repo = "github.com/walteh/copyrc"
    ref = 
  }
}`,
			wantErr:     true,
			errContains: "parsing HCL",
		},
		{
			name: "invalid_block_type",
			config: `
unknown_block {
  foo = "bar"
}`,
			wantErr:     true,
			errContains: "decoding HCL",
		},
	}

	parser := &HCLParser{}
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

// 🧪 TestPathNormalization tests path normalization in config
func TestPathNormalization(t *testing.T) {
	tests := []struct {
		name        string
		provPath    string
		destPath    string
		wantProv    string
		wantDest    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "clean_paths",
			provPath: "pkg/provider",
			destPath: "/tmp/copyrc",
			wantProv: "pkg/provider",
			wantDest: "/tmp/copyrc",
		},
		{
			name:     "normalize_slashes",
			provPath: filepath.FromSlash("pkg/provider"),
			destPath: "/tmp//copyrc/",
			wantProv: "pkg/provider",
			wantDest: "/tmp/copyrc",
		},
		{
			name:     "remove_dots",
			provPath: "./pkg/provider/../provider",
			destPath: "/tmp/./copyrc",
			wantProv: "pkg/provider",
			wantDest: "/tmp/copyrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Provider: ProviderArgs{
					Repo: "github.com/walteh/copyrc",
					Path: tt.provPath,
				},
				Destination: tt.destPath,
			}

			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantProv, cfg.Provider.Path, "provider path should be normalized")
			assert.Equal(t, tt.wantDest, cfg.Destination, "destination path should be normalized")
		})
	}
}
