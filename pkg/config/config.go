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

	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"
	"gopkg.in/yaml.v3"
)

// 🔄 Replacement represents a string replacement in files
type Replacement struct {
	Old  string  `yaml:"old"`  // Original string to replace
	New  string  `yaml:"new"`  // New string to use
	File *string `yaml:"file"` // Optional specific file to apply to
}

// 📦 ProviderArgs represents arguments for a repository provider
type ProviderArgs struct {
	Repo string `yaml:"repo"` // Full repo URL (e.g. github.com/org/repo)
	Ref  string `yaml:"ref"`  // Branch or tag
	Path string `yaml:"path"` // Path within repo
}

// 🔧 CopyArgs represents file copy configuration
type CopyArgs struct {
	Replacements []Replacement `yaml:"replacements"` // String replacements to apply
	IgnoreFiles  []string      `yaml:"ignore_files"` // Files to ignore
}

// 📚 Config represents the complete configuration
type Config struct {
	Provider     ProviderArgs `yaml:"provider"`      // Repository provider configuration
	Destination  string       `yaml:"destination"`   // Local destination path
	Copy         *CopyArgs    `yaml:"copy"`          // Copy configuration
	GoEmbed      bool         `yaml:"go_embed"`      // Whether to generate Go embed code
	Clean        bool         `yaml:"clean"`         // Whether to clean destination directory
	Status       bool         `yaml:"status"`        // Whether to check local status
	RemoteStatus bool         `yaml:"remote_status"` // Whether to check remote status
	Force        bool         `yaml:"force"`         // Whether to force update even if status is ok
	Async        bool         `yaml:"async"`         // Whether to process files asynchronously
}

// 🎯 Load loads the configuration from a file
func Load(ctx context.Context, path string) (*Config, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", path).Msg("loading configuration")

	// Read config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("reading config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Errorf("parsing config file: %w", err)
	}

	// Validate config
	if err := cfg.validate(); err != nil {
		return nil, errors.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// 🔍 validate checks if the configuration is valid
func (cfg *Config) validate() error {
	// Check required fields
	if cfg.Provider.Repo == "" {
		return errors.Errorf("provider.repo is required")
	}
	if cfg.Provider.Path == "" {
		return errors.Errorf("provider.path is required")
	}
	if cfg.Destination == "" {
		return errors.Errorf("destination is required")
	}

	// Clean up paths
	cfg.Provider.Path = filepath.Clean(cfg.Provider.Path)
	cfg.Destination = filepath.Clean(cfg.Destination)

	// Set defaults
	if cfg.Provider.Ref == "" {
		cfg.Provider.Ref = "main"
	}

	return nil
}

// 📝 String returns a string representation of the config
func (cfg *Config) String() string {
	return cfg.Provider.Repo + "@" + cfg.Provider.Ref + ":" + cfg.Provider.Path + " -> " + cfg.Destination
}
