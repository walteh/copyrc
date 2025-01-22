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

package parser

import (
	"context"
	"strings"

	"github.com/walteh/copyrc/pkg/config/model"
	"gitlab.com/tozd/go/errors"
	"gopkg.in/yaml.v3"
)

func init() {
	Register(&YAMLParser{})
}

// 🔧 YAMLParser implements the Parser interface for YAML files
type YAMLParser struct{}

// 🔍 CanParse checks if this parser can handle the given file
func (p *YAMLParser) CanParse(filename string) bool {
	return strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
}

// 📝 Parse parses the config from YAML
func (p *YAMLParser) Parse(ctx context.Context, data []byte) (*model.Config, error) {
	// Define YAML schema
	type yamlConfig struct {
		Provider struct {
			Repo string `yaml:"repo"`
			Ref  string `yaml:"ref,omitempty"`
			Path string `yaml:"path"`
		} `yaml:"provider"`
		Destination string `yaml:"destination"`
		Copy        *struct {
			Replacements []struct {
				Old  string  `yaml:"old"`
				New  string  `yaml:"new"`
				File *string `yaml:"file,omitempty"`
			} `yaml:"replacements,omitempty"`
			IgnoreFiles []string `yaml:"ignore_files,omitempty"`
		} `yaml:"copy,omitempty"`
		GoEmbed      bool `yaml:"go_embed,omitempty"`
		Clean        bool `yaml:"clean,omitempty"`
		Status       bool `yaml:"status,omitempty"`
		RemoteStatus bool `yaml:"remote_status,omitempty"`
		Force        bool `yaml:"force,omitempty"`
		Async        bool `yaml:"async,omitempty"`
	}

	// Parse YAML
	var yamlCfg yamlConfig
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&yamlCfg); err != nil {
		return nil, errors.Errorf("parsing YAML: %w", err)
	}

	// Convert to model
	cfg := &model.Config{
		Provider: model.ProviderArgs{
			Repo: yamlCfg.Provider.Repo,
			Ref:  yamlCfg.Provider.Ref,
			Path: yamlCfg.Provider.Path,
		},
		Destination:  yamlCfg.Destination,
		GoEmbed:      yamlCfg.GoEmbed,
		Clean:        yamlCfg.Clean,
		Status:       yamlCfg.Status,
		RemoteStatus: yamlCfg.RemoteStatus,
		Force:        yamlCfg.Force,
		Async:        yamlCfg.Async,
	}

	if yamlCfg.Copy != nil {
		cfg.Copy = &model.CopyArgs{
			IgnoreFiles: yamlCfg.Copy.IgnoreFiles,
		}
		for _, r := range yamlCfg.Copy.Replacements {
			cfg.Copy.Replacements = append(cfg.Copy.Replacements, model.Replacement{
				Old:  r.Old,
				New:  r.New,
				File: r.File,
			})
		}
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("validating config: %w", err)
	}

	return cfg, nil
}
