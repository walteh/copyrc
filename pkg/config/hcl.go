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
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"gitlab.com/tozd/go/errors"
)

func init() {
	Register(&HCLParser{})
}

// 🔧 HCLParser implements the Parser interface for HCL files
type HCLParser struct{}

// 🔍 CanParse checks if this parser can handle the given file
func (p *HCLParser) CanParse(filename string) bool {
	return strings.HasSuffix(filename, ".hcl")
}

// 📝 Parse parses the config from HCL
func (p *HCLParser) Parse(ctx context.Context, data []byte) (*Config, error) {
	parser := hclparse.NewParser()
	hclFile, diags := parser.ParseHCL(data, "config.hcl")
	if diags.HasErrors() {
		return nil, errors.Errorf("parsing HCL: %s", diags.Error())
	}

	// Create evaluation context
	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Define HCL schema
	type hclConfig struct {
		Provider struct {
			Repo string `hcl:"repo"`
			Ref  string `hcl:"ref,optional"`
			Path string `hcl:"path"`
		} `hcl:"provider,block"`
		Destination string `hcl:"destination"`
		Copy        *struct {
			Replacements []struct {
				Old  string  `hcl:"old"`
				New  string  `hcl:"new"`
				File *string `hcl:"file,optional"`
			} `hcl:"replacements,optional"`
			IgnoreFiles []string `hcl:"ignore_files,optional"`
		} `hcl:"copy,block,optional"`
		GoEmbed      bool `hcl:"go_embed,optional"`
		Clean        bool `hcl:"clean,optional"`
		Status       bool `hcl:"status,optional"`
		RemoteStatus bool `hcl:"remote_status,optional"`
		Force        bool `hcl:"force,optional"`
		Async        bool `hcl:"async,optional"`
	}

	// Decode HCL
	var hclCfg hclConfig
	diags = gohcl.DecodeBody(hclFile.Body, evalCtx, &hclCfg)
	if diags.HasErrors() {
		return nil, errors.Errorf("decoding HCL: %s", diags.Error())
	}

	// Convert to model
	cfg := &Config{
		Provider: ProviderArgs{
			Repo: hclCfg.Provider.Repo,
			Ref:  hclCfg.Provider.Ref,
			Path: hclCfg.Provider.Path,
		},
		Destination:  hclCfg.Destination,
		GoEmbed:      hclCfg.GoEmbed,
		Clean:        hclCfg.Clean,
		Status:       hclCfg.Status,
		RemoteStatus: hclCfg.RemoteStatus,
		Force:        hclCfg.Force,
		Async:        hclCfg.Async,
	}

	if hclCfg.Copy != nil {
		cfg.Copy = &CopyArgs{
			IgnoreFiles: hclCfg.Copy.IgnoreFiles,
		}
		for _, r := range hclCfg.Copy.Replacements {
			cfg.Copy.Replacements = append(cfg.Copy.Replacements, Replacement{
				Old:  r.Old,
				New:  r.New,
				File: r.File,
			})
		}
	}

	return cfg, nil
}
