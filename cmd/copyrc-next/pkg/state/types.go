package state

import (
	"context"
	"sync"
	"time"

	"github.com/walteh/copyrc/cmd/copyrc-next/pkg/config"
	"github.com/walteh/copyrc/cmd/copyrc-next/pkg/remote"
)

// StateManager defines the interface for managing state
type StateManager interface {
	// Load loads the state from disk
	Load(ctx context.Context) error
	// Save saves the state to disk
	Save(ctx context.Context) error
	// Dir returns the directory containing the state
	Dir() string
	// PutRemoteTextFile adds or updates a remote text file in the state
	PutRemoteTextFile(ctx context.Context, file remote.RawTextFile, localPath string) (*RemoteTextFile, error)
	// ValidateLocalState validates that all files match their recorded state
	ValidateLocalState(ctx context.Context) error
	// IsConsistent checks if the current memory state matches the filesystem
	IsConsistent(ctx context.Context) (bool, error)
	// ConfigHash returns a hash of the current configuration
	ConfigHash() string
	// CleanupOrphanedFiles removes files that are no longer referenced in the state
	CleanupOrphanedFiles(ctx context.Context) error
	// Reset resets the state to empty
	Reset(ctx context.Context) error
}

// 🔒 State represents the in-memory state manager with locking capabilities
type State struct {
	mu     sync.RWMutex // protects concurrent access to state
	file   *StateFile   // current state data
	path   string       // path to .copyrc.lock file
	logger *UserLogger  // user-friendly logging
}

// 📄 StateFile represents the on-disk state format
type StateFile struct {
	SchemaVersion   string               `json:"schema_version"`
	LastUpdated     time.Time            `json:"last_updated"`
	Repositories    []Repository         `json:"repositories"`
	RemoteTextFiles []RemoteTextFile     `json:"remote_text_files"`
	GeneratedFiles  []GeneratedFile      `json:"generated_files"`
	Config          *config.StaticConfig `json:"config"`
	ArchiveFiles    []ArchiveFile        `json:"archive_files"`
	RawGzipData     []byte               `json:"raw_gzip_data"`

	decompressedRawGzipData map[string][]byte
}

// 📦 Repository tracks state for a single repository
type Repository struct {
	Provider  string   `json:"provider"`   // e.g. "github"
	Name      string   `json:"name"`       // repository name
	LatestRef string   `json:"latest_ref"` // latest known ref
	Release   *Release `json:"release"`    // state of specific release
}

// 🏷️ Release tracks state for a specific release/ref
type Release struct {
	LastUpdated  time.Time `json:"last_updated"`
	Ref          string    `json:"ref"`           // release ref/tag
	RefHash      string    `json:"ref_hash"`      // hash of the ref
	Archive      *Archive  `json:"archive"`       // state of release archive
	WebPermalink string    `json:"web_permalink"` // link to view release
	License      *License  `json:"license"`       // license information
}

// 📚 Archive tracks state of a release archive
type Archive struct {
	Hash        string `json:"hash"`         // content hash
	ContentType string `json:"content_type"` // MIME type
	DownloadURL string `json:"download_url"` // URL to download
	LocalPath   string `json:"local_path"`   // where archive is stored
}

// ⚖️ License tracks state of repository license
type License struct {
	SPDX            string `json:"spdx"`             // SPDX identifier
	RemotePermalink string `json:"remote_permalink"` // link to license
	LocalPath       string `json:"local_path"`       // where license is stored
}

// 📝 RemoteTextFile tracks state of a copied text file
type RemoteTextFile struct {
	Metadata    map[string]string `json:"metadata"`        // file metadata
	RepoName    string            `json:"repository_name"` // source repository
	ReleaseRef  string            `json:"release_ref"`     // source release ref
	LocalPath   string            `json:"local_path"`      // where file is stored
	LastUpdated time.Time         `json:"last_updated"`    // when file was updated
	IsPatched   bool              `json:"is_patched"`      // whether file is patched
	ContentHash string            `json:"content_hash"`    // hash of content
	Patch       *Patch            `json:"patch"`           // patch information if patched
	Permalink   string            `json:"permalink"`       // link to source
	License     *License          `json:"license"`         // license information
	IsIgnored   bool              `json:"is_ignored"`      // whether file is ignored

	content []byte
}

// 🔄 Patch tracks state of a file patch
type Patch struct {
	PatchDiff     string `json:"patch_diff"`     // patch in gopatch format
	PatchHash     string `json:"patch_hash"`     // hash of patch
	RemoteContent string `json:"remote_content"` // original content (gzipped)
	PatchPath     string `json:"patch_path"`     // where patch is stored
}

// ⚙️ GeneratedFile tracks state of a generated file
type GeneratedFile struct {
	LocalPath     string    `json:"local_path"`     // where file is stored
	LastUpdated   time.Time `json:"last_updated"`   // when file was generated
	ReferenceFile string    `json:"reference_file"` // source file reference
}

// ArchiveFile represents a downloaded archive file
type ArchiveFile struct {
	LocalPath   string // Path to the local file
	ContentHash string // SHA-256 hash of the file content
}
