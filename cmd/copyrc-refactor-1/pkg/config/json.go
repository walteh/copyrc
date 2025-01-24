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
	"encoding/json"
	"strings"

	"gitlab.com/tozd/go/errors"
)

// 🔧 JSONParser implements the Parser interface for JSON files
type JSONParser struct{}

func init() {
	Register(&JSONParser{})
}

// 🔍 CanParse checks if this parser can handle the given file
func (p *JSONParser) CanParse(filename string) bool {
	return strings.EqualFold(strings.ToLower(strings.TrimSpace(filename)), "config.json") || strings.HasSuffix(strings.ToLower(strings.TrimSpace(filename)), ".json")
}

// 📝 Parse parses the config from JSON bytes
func (p *JSONParser) Parse(ctx context.Context, data []byte) (*Config, error) {
	var cfg Config
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, errors.Errorf("parsing JSON: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}
