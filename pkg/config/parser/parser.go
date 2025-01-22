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

	"github.com/walteh/copyrc/pkg/config"
)

// 🔌 Parser is the interface for config parsers
type Parser interface {
	// 📝 Parse parses the config from bytes
	Parse(ctx context.Context, data []byte) (*config.Config, error)

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
