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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
	"gitlab.com/tozd/go/errors"
	"gopkg.in/yaml.v3"
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
	Replacements   []Replacement `json:"replacements" yaml:"replacements"`       // String replacements to apply
	IgnorePatterns []string      `json:"ignore_patterns" yaml:"ignore_patterns"` // Glob patterns for files to ignore
}

// 📚 Config represents the complete configuration
type Config struct {
	Provider     ProviderArgs `json:"provider" yaml:"provider"`
	Destination  string       `json:"destination" yaml:"destination"`
	Copy         *CopyArgs    `json:"copy,omitempty" yaml:"copy,omitempty"`
	GoEmbed      bool         `json:"go_embed,omitempty" yaml:"go_embed,omitempty"`
	Clean        bool         `json:"clean,omitempty" yaml:"clean,omitempty"`
	Status       bool         `json:"status,omitempty" yaml:"status,omitempty"`
	RemoteStatus bool         `json:"remote_status,omitempty" yaml:"remote_status,omitempty"`
	Force        bool         `json:"force,omitempty" yaml:"force,omitempty"`
	Async        bool         `json:"async,omitempty" yaml:"async,omitempty"`
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
	cfg.Provider.Path = filepath.ToSlash(filepath.Clean(cfg.Provider.Path))
	cfg.Destination = filepath.ToSlash(filepath.Clean(cfg.Destination))

	// Set defaults
	if cfg.Provider.Ref == "" {
		cfg.Provider.Ref = "main"
	}

	return nil
}

// 📝 String returns a string representation of the config
func (cfg *Config) String() string {
	ref := cfg.Provider.Ref
	if ref == "" {
		ref = "main"
	}
	return fmt.Sprintf("%s@%s:%s -> %s", cfg.Provider.Repo, ref, cfg.Provider.Path, cfg.Destination)
}

// 🔧 YAMLParser implements the Parser interface for YAML files
type YAMLParser struct{}

func init() {
	Register(&YAMLParser{})
}

func (p *YAMLParser) CanParse(filename string) bool {
	return strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
}

func (p *YAMLParser) Parse(ctx context.Context, data []byte) (*Config, error) {
	var cfg Config
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, errors.Errorf("parsing YAML: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// 🔧 HCLParser implements the Parser interface for HCL files
type HCLParser struct{}

// 🔄 HCLConfig represents the HCL configuration structure
type HCLConfig struct {
	Copy []struct {
		Source struct {
			Repo string `hcl:"repo" cty:"repo"`
			Ref  string `hcl:"ref,optional" cty:"ref"`
			Path string `hcl:"path" cty:"path"`
		} `hcl:"source,block" cty:"source"`
		Destination struct {
			Path string `hcl:"path" cty:"path"`
		} `hcl:"destination,block" cty:"destination"`
		Options struct {
			Replacements cty.Value `hcl:"replacements,optional" cty:"replacements"`
			IgnoreFiles  []string  `hcl:"ignore_files,optional" cty:"ignore_files"`
		} `hcl:"options,block" cty:"options"`
	} `hcl:"copy,block" cty:"copy"`
}

func init() {
	Register(&HCLParser{})
}

func (p *HCLParser) CanParse(filename string) bool {
	return strings.HasSuffix(filename, ".hcl")
}

func (p *HCLParser) Parse(ctx context.Context, data []byte) (*Config, error) {
	parser := hclparse.NewParser()
	hclFile, diags := parser.ParseHCL(data, "config.hcl")
	if diags.HasErrors() {
		return nil, errors.Errorf("parsing HCL: %s", diags.Error())
	}

	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	var hclCfg HCLConfig
	diags = gohcl.DecodeBody(hclFile.Body, evalCtx, &hclCfg)
	if diags.HasErrors() {
		return nil, errors.Errorf("decoding HCL: %s", diags.Error())
	}

	if len(hclCfg.Copy) == 0 {
		return nil, errors.Errorf("no copy block found")
	}

	// Convert to standard config
	cfg := &Config{
		Provider: ProviderArgs{
			Repo: hclCfg.Copy[0].Source.Repo,
			Ref:  hclCfg.Copy[0].Source.Ref,
			Path: hclCfg.Copy[0].Source.Path,
		},
		Destination: hclCfg.Copy[0].Destination.Path,
	}

	// Convert replacements and ignore patterns
	if !hclCfg.Copy[0].Options.Replacements.IsNull() || len(hclCfg.Copy[0].Options.IgnoreFiles) > 0 {
		cfg.Copy = &CopyArgs{
			IgnorePatterns: hclCfg.Copy[0].Options.IgnoreFiles,
		}

		if !hclCfg.Copy[0].Options.Replacements.IsNull() {
			replacements := hclCfg.Copy[0].Options.Replacements.AsValueSlice()
			cfg.Copy.Replacements = make([]Replacement, len(replacements))

			for i, r := range replacements {
				m := r.AsValueMap()
				old := m["old"].AsString()
				new := m["new"].AsString()
				var file *string
				if f, ok := m["file"]; ok && !f.IsNull() {
					s := f.AsString()
					file = &s
				}
				cfg.Copy.Replacements[i] = Replacement{
					Old:  old,
					New:  new,
					File: file,
				}
			}
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// 📝 LoadFile loads and parses a config file
func LoadFile(path string) (*Config, error) {
	// Read file
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
	if err := validateConfig(&cfg); err != nil {
		return nil, errors.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// 🔍 validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	// Check required fields
	if cfg.Provider.Repo == "" {
		return errors.New("provider.repo is required")
	}
	if cfg.Provider.Ref == "" {
		return errors.New("provider.ref is required")
	}
	if cfg.Destination == "" {
		return errors.New("destination is required")
	}

	// Clean up destination path
	cfg.Destination = filepath.Clean(cfg.Destination)

	return nil
}
