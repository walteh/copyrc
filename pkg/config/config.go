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
)

// 🔌 Parser is the interface for config parsers
type Parser interface {
	// 📝 Parse parses the config from bytes
	Parse(ctx context.Context, data []byte) (*Config, error)

	// 🔍 CanParse checks if this parser can handle the given file
	CanParse(filename string) bool
}

var (
	// 🗺️ parsers is a list of available parsers
	parsers []Parser
)

// 📝 Register registers a parser
func Register(p Parser) {
	parsers = append(parsers, p)
}

// 🎯 GetParser returns a parser that can handle the given file
func GetParser(filename string) Parser {
	for _, p := range parsers {
		if p.CanParse(filename) {
			return p
		}
	}
	return nil
}

// 🔄 Replacement represents a string replacement in files
type Replacement struct {
	Old  string  // Original string to replace
	New  string  // New string to use
	File *string // Optional specific file to apply to
}

// 📦 ProviderArgs represents arguments for a repository provider
type ProviderArgs struct {
	Repo string // Full repo URL (e.g. github.com/org/repo)
	Ref  string // Branch or tag
	Path string // Path within repo
}

// 🔧 CopyArgs represents file copy configuration
type CopyArgs struct {
	Replacements []Replacement // String replacements to apply
	IgnoreFiles  []string      // Files to ignore
}

// 📚 Config represents the complete configuration
type Config struct {
	Provider     ProviderArgs // Repository provider configuration
	Destination  string       // Local destination path
	Copy         *CopyArgs    // Copy configuration
	GoEmbed      bool         // Whether to generate Go embed code
	Clean        bool         // Whether to clean destination directory
	Status       bool         // Whether to check local status
	RemoteStatus bool         // Whether to check remote status
	Force        bool         // Whether to force update even if status is ok
	Async        bool         // Whether to process files asynchronously
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

	// Get parser
	p := GetParser(path)
	if p == nil {
		return nil, errors.Errorf("no parser found for file: %s", path)
	}

	// Parse config
	cfg, err := p.Parse(ctx, data)
	if err != nil {
		return nil, errors.Errorf("parsing config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// 🔍 Validate checks if the configuration is valid
func (cfg *Config) Validate() error {
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
