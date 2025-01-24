package config

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"gitlab.com/tozd/go/errors"
)

type Config interface {
	GetLocation() string
	GetHash() string
	GetRepositories() []RepositoryDefinition
	GetCopies() []Copy
}

var _ Config = &StaticConfig{}

type StaticConfig struct {
	Location     string                 `json:"location" yaml:"location" hcl:"location,attr" cty:"location"`
	Repositories []RepositoryDefinition `json:"repositories" yaml:"repositories" hcl:"repositories,block" cty:"repositories"`
	Copies       []Copy                 `json:"copies"       yaml:"copies"       hcl:"copy,block"         cty:"copies"`
	Hash         string                 `json:"hash"         yaml:"hash"         hcl:"hash,attr"          cty:"hash"`
}

func NewStaticConfig(ctx context.Context, cfg Config) *StaticConfig {
	return &StaticConfig{
		Location:     cfg.GetLocation(),
		Repositories: cfg.GetRepositories(),
		Copies:       cfg.GetCopies(),
		Hash:         cfg.GetHash(),
	}
}

func (c *StaticConfig) GetLocation() string {
	return c.Location
}

func (c *StaticConfig) GetHash() string {
	return c.Hash
}

func (c *StaticConfig) GetRepositories() []RepositoryDefinition {
	return c.Repositories
}

func (c *StaticConfig) GetCopies() []Copy {
	return c.Copies
}

// CopyrcConfig is the top-level configuration structure defining what to copy from where
type CopyrcConfig struct {
	location string
	// Repositories defines the list of remote repositories to copy from
	Repositories []RepositoryDefinition `json:"repositories" yaml:"repositories" hcl:"repositories,block" cty:"repositories"`

	// Copies defines the list of copy operations to perform
	Copies []Copy `json:"copies" yaml:"copies" hcl:"copy,block" cty:"copies"`
}

var _ Config = &CopyrcConfig{}

