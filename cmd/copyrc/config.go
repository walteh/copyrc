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
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"gitlab.com/tozd/go/errors"
	"gopkg.in/yaml.v3"
)

// 📝 Config file structure
type CopyConfig struct {
	// 📝 Copy configurations
	Copies []*CopyEntry `json:"copies" hcl:"copy,block" yaml:"copies"`
	// 📝 Archive configurations
	Archives []*ArchiveEntry `json:"archives" hcl:"archive,block" yaml:"archives"`
	// 🔧 Flags block
	Flags *FlagsBlock `json:"flags,omitempty" hcl:"flags,block" yaml:"flags,omitempty"`
}

type SingleConfig struct {
	Source      Source
	Destination Destination
	CopyArgs    *CopyEntry_Options
	ArchiveArgs *ArchiveEntry_Options
	Flags       FlagsBlock
}

type FlagsBlock struct {
	Clean        bool `json:"clean,omitempty" hcl:"clean,optional" yaml:"clean,omitempty"`
	Status       bool `json:"status,omitempty" hcl:"status,optional" yaml:"status,omitempty"`
	RemoteStatus bool `json:"remote_status,omitempty" hcl:"remote_status,optional" yaml:"remote_status,omitempty"`
	Force        bool `json:"force,omitempty" hcl:"force,optional" yaml:"force,omitempty"`
	Async        bool `json:"async,omitempty" hcl:"async,optional" yaml:"async,omitempty"`
}

// 🎯 Source configuration
type Source struct {
	Repo    string `json:"repo" yaml:"repo" hcl:"repo,attr"`
	Ref     string `json:"ref,omitempty" yaml:"ref,omitempty" hcl:"ref,attr"`
	Path    string `json:"path" yaml:"path" hcl:"path,optional"`
	RefType string `json:"ref_type" yaml:"ref_type" hcl:"ref_type,optional"`
}

// 📦 Destination configuration
type Destination struct {
	Path string `json:"path" yaml:"path" hcl:"path,attr"`
}

// 🔧 Processing options (internal)
type CopyEntry_Options struct {
	Replacements     []Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty" hcl:"replacements,optional" cty:"replacements"`
	IgnoreFiles      []string      `json:"ignore_files,omitempty" yaml:"ignore_files,omitempty" hcl:"ignore_files,optional" cty:"ignore_files"`
	FilePatterns     []string      `json:"file_patterns,omitempty" yaml:"file_patterns,omitempty" hcl:"file_patterns,optional" cty:"file_patterns"`
	Recursive        bool          `json:"recursive,omitempty" yaml:"recursive,omitempty" hcl:"recursive,optional" cty:"recursive"` // 📁 Enable recursive directory copying
	ExtensionPrefix  string        `json:"extension_prefix,omitempty" yaml:"extension_prefix,omitempty" hcl:"extension_prefix,optional" cty:"extension_prefix"`
	NoHeaderComments bool          `json:"skip_header_comments,omitempty" yaml:"skip_header_comments,omitempty" hcl:"skip_header_comments,optional" cty:"skip_header_comments"`
}

// 📝 Individual copy entry
type CopyEntry struct {
	Source      Source             `json:"source" yaml:"source" hcl:"source,block"`
	Destination Destination        `json:"destination" yaml:"destination" hcl:"destination,block"`
	Options     *CopyEntry_Options `json:"options" yaml:"options" hcl:"options,block"`
}

// 📝 Archive entry
type ArchiveEntry struct {
	Source      Source                `yaml:"source" hcl:"source,block"`
	Destination Destination           `yaml:"destination" hcl:"destination,block"`
	Options     *ArchiveEntry_Options `yaml:"options,omitempty" hcl:"options,block"`
}

type ArchiveEntry_Options struct {
	GoEmbed bool `yaml:"go_embed,omitempty" hcl:"go_embed,optional"`
}

// 📝 Load config from file (supports YAML and HCL)
func LoadConfig(path string, input Input) (*CopyConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("reading config file: %w", err)
	}

	// Try YAML first
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		var cfg CopyConfig
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		if err := decoder.Decode(&cfg); err != nil {
			return nil, errors.Errorf("parsing YAML: %w", err)
		}
		return &cfg, nil
	}
	parser := hclparse.NewParser()
	hclFile, diags := parser.ParseHCL(data, path)
	if diags.HasErrors() {
		return nil, errors.Errorf("parsing HCL: %s", diags.Error())
	}

	// Create evaluation context
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Decode HCL into our HCL-specific schema
	var cfg CopyConfig
	diags = gohcl.DecodeBody(hclFile.Body, ctx, &cfg)
	if diags.HasErrors() {
		return nil, errors.Errorf("decoding HCL: %s", diags.Error())
	}

	if cfg.Flags == nil {
		cfg.Flags = &FlagsBlock{}
	}

	if input.Status.IsSet() {
		cfg.Flags.Status = input.Status.value
	}
	if input.RemoteStatus.IsSet() {
		cfg.Flags.RemoteStatus = input.RemoteStatus.value
	}

	if input.Force.IsSet() {
		cfg.Flags.Force = input.Force.value
	}

	if input.Async.IsSet() {
		cfg.Flags.Async = input.Async.value
	}

	if input.Clean.IsSet() {
		cfg.Flags.Clean = input.Clean.value
	}

	// remove all ./ from dest and source
	for _, copy := range cfg.Copies {
		copy.Destination.Path = strings.TrimPrefix(copy.Destination.Path, "./")
		copy.Source.Path = strings.TrimPrefix(copy.Source.Path, "./")
	}

	for _, archive := range cfg.Archives {
		archive.Destination.Path = strings.TrimPrefix(archive.Destination.Path, "./")
		archive.Source.Path = strings.TrimPrefix(archive.Source.Path, "./")
	}

	// Convert to internal format
	return &cfg, nil

}

// 🏃 Run all copy operations
func (cfg *CopyConfig) RunAll(ctx context.Context, provider RepoProvider) error {
	logger := loggerFromContext(ctx)
	logger.Header("Copying files from repositories")

	// Process copies
	for _, copy := range cfg.Copies {
		config := &SingleConfig{
			Source:      copy.Source,
			Destination: copy.Destination,
			CopyArgs:    copy.Options,
			ArchiveArgs: nil,
		}
		if cfg.Flags != nil {
			config.Flags = *cfg.Flags
		}

		if err := process(ctx, config, provider); err != nil {
			return errors.Errorf("running copy %s: %w", copy.Destination.Path, err)
		}
	}

	// Process archives
	for _, archive := range cfg.Archives {
		config := &SingleConfig{
			Source:      archive.Source,
			Destination: archive.Destination,
			ArchiveArgs: archive.Options,
			CopyArgs:    nil,
		}
		if cfg.Flags != nil {
			config.Flags = *cfg.Flags
		}

		if err := process(ctx, config, provider); err != nil {
			return errors.Errorf("running archive %s: %w", archive.Destination.Path, err)
		}
	}

	return nil
}
