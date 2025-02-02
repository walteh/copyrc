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
	// 🔧 Default settings block
	Defaults *DefaultsBlock `json:"defaults,omitempty" hcl:"defaults,block" yaml:"defaults,omitempty"`

	// 📝 Copy configurations
	Copies []*CopyEntry `json:"copies" hcl:"copy,block" yaml:"copies"`
	// 📝 Archive configurations
	Archives []*ArchiveEntry `json:"archives" hcl:"archive,block" yaml:"archives"`
}

// 🔧 Default settings that apply to all copies
type DefaultsBlock struct {
	CopyOptions    *CopyEntry_Options    `json:"copy_options,omitempty" yaml:"copy_options,omitempty" hcl:"copy_options,block"`
	ArchiveOptions *ArchiveEntry_Options `json:"archive_options,omitempty" yaml:"archive_options,omitempty" hcl:"archive_options,block"`
}

// 🎯 Source configuration
type Source struct {
	Repo      string `json:"repo" yaml:"repo" hcl:"repo,attr"`
	Ref       string `json:"ref,omitempty" yaml:"ref,omitempty" hcl:"ref,attr"`
	Path      string `json:"path" yaml:"path" hcl:"path,optional"`
	RefType   string `json:"ref_type" yaml:"ref_type" hcl:"ref_type,optional"`
	Recursive bool   `json:"recursive,omitempty" yaml:"recursive,omitempty" hcl:"recursive,optional"`
}

// 📦 Destination configuration
type Destination struct {
	Path string `json:"path" yaml:"path" hcl:"path,attr"`
}

// 🔧 Processing options (internal)
type CopyEntry_Options struct {
	Replacements []Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty" hcl:"replacements,optional" cty:"replacements"`
	IgnoreFiles  []string      `json:"ignore_files,omitempty" yaml:"ignore_files,omitempty" hcl:"ignore_files,optional" cty:"ignore_files"`
	FilePatterns []string      `json:"file_patterns,omitempty" yaml:"file_patterns,omitempty" hcl:"file_patterns,optional" cty:"file_patterns"`
	Recursive    bool          `json:"recursive,omitempty" yaml:"recursive,omitempty" hcl:"recursive,optional" cty:"recursive"` // 📁 Enable recursive directory copying
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
func LoadConfig(path string) (*CopyConfig, error) {
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

	// Convert to internal format
	return &cfg, nil

}

// 🏃 Run all copy operations
func (cfg *CopyConfig) RunAll(ctx context.Context, clean, status, remoteStatus, force bool, provider RepoProvider) error {
	logger := loggerFromContext(ctx)
	logger.Header("Copying files from repositories")

	// Process copies
	for _, copy := range cfg.Copies {
		config := &Config{
			Source: Source{
				Repo:    copy.Source.Repo,
				Ref:     copy.Source.Ref,
				Path:    copy.Source.Path,
				RefType: copy.Source.RefType,
			},
			DestPath:     copy.Destination.Path,
			CopyArgs:     copy.Options,
			Clean:        clean,
			Status:       status,
			RemoteStatus: remoteStatus,
			Force:        force,
		}

		if err := process(ctx, config, provider); err != nil {
			return errors.Errorf("running copy %s: %w", copy.Destination.Path, err)
		}
	}

	// Process archives
	for _, archive := range cfg.Archives {
		config := &Config{
			Source: Source{
				Repo:    archive.Source.Repo,
				Ref:     archive.Source.Ref,
				RefType: archive.Source.RefType,
			},
			DestPath:     archive.Destination.Path,
			ArchiveArgs:  archive.Options,
			Clean:        clean,
			Status:       status,
			RemoteStatus: remoteStatus,
			Force:        force,
		}

		if err := process(ctx, config, provider); err != nil {
			return errors.Errorf("running archive %s: %w", archive.Destination.Path, err)
		}
	}

	return nil
}