// Hash returns a hash of the config used to detect changes
func (c *CopyrcConfig) Hash() string {
	data, _ := json.Marshal(c)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (c *CopyrcConfig) GetLocation() string {
	return c.location
}

func (c *CopyrcConfig) GetHash() string {
	return c.Hash()
}

// Validate checks that all required fields are set
func Validate(ctx context.Context, cfg Config) error {
	if len(cfg.GetRepositories()) == 0 {
		return errors.Errorf("at least one repository must be defined")
	}

	for i, repo := range cfg.GetRepositories() {
		if err := repo.Validate(); err != nil {
			return errors.Errorf("repository %d: %w", i, err)
		}
	}

	if len(cfg.GetCopies()) == 0 {
		return errors.Errorf("at least one copy operation must be defined")
	}

	for i, copy := range cfg.GetCopies() {
		if err := copy.Validate(); err != nil {
			return errors.Errorf("copy %d: %w", i, err)
		}
	}

	return nil
}

// GetRepositories returns the list of repositories to sync
func (c *CopyrcConfig) GetRepositories() []RepositoryDefinition {
	return c.Repositories
}

// GetCopies returns the list of copy operations
func (c *CopyrcConfig) GetCopies() []Copy {
	return c.Copies
}

// Copy defines a single copy operation from a remote source to local destination
type Copy struct {
	// Repository identifies the source repository and version
	Repository RepositoryDefinition `json:"repository" yaml:"repository" hcl:"repository,block" cty:"repository"`

	// Paths defines the source and destination paths
	Paths CopyPaths `json:"paths" yaml:"paths" hcl:"paths,block" cty:"paths"`

	// Options defines additional copy options
	Options CopyOptions `json:"options" yaml:"options" hcl:"options,block" cty:"options"`
}

// Validate checks that all required fields are set
func (c *Copy) Validate() error {
	if err := c.Repository.Validate(); err != nil {
		return errors.Errorf("repository: %w", err)
	}

	if err := c.Paths.Validate(); err != nil {
		return errors.Errorf("paths: %w", err)
	}

	if err := c.Options.Validate(); err != nil {
		return errors.Errorf("options: %w", err)
	}

	return nil
}

// RepositoryDefinition identifies a specific remote repository and version
type RepositoryDefinition struct {
	// Provider is the repository provider (e.g. "github")
	Provider string `json:"provider" yaml:"provider" hcl:"provider,attr" cty:"provider"`

	// Name is the repository name (e.g. "owner/repo")
	Name string `json:"name" yaml:"name" hcl:"name,attr" cty:"name"`

	// Ref is the git reference (tag, branch, or commit)
	Ref string `json:"ref" yaml:"ref" hcl:"ref,attr" cty:"ref"`
}

// Validate checks that all required fields are set
func (r *RepositoryDefinition) Validate() error {
	if r.Provider == "" {
		return errors.Errorf("provider is required")
	}
	if r.Name == "" {
		return errors.Errorf("name is required")
	}
	if r.Ref == "" {
		return errors.Errorf("ref is required")
	}
	return nil
}

// CopyPaths defines the source and destination paths for a copy operation
type CopyPaths struct {
	// Remote is the path in the remote repository
	Remote string `json:"remote" yaml:"remote" hcl:"remote,attr" cty:"remote"`

	// Local is the path on the local filesystem
	Local string `json:"local" yaml:"local" hcl:"local,attr" cty:"local"`
}

// Validate checks that all required fields are set
func (p *CopyPaths) Validate() error {
	if p.Remote == "" {
		return errors.Errorf("remote path is required")
	}
	if p.Local == "" {
		return errors.Errorf("local path is required")
	}
	return nil
}

// CopyOptions defines additional options for a copy operation
type CopyOptions struct {
	// TextReplacements defines text replacements to apply to copied files
	TextReplacements []TextReplacement `json:"text_replacements" yaml:"text_replacements" hcl:"text_replacements,optional" cty:"text_replacements"`

	// SaveArchiveToPath is the path to save the repository archive to
	SaveArchiveToPath string `json:"save_archive_to_path" yaml:"save_archive_to_path" hcl:"save_archive_to_path,optional" cty:"save_archive_to_path"`

	// CreateGoEmbedForArchive indicates whether to create a go:embed file for the archive
	CreateGoEmbedForArchive bool `json:"create_go_embed_for_archive" yaml:"create_go_embed_for_archive" hcl:"create_go_embed_for_archive,optional" cty:"create_go_embed_for_archive"`

	// IgnoreFilesGlobs is a list of glob patterns to ignore files in the copy operation
	IgnoreFilesGlobs []string `json:"ignore_files_globs" yaml:"ignore_files_globs" hcl:"ignore_files_globs,optional" cty:"ignore_files_globs"`
}

// Validate checks that all required fields are set
func (o *CopyOptions) Validate() error {
	for i, tr := range o.TextReplacements {
		if err := tr.Validate(); err != nil {
			return errors.Errorf("text replacement %d: %w", i, err)
		}
	}
	return nil
}

// TextReplacement defines a text replacement rule
type TextReplacement struct {
	// FromText is the text to replace
	FromText string `json:"from_text" yaml:"from_text" hcl:"from_text,attr" cty:"from_text"`

	// ToText is the replacement text
	ToText string `json:"to_text" yaml:"to_text" hcl:"to_text,attr" cty:"to_text"`

	// FileFilterGlob is a glob pattern to filter which files to apply the replacement to
	FileFilterGlob string `json:"file_filter_glob" yaml:"file_filter_glob" hcl:"file_filter_glob,attr" cty:"file_filter_glob"`
}

// Validate checks that all required fields are set
func (t *TextReplacement) Validate() error {
	if t.FromText == "" {
		return errors.Errorf("from_text is required")
	}
	if t.ToText == "" {
		return errors.Errorf("to_text is required")
	}
	if t.FileFilterGlob == "" {
		return errors.Errorf("file_filter_glob is required")
	}
	return nil
}

// TODO(dr.methodical): 🧪 Add tests for config loading/validation
// TODO(dr.methodical): 🧪 Add tests for Hash() method
// TODO(dr.methodical): 📝 Add examples of config usage

func FilterCopiesByRepository(copies []Copy, repository RepositoryDefinition) []Copy {
	return slices.DeleteFunc(copies, func(copy Copy) bool {
		return copy.Repository.Name != repository.Name ||
			copy.Repository.Provider != repository.Provider ||
			copy.Repository.Ref != repository.Ref
	})
}
