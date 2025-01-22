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

package model

import (
	"path/filepath"

	"gitlab.com/tozd/go/errors"
)

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
