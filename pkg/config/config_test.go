package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		format      string // "json", "yaml", "hcl", or "copyrc"
		config      string
		expectError bool
		validate    func(t *testing.T, cfg *CopyrcConfig)
	}{
		{
			name:   "valid_json_config",
			format: "json",
			config: `{
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
							],
							"save_archive_to_path": "/tmp/archive",
							"create_go_embed_for_archive": true
						}
					}
				]
			}`,
			validate: func(t *testing.T, cfg *CopyrcConfig) {
				require.NotNil(t, cfg)
				require.Len(t, cfg.Repositories, 1)
				require.Equal(t, "github", cfg.Repositories[0].Provider)
				require.Equal(t, "org/repo", cfg.Repositories[0].Name)
				require.Equal(t, "main", cfg.Repositories[0].Ref)

				require.Len(t, cfg.Copies, 1)
				require.Equal(t, "github", cfg.Copies[0].Repository.Provider)
				require.Equal(t, "org/repo", cfg.Copies[0].Repository.Name)
				require.Equal(t, "main", cfg.Copies[0].Repository.Ref)
				require.Equal(t, "/src", cfg.Copies[0].Paths.Remote)
				require.Equal(t, "/dest", cfg.Copies[0].Paths.Local)

				require.Len(t, cfg.Copies[0].Options.TextReplacements, 1)
				require.Equal(t, "foo", cfg.Copies[0].Options.TextReplacements[0].FromText)
				require.Equal(t, "bar", cfg.Copies[0].Options.TextReplacements[0].ToText)
				require.Equal(t, "*.go", cfg.Copies[0].Options.TextReplacements[0].FileFilterGlob)
				require.Equal(t, "/tmp/archive", cfg.Copies[0].Options.SaveArchiveToPath)
				require.True(t, cfg.Copies[0].Options.CreateGoEmbedForArchive)
			},
		},
		{
			name:   "valid_yaml_config",
			format: "yaml",
			config: `
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
      save_archive_to_path: /tmp/archive
      create_go_embed_for_archive: true
`,
			validate: func(t *testing.T, cfg *CopyrcConfig) {
				require.NotNil(t, cfg)
				require.Len(t, cfg.Repositories, 1)
				require.Equal(t, "github", cfg.Repositories[0].Provider)
				require.Equal(t, "org/repo", cfg.Repositories[0].Name)
				require.Equal(t, "main", cfg.Repositories[0].Ref)

				require.Len(t, cfg.Copies, 1)
				require.Equal(t, "github", cfg.Copies[0].Repository.Provider)
				require.Equal(t, "org/repo", cfg.Copies[0].Repository.Name)
				require.Equal(t, "main", cfg.Copies[0].Repository.Ref)
				require.Equal(t, "/src", cfg.Copies[0].Paths.Remote)
				require.Equal(t, "/dest", cfg.Copies[0].Paths.Local)

				require.Len(t, cfg.Copies[0].Options.TextReplacements, 1)
				require.Equal(t, "foo", cfg.Copies[0].Options.TextReplacements[0].FromText)
				require.Equal(t, "bar", cfg.Copies[0].Options.TextReplacements[0].ToText)
				require.Equal(t, "*.go", cfg.Copies[0].Options.TextReplacements[0].FileFilterGlob)
				require.Equal(t, "/tmp/archive", cfg.Copies[0].Options.SaveArchiveToPath)
				require.True(t, cfg.Copies[0].Options.CreateGoEmbedForArchive)
			},
		},
		{
			name:   "valid_hcl_config",
			format: "hcl",
			config: `
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
    save_archive_to_path = "/tmp/archive"
    create_go_embed_for_archive = true
  }
}
`,
			validate: func(t *testing.T, cfg *CopyrcConfig) {
				require.NotNil(t, cfg)
				require.Len(t, cfg.Repositories, 1)
				require.Equal(t, "github", cfg.Repositories[0].Provider)
				require.Equal(t, "org/repo", cfg.Repositories[0].Name)
				require.Equal(t, "main", cfg.Repositories[0].Ref)

				require.Len(t, cfg.Copies, 1)
				require.Equal(t, "github", cfg.Copies[0].Repository.Provider)
				require.Equal(t, "org/repo", cfg.Copies[0].Repository.Name)
				require.Equal(t, "main", cfg.Copies[0].Repository.Ref)
				require.Equal(t, "/src", cfg.Copies[0].Paths.Remote)
				require.Equal(t, "/dest", cfg.Copies[0].Paths.Local)

				require.Len(t, cfg.Copies[0].Options.TextReplacements, 1)
				require.Equal(t, "foo", cfg.Copies[0].Options.TextReplacements[0].FromText)
				require.Equal(t, "bar", cfg.Copies[0].Options.TextReplacements[0].ToText)
				require.Equal(t, "*.go", cfg.Copies[0].Options.TextReplacements[0].FileFilterGlob)
				require.Equal(t, "/tmp/archive", cfg.Copies[0].Options.SaveArchiveToPath)
				require.True(t, cfg.Copies[0].Options.CreateGoEmbedForArchive)
			},
		},
		{
			name:   "invalid_json_syntax",
			format: "json",
			config: `{
				"repositories": [
					{
						"provider": "github"
						"name": "org/repo", // Missing comma
						"ref": "main"
					}
				]
			}`,
			expectError: true,
		},
		{
			name:   "invalid_yaml_syntax",
			format: "yaml",
			config: `
repositories:
  - provider: github
    name: org/repo
    ref: main
  copies: # Wrong indentation
  - repository:
      provider: github
`,
			expectError: true,
		},
		{
			name:   "invalid_hcl_syntax",
			format: "hcl",
			config: `
repositories {
  provider = "github"
  name = "org/repo"
  ref = main // Missing quotes
}
`,
			expectError: true,
		},
		{
			name:   "missing_required_fields_json",
			format: "json",
			config: `{
				"repositories": [
					{
						"provider": "github"
					}
				]
			}`,
			expectError: true,
		},
		{
			name:   "missing_required_fields_yaml",
			format: "yaml",
			config: `
repositories:
  - provider: github
`,
			expectError: true,
		},
		{
			name:   "missing_required_fields_hcl",
			format: "hcl",
			config: `
repositories {
  provider = "github"
}
`,
			expectError: true,
		},
		{
			name:   "copyrc_yaml_format",
			format: "copyrc",
			config: `
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
`,
			validate: func(t *testing.T, cfg *CopyrcConfig) {
				require.NotNil(t, cfg)
				require.Len(t, cfg.Repositories, 1)
				require.Equal(t, "github", cfg.Repositories[0].Provider)
				require.Equal(t, "org/repo", cfg.Repositories[0].Name)
				require.Equal(t, "main", cfg.Repositories[0].Ref)
			},
		},
		{
			name:   "copyrc_hcl_format",
			format: "copyrc",
			config: `
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
`,
			validate: func(t *testing.T, cfg *CopyrcConfig) {
				require.NotNil(t, cfg)
				require.Len(t, cfg.Repositories, 1)
				require.Equal(t, "github", cfg.Repositories[0].Provider)
				require.Equal(t, "org/repo", cfg.Repositories[0].Name)
				require.Equal(t, "main", cfg.Repositories[0].Ref)
			},
		},
		{
			name:   "copyrc_invalid_format",
			format: "copyrc",
			config: `
this is not valid YAML or HCL
it should fail to parse
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with appropriate extension
			tmpDir := t.TempDir()
			var configPath string
			switch tt.format {
			case "json":
				configPath = filepath.Join(tmpDir, "config.json")
			case "yaml":
				configPath = filepath.Join(tmpDir, "config.yaml")
			case "hcl":
				configPath = filepath.Join(tmpDir, "config.hcl")
			case "copyrc":
				configPath = filepath.Join(tmpDir, ".copyrc")
			}
			err := os.WriteFile(configPath, []byte(tt.config), 0644)
			require.NoError(t, err)

			// Load config
			cfg, err := LoadConfig(configPath)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			if tt.validate != nil {
				tt.validate(t, cfg)
			}

			// Test config hashing
			hash1 := cfg.Hash()
			require.NotEmpty(t, hash1)

			// Modify config and verify hash changes
			cfg.Repositories[0].Ref = "different-ref"
			hash2 := cfg.Hash()
			require.NotEqual(t, hash1, hash2)
		})
	}
}

func TestHash(t *testing.T) {
	cfg1 := &CopyrcConfig{
		Repositories: []RepositoryDefinition{
			{
				Provider: "github",
				Name:     "org/repo",
				Ref:      "main",
			},
		},
	}

	cfg2 := &CopyrcConfig{
		Repositories: []RepositoryDefinition{
			{
				Provider: "github",
				Name:     "org/repo",
				Ref:      "main",
			},
		},
	}

	cfg3 := &CopyrcConfig{
		Repositories: []RepositoryDefinition{
			{
				Provider: "github",
				Name:     "org/repo",
				Ref:      "different",
			},
		},
	}

	// Same config should have same hash
	assert.Equal(t, cfg1.Hash(), cfg2.Hash())

	// Different config should have different hash
	assert.NotEqual(t, cfg1.Hash(), cfg3.Hash())
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *CopyrcConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty_config",
			config: &CopyrcConfig{
				Repositories: nil,
				Copies:       nil,
			},
			expectError: true,
			errorMsg:    "at least one repository must be defined",
		},
		{
			name: "empty_repositories",
			config: &CopyrcConfig{
				Repositories: []RepositoryDefinition{},
				Copies:       []Copy{},
			},
			expectError: true,
			errorMsg:    "at least one repository must be defined",
		},
		{
			name: "empty_copies",
			config: &CopyrcConfig{
				Repositories: []RepositoryDefinition{
					{
						Provider: "github",
						Name:     "org/repo",
						Ref:      "main",
					},
				},
				Copies: []Copy{},
			},
			expectError: true,
			errorMsg:    "at least one copy operation must be defined",
		},
		{
			name: "missing_provider",
			config: &CopyrcConfig{
				Repositories: []RepositoryDefinition{
					{
						Name: "org/repo",
						Ref:  "main",
					},
				},
				Copies: []Copy{
					{
						Repository: RepositoryDefinition{
							Provider: "github",
							Name:     "org/repo",
							Ref:      "main",
						},
						Paths: CopyPaths{
							Remote: "/src",
							Local:  "/dest",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "repository 0: provider is required",
		},
		{
			name: "missing_name",
			config: &CopyrcConfig{
				Repositories: []RepositoryDefinition{
					{
						Provider: "github",
						Ref:      "main",
					},
				},
				Copies: []Copy{
					{
						Repository: RepositoryDefinition{
							Provider: "github",
							Name:     "org/repo",
							Ref:      "main",
						},
						Paths: CopyPaths{
							Remote: "/src",
							Local:  "/dest",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "repository 0: name is required",
		},
		{
			name: "missing_ref",
			config: &CopyrcConfig{
				Repositories: []RepositoryDefinition{
					{
						Provider: "github",
						Name:     "org/repo",
					},
				},
				Copies: []Copy{
					{
						Repository: RepositoryDefinition{
							Provider: "github",
							Name:     "org/repo",
							Ref:      "main",
						},
						Paths: CopyPaths{
							Remote: "/src",
							Local:  "/dest",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "repository 0: ref is required",
		},
		{
			name: "missing_remote_path",
			config: &CopyrcConfig{
				Repositories: []RepositoryDefinition{
					{
						Provider: "github",
						Name:     "org/repo",
						Ref:      "main",
					},
				},
				Copies: []Copy{
					{
						Repository: RepositoryDefinition{
							Provider: "github",
							Name:     "org/repo",
							Ref:      "main",
						},
						Paths: CopyPaths{
							Local: "/dest",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "copy 0: paths: remote path is required",
		},
		{
			name: "missing_local_path",
			config: &CopyrcConfig{
				Repositories: []RepositoryDefinition{
					{
						Provider: "github",
						Name:     "org/repo",
						Ref:      "main",
					},
				},
				Copies: []Copy{
					{
						Repository: RepositoryDefinition{
							Provider: "github",
							Name:     "org/repo",
							Ref:      "main",
						},
						Paths: CopyPaths{
							Remote: "/src",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "copy 0: paths: local path is required",
		},
		{
			name: "invalid_text_replacement",
			config: &CopyrcConfig{
				Repositories: []RepositoryDefinition{
					{
						Provider: "github",
						Name:     "org/repo",
						Ref:      "main",
					},
				},
				Copies: []Copy{
					{
						Repository: RepositoryDefinition{
							Provider: "github",
							Name:     "org/repo",
							Ref:      "main",
						},
						Paths: CopyPaths{
							Remote: "/src",
							Local:  "/dest",
						},
						Options: CopyOptions{
							TextReplacements: []TextReplacement{
								{
									FromText: "",
									ToText:   "bar",
								},
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "copy 0: options: text replacement 0: from_text is required",
		},
		{
			name: "valid_config",
			config: &CopyrcConfig{
				Repositories: []RepositoryDefinition{
					{
						Provider: "github",
						Name:     "org/repo",
						Ref:      "main",
					},
				},
				Copies: []Copy{
					{
						Repository: RepositoryDefinition{
							Provider: "github",
							Name:     "org/repo",
							Ref:      "main",
						},
						Paths: CopyPaths{
							Remote: "/src",
							Local:  "/dest",
						},
						Options: CopyOptions{
							TextReplacements: []TextReplacement{
								{
									FromText:       "foo",
									ToText:         "bar",
									FileFilterGlob: "*.go",
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
