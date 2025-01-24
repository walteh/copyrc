package config

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"gitlab.com/tozd/go/errors"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads a configuration file from the given path.
// The format is determined by the file extension:
// - .json for JSON
// - .yaml or .yml for YAML
// - .hcl for HCL
// - .copyrc will try both YAML and HCL formats
func LoadConfig(ctx context.Context, path string) (*CopyrcConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("reading config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	var cfg *CopyrcConfig

	// For .copyrc files, try both YAML and HCL
	if ext == ".copyrc" || path == ".copyrc" {
		// Try YAML first
		cfg, err = loadYAML(data)
		if err == nil {
			return cfg, nil
		}

		// Try HCL next
		cfg, err = loadHCL(data, path)
		if err == nil {
			return cfg, nil
		}

		return nil, errors.Errorf("failed to parse .copyrc as YAML or HCL: %w", err)
	}

	// Otherwise use extension to determine format
	switch ext {
	case ".json":
		cfg, err = loadJSON(data)
	case ".yaml", ".yml":
		cfg, err = loadYAML(data)
	case ".hcl":
		cfg, err = loadHCL(data, path)
	default:
		return nil, errors.Errorf("unsupported file extension %q", ext)
	}

	if err != nil {
		return nil, err
	}
	cfg.location = path
	if err := Validate(ctx, cfg); err != nil {
		return nil, errors.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// loadJSON loads a configuration from JSON data
func loadJSON(data []byte) (*CopyrcConfig, error) {
	var cfg CopyrcConfig
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, errors.Errorf("parsing JSON: %w", err)
	}
	return &cfg, nil
}

// loadYAML loads a configuration from YAML data
func loadYAML(data []byte) (*CopyrcConfig, error) {
	var cfg CopyrcConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, errors.Errorf("parsing YAML: %w", err)
	}
	return &cfg, nil
}

// loadHCL loads a configuration from HCL data
func loadHCL(data []byte, filename string) (*CopyrcConfig, error) {
	parser := hclparse.NewParser()
	hclFile, diags := parser.ParseHCL(data, filename)
	if diags.HasErrors() {
		return nil, errors.Errorf("parsing HCL: %s", diags.Error())
	}

	// Create evaluation context
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Decode HCL into our config struct
	var cfg CopyrcConfig
	diags = gohcl.DecodeBody(hclFile.Body, ctx, &cfg)
	if diags.HasErrors() {
		return nil, errors.Errorf("decoding HCL: %s", diags.Error())
	}

	return &cfg, nil
}

// TODO(dr.methodical): 🧪 Add tests for each format's specific features
// TODO(dr.methodical): 🧪 Add tests for file extension validation
// TODO(dr.methodical): 🧪 Add tests for unknown field handling
// TODO(dr.methodical): 📝 Add examples of loading different formats
