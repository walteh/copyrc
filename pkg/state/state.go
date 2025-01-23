package state

import (
	"context"
	"io"
	"time"

	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/remote"
	"gitlab.com/tozd/go/errors"
)

// State is the top-level state tracking structure that manages file state and cleanup
type State struct {
	LastUpdated time.Time `json:"last_updated"`

	// ConfigHash is used to detect if the state matches the current config
	ConfigHash string `json:"config_hash"`

	// Repositories tracks the state of each repository we're copying from
	Repositories []RepositoryState `json:"repositories"`

	// RemoteTextFiles tracks all text files copied from remote sources
	RemoteTextFiles []RemoteTextFile `json:"remote_text_files"`

	// GeneratedFiles tracks files that were generated locally
	GeneratedFiles []GeneratedFile `json:"generated_files"`
}

// RepositoryState tracks the state of a specific repository
type RepositoryState struct {
	Provider  string `json:"provider"`
	Name      string `json:"name"`
	LatestRef string `json:"latest_ref"`

	// TargetRelease represents the specific release we're copying from
	TargetRelease ReleaseState `json:"target_release"`
}

// ReleaseState tracks the state of a specific release
type ReleaseState struct {
	LastUpdated time.Time `json:"last_updated"`
	Ref         string    `json:"ref"`
	RefHash     string    `json:"ref_hash"`

	// Archive holds information about the release's tarball
	Archive ArchiveState `json:"archive"`

	// WebPermalink is a link to view this release on the web
	WebPermalink string `json:"web_permalink"`

	// License holds information about the repository's license
	License LicenseState `json:"license"`
}

// ArchiveState tracks the state of a release's archive/tarball
type ArchiveState struct {
	Hash        string `json:"hash"`
	ContentType string `json:"content_type"`
	DownloadURL string `json:"download_url"`
	LocalPath   string `json:"local_path,omitempty"` // Only set if save_archive_to_path is true
}

// LicenseState tracks the state of a repository's license
type LicenseState struct {
	SPDX            string `json:"spdx"`
	RemotePermalink string `json:"remote_permalink"`
	LocalPath       string `json:"local_path,omitempty"` // Only set if save_archive_to_path is true
}

// RemoteTextFile represents a text file copied from a remote source
type RemoteTextFile struct {
	Metadata    map[string]string `json:"metadata"`
	RepoName    string            `json:"repository_name"`
	ReleaseRef  string            `json:"release_ref"`
	LocalPath   string            `json:"local_path"` // Will have either .copy. or .patch. suffix
	LastUpdated time.Time         `json:"last_updated"`
	IsPatched   bool              `json:"is_patched"`
	ContentHash string            `json:"remote_content_hash"`

	// Patch holds information about file patches if this is a patched file
	Patch *PatchState `json:"patch,omitempty"`

	// Permalink is a direct link to the file in its remote source
	Permalink string `json:"permalink"`

	// TextReplacements tracks any text replacements applied to this file
	TextReplacements []config.TextReplacement `json:"text_replacements"`

	// License holds information about this file's license
	License *LicenseState `json:"license,omitempty"`
}

// PatchState tracks patch-related information for a file
type PatchState struct {
	PatchDiff            string `json:"patch_diff"` // gopatch format
	GzippedRemoteContent string `json:"gzipped_remote_content"`
	PatchPath            string `json:"patch_path"`
}

// GeneratedFile represents a file that was generated locally
type GeneratedFile struct {
	LocalPath     string    `json:"local_path"`
	LastUpdated   time.Time `json:"last_updated"`
	ReferenceFile string    `json:"reference_file"`
}

// LoadState loads state from a .copyrc.lock file
func LoadState(ctx context.Context, path string) (*State, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", path).Msg("loading state")

	// TODO(dr.methodical): 🔨 Implement state loading from JSON file
	return nil, errors.Errorf("not implemented")
}

// WriteState writes state to a .copyrc.lock file
func WriteState(ctx context.Context, path string, state *State) error {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", path).Msg("writing state")

	// TODO(dr.methodical): 🔨 Implement state writing to JSON file
	return errors.Errorf("not implemented")
}

// PutRemoteTextFile adds or updates a remote text file in the state
func (s *State) PutRemoteTextFile(ctx context.Context, file remote.RawTextFile, localPath string) (*RemoteTextFile, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", localPath).Msg("putting remote text file")

	// TODO(dr.methodical): 🔨 Implement remote text file state tracking
	return nil, errors.Errorf("not implemented")
}

// PutGeneratedFile adds or updates a generated file in the state
func (s *State) PutGeneratedFile(ctx context.Context, file *GeneratedFile) (*GeneratedFile, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", file.LocalPath).Msg("putting generated file")

	// TODO(dr.methodical): 🔨 Implement generated file state tracking
	return nil, errors.Errorf("not implemented")
}

// PutArchiveFile adds or updates an archive file in the state
func (s *State) PutArchiveFile(ctx context.Context, release remote.Release, localPath string) (*ArchiveState, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", localPath).Msg("putting archive file")

	// TODO(dr.methodical): 🔨 Implement archive file state tracking
	return nil, errors.Errorf("not implemented")
}

// RawRemoteContent returns the raw content of a remote text file
func (f *RemoteTextFile) RawRemoteContent() (io.ReadCloser, error) {
	// TODO(dr.methodical): 🔨 Implement raw content reading
	return nil, errors.Errorf("not implemented")
}

// RawPatchContent returns the raw content of a patch file
func (f *RemoteTextFile) RawPatchContent() (io.ReadCloser, error) {
	// TODO(dr.methodical): 🔨 Implement patch content reading
	return nil, errors.Errorf("not implemented")
}

// ApplyModificationToRawRemoteContent applies a text modification to the file's content
func (f *RemoteTextFile) ApplyModificationToRawRemoteContent(ctx context.Context, mod config.TextReplacement) error {
	// TODO(dr.methodical): 🔨 Implement text modification
	return errors.Errorf("not implemented")
}

// TODO(dr.methodical): 🧪 Add tests for state loading/saving
// TODO(dr.methodical): 🧪 Add tests for file operations
// TODO(dr.methodical): 🧪 Add tests for text modifications
// TODO(dr.methodical): 📝 Add examples of state usage
