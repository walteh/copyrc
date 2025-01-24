package config_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/walteh/copyrc/pkg/config"
)

func ExampleLoadConfig_json() {
	ctx := context.Background()
	// Create a temporary JSON config file
	configJSON := `{
		"repositories": [
			{
				"provider": "github",
				"name": "org/repo",
				"ref": "main"
			}
		],
		"copies": [
			{
				"repository": {
					"provider": "github",
					"name": "org/repo",
					"ref": "main"
				},
				"paths": {
					"remote": "/src",
					"local": "/dest"
				},
				"options": {
					"text_replacements": [
						{
							"from_text": "foo",
							"to_text": "bar",
							"file_filter_glob": "*.go"
						}
					]
				}
			}
		]
	}`

	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
		return
	}

	// Load and validate the config
	cfg, err := config.LoadConfig(ctx, configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Print some config details
	fmt.Printf("Loaded %d repositories and %d copy operations\n", len(cfg.Repositories), len(cfg.Copies))
	fmt.Printf("First repository: %s/%s@%s\n", cfg.Repositories[0].Provider, cfg.Repositories[0].Name, cfg.Repositories[0].Ref)
	fmt.Printf("First copy operation: %s -> %s\n", cfg.Copies[0].Paths.Remote, cfg.Copies[0].Paths.Local)

	// Output:
	// Loaded 1 repositories and 1 copy operations
	// First repository: github/org/repo@main
	// First copy operation: /src -> /dest
}

func ExampleLoadConfig_yaml() {
	ctx := context.Background()
	// Create a temporary YAML config file
	configYAML := `
repositories:
  - provider: github
    name: org/repo
    ref: main

copies:
  - repository:
      provider: github
      name: org/repo
      ref: main
    paths:
      remote: /src
      local: /dest
    options:
      text_replacements:
        - from_text: foo
          to_text: bar
          file_filter_glob: "*.go"
`

	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
		return
	}

	// Load and validate the config
	cfg, err := config.LoadConfig(ctx, configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Print some config details
	fmt.Printf("Loaded %d repositories and %d copy operations\n", len(cfg.Repositories), len(cfg.Copies))
	fmt.Printf("First repository: %s/%s@%s\n", cfg.Repositories[0].Provider, cfg.Repositories[0].Name, cfg.Repositories[0].Ref)
	fmt.Printf("First copy operation: %s -> %s\n", cfg.Copies[0].Paths.Remote, cfg.Copies[0].Paths.Local)

	// Output:
	// Loaded 1 repositories and 1 copy operations
	// First repository: github/org/repo@main
	// First copy operation: /src -> /dest
}

func ExampleLoadConfig_hcl() {
	ctx := context.Background()
	// Create a temporary HCL config file
	configHCL := `
repositories {
  provider = "github"
  name = "org/repo"
  ref = "main"
}

copy {
  repository {
    provider = "github"
    name = "org/repo"
    ref = "main"
  }
  paths {
    remote = "/src"
    local = "/dest"
  }
  options {
    text_replacements = [
      {
        from_text = "foo"
        to_text = "bar"
        file_filter_glob = "*.go"
      }
    ]
  }
}
`

	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "config.hcl")
	if err := os.WriteFile(configPath, []byte(configHCL), 0644); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
		return
	}

	// Load and validate the config
	cfg, err := config.LoadConfig(ctx, configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Print some config details
	fmt.Printf("Loaded %d repositories and %d copy operations\n", len(cfg.Repositories), len(cfg.Copies))
	fmt.Printf("First repository: %s/%s@%s\n", cfg.Repositories[0].Provider, cfg.Repositories[0].Name, cfg.Repositories[0].Ref)
	fmt.Printf("First copy operation: %s -> %s\n", cfg.Copies[0].Paths.Remote, cfg.Copies[0].Paths.Local)

	// Output:
	// Loaded 1 repositories and 1 copy operations
	// First repository: github/org/repo@main
	// First copy operation: /src -> /dest
}

func ExampleCopyrcConfig_Hash() {
	// Create a config
	cfg := &config.CopyrcConfig{
		Repositories: []config.RepositoryDefinition{
			{
				Provider: "github",
				Name:     "org/repo",
				Ref:      "main",
			},
		},
	}

	// Get initial hash
	hash1 := cfg.Hash()
	fmt.Printf("Initial hash length: %d\n", len(hash1))
	fmt.Printf("Initial hash is hex: %v\n", len(hash1) == 64)

	// Modify config
	cfg.Repositories[0].Ref = "v1.0.0"

	// Get new hash
	hash2 := cfg.Hash()
	fmt.Printf("Modified hash length: %d\n", len(hash2))
	fmt.Printf("Modified hash is hex: %v\n", len(hash2) == 64)
	fmt.Printf("Hashes are different: %v\n", hash1 != hash2)

	// Output:
	// Initial hash length: 64
	// Initial hash is hex: true
	// Modified hash length: 64
	// Modified hash is hex: true
	// Hashes are different: true
}

func ExampleCopyrcConfig_Validate() {
	ctx := context.Background()
	// Create an invalid config
	cfg := &config.CopyrcConfig{
		Repositories: []config.RepositoryDefinition{
			{
				// Missing required fields
			},
		},
	}

	// Try to validate
	err := config.Validate(ctx, cfg)
	fmt.Printf("Validation error: %v\n", err)

	// Fix the config
	cfg.Repositories[0].Provider = "github"
	cfg.Repositories[0].Name = "org/repo"
	cfg.Repositories[0].Ref = "main"

	// Add required copy operation
	cfg.Copies = []config.Copy{
		{
			Repository: cfg.Repositories[0],
			Paths: config.CopyPaths{
				Remote: "/src",
				Local:  "/dest",
			},
		},
	}

	// Validate again
	err = config.Validate(ctx, cfg)
	fmt.Printf("Config is valid: %v\n", err == nil)

	// Output:
	// Validation error: repository 0: provider is required
	// Config is valid: true
}
