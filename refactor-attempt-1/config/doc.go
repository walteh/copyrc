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

/*
Package config provides configuration management for the copyrc tool.

🎯 Purpose:
The config package is responsible for loading, parsing, and validating configuration
files in multiple formats (YAML, JSON, HCL). It provides a unified configuration structure
that other packages can use to control various operations.

🔄 Configuration Flow:

	┌─────────────┐
	│   Config    │
	│    File     │
	└─────────────┘
	       │
	       ▼

┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│    YAML     │   │    JSON     │   │    HCL      │
│   Parser    │◄──│   Parser    │◄──│   Parser    │
└─────────────┘   │  Selection  │   └─────────────┘

	└─────────────┘
	       │
	       ▼
	┌─────────────┐
	│  Validated  │
	│   Config    │
	└─────────────┘
	       │
	       ▼
	┌─────────────┐
	│ Operations  │
	│  Execution  │
	└─────────────┘

🔑 Key Components:

1. Configuration Structure:

  - Provider: Repository source configuration
    ├── Repo: Repository URL
    ├── Ref: Branch/tag reference
    └── Path: Source path within repo

  - Destination: Target path for copied files

  - Copy Options:
    ├── Replacements: Text substitutions
    └── IgnorePatterns: Files to exclude

    2. Operations (Planned):
    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
    │    Copy     │ │    Clean    │ │   Status    │
    └─────────────┘ └─────────────┘ └─────────────┘
    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
    │   Remote    │ │    Async    │ │    Force    │
    │   Status    │ │             │ │             │
    └─────────────┘ └─────────────┘ └─────────────┘

🔌 Interfaces:

Parser Interface:

	type Parser interface {
	    Parse(ctx context.Context, data []byte) (*Config, error)
	    CanParse(filename string) bool
	}

🎨 Design Principles:
1. Format Agnostic: Support multiple config formats
2. Extensible: Easy to add new parsers
3. Validated: Strong validation at parse time
4. Normalized: Consistent path handling
5. Operation-Driven: Config drives operations

📝 Example Usage:

	cfg, err := config.Load(ctx, "config.yaml")
	if err != nil {
	    return err
	}

	// Access configuration
	fmt.Printf("Copying from %s to %s\n",
	    cfg.Provider.Repo,
	    cfg.Destination)

🔍 Configuration Formats:

 1. YAML (Default):
    ```yaml
    provider:
    repo: github.com/org/repo
    ref: main
    path: src/pkg
    destination: /local/path
    copy:
    replacements:
    - old: "foo"
    new: "bar"
    ignore_patterns:
    - "*.tmp"
    ```

 2. JSON:
    ```json
    {
    "provider": {
    "repo": "github.com/org/repo",
    "ref": "main",
    "path": "src/pkg"
    },
    "destination": "/local/path"
    }
    ```

 3. HCL:
    ```hcl
    copy {
    source {
    repo = "github.com/org/repo"
    ref  = "main"
    path = "src/pkg"
    }
    destination {
    path = "/local/path"
    }
    }
    ```

🎯 Operation Flags:
- go_embed: Generate Go embed directives
- clean: Clean destination before copy
- status: Show operation status
- remote_status: Check remote source status
- force: Force operation execution
- async: Run operations asynchronously

🔜 Planned Enhancements:
1. [ ] Operation Registry System
2. [ ] Progress Reporting
3. [ ] Dependency Resolution
4. [ ] Templating Support
5. [ ] Remote Config Support

🔍 Current Issues & TODOs:
- [ ] Improve error messages with line numbers
- [ ] Add schema validation for YAML/JSON
- [ ] Support environment variable interpolation
- [ ] Add config merge functionality
- [ ] Add config versioning

🤔 Deeper Reflection:
1. Parser System:
  - Current registration system is global
  - Consider using dependency injection
  - Add parser priority system

2. Validation:
  - Add custom validation hooks
  - Support conditional validation
  - Add path existence checks

3. Testing:
  - Add fuzz testing for parsers
  - Test edge cases in path normalization
  - Add benchmarks for large configs

4. Security:
  - Add sensitive field masking
  - Validate file permissions
  - Add config source verification

ASCII Diagram:
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Config File │ -> │    Parser    │ -> │  Validated   │
│(.yml/json/hcl)│    │  Selection   │    │   Config     │
└──────────────┘    └──────────────┘    └──────────────┘

	       │
	┌──────┴───────┐
	│   Parsers    │
	├──────────────┤
	│ YAML Parser  │
	│ JSON Parser  │
	│ HCL Parser   │
	└──────────────┘
*/
package config
