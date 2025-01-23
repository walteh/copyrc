package state

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/remote"
	"gitlab.com/tozd/go/errors"
)

// CurrentSchemaVersion is the current version of the state file schema
const CurrentSchemaVersion = "1.0.0"

// 🔒 State represents the in-memory state manager with locking capabilities
type State struct {
	mu     sync.RWMutex // protects concurrent access to state
	file   *StateFile   // current state data
	path   string       // path to .copyrc.lock file
	logger *UserLogger  // user-friendly logging
}

// 📄 StateFile represents the on-disk state format
type StateFile struct {
	SchemaVersion   string                 `json:"schema_version"`
	LastUpdated     time.Time              `json:"last_updated"`
	Repositories    []Repository           `json:"repositories"`
	RemoteTextFiles []RemoteTextFile       `json:"remote_text_files"`
	GeneratedFiles  []GeneratedFile        `json:"generated_files"`
	Config          map[string]interface{} `json:"config"`
	ArchiveFiles    []ArchiveFile          `json:"archive_files"`
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
}

// 🔄 Patch tracks state of a file patch
type Patch struct {
	PatchDiff     string `json:"patch_diff"`     // patch in gopatch format
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

// 🆕 New creates a new state manager for the given directory
func New(dir string) (*State, error) {
	path := filepath.Join(dir, ".copyrc.lock")
	return &State{
		path: path,
		file: &StateFile{
			SchemaVersion: CurrentSchemaVersion,
			LastUpdated:   time.Now(),
		},
		logger: NewUserLogger(context.Background()),
	}, nil
}

// 💾 Load reads the state from disk, creating a new one if it doesn't exist
func (s *State) Load(ctx context.Context) error {
	s.logger = NewUserLogger(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if file exists
	_, err := os.Stat(s.path)
	if os.IsNotExist(err) {
		s.logger.LogStateChange("Starting with clean state")
		return nil // Clean state already initialized in New()
	}
	if err != nil {
		return errors.Errorf("checking state file: %w", err)
	}

	// Read and parse state file
	data, err := os.ReadFile(s.path)
	if err != nil {
		return errors.Errorf("reading state file: %w", err)
	}

	if err := json.Unmarshal(data, s.file); err != nil {
		return errors.Errorf("parsing state file: %w", err)
	}

	s.logger.LogStateChange("Loaded existing state")
	return nil
}

// 💾 Save writes the current state to disk with file locking
func (s *State) Save(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create lock file
	lockPath := s.path + ".lock"

	// Check if lock file exists
	if _, err := os.Stat(lockPath); err == nil {
		s.logger.LogLockOperation(false, s.path, errors.Errorf("lock file already exists"))
		return errors.Errorf("creating lock file: lock file already exists")
	} else if !os.IsNotExist(err) {
		s.logger.LogLockOperation(false, s.path, err)
		return errors.Errorf("checking lock file: %w", err)
	}

	// Create lock file
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		s.logger.LogLockOperation(false, s.path, err)
		return errors.Errorf("creating lock file: %w", err)
	}
	defer os.Remove(lockPath) // Clean up lock file
	defer lockFile.Close()

	s.logger.LogLockOperation(true, s.path, nil)
	defer s.logger.LogLockOperation(false, s.path, nil)

	// Update timestamp
	s.file.LastUpdated = time.Now()

	// Marshal state
	data, err := json.MarshalIndent(s.file, "", "  ")
	if err != nil {
		return errors.Errorf("marshaling state: %w", err)
	}

	// Write to temp file first
	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return errors.Errorf("writing temp state file: %w", err)
	}
	defer os.Remove(tempPath) // Clean up temp file

	// Rename temp file to actual state file
	if err := os.Rename(tempPath, s.path); err != nil {
		return errors.Errorf("replacing state file: %w", err)
	}

	s.logger.LogStateChange("Saved state file")
	return nil
}

// TODO: Add validation methods
// TODO: Add methods for adding/removing files
// TODO: Add methods for checking consistency

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

// 🔏 hashFile computes the SHA-256 hash of a file
func (s *State) hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Errorf("opening file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", errors.Errorf("reading file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// 🔏 hashContent computes the SHA-256 hash of content from a reader
func (s *State) hashContent(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", errors.Errorf("reading content: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// 🔄 PutRemoteTextFile adds or updates a remote text file in the state
func (s *State) PutRemoteTextFile(ctx context.Context, file remote.RawTextFile, localPath string) (*RemoteTextFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", localPath).Msg("putting remote text file")

	// Validate file suffix
	if !strings.Contains(localPath, ".copy.") && !strings.Contains(localPath, ".patch.") {
		return nil, errors.Errorf("invalid file suffix: %s (must contain .copy. or .patch.)", localPath)
	}

	// Get content hash and data
	content, err := file.GetContent(ctx)
	if err != nil {
		return nil, errors.Errorf("getting file content: %w", err)
	}
	defer content.Close()

	// Read all content
	data, err := io.ReadAll(content)
	if err != nil {
		return nil, errors.Errorf("reading file content: %w", err)
	}

	// Write content to file
	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return nil, errors.Errorf("writing file content: %w", err)
	}

	// Compute hash of content
	hash, err := s.hashContent(bytes.NewReader(data))
	if err != nil {
		return nil, errors.Errorf("hashing content: %w", err)
	}

	// Create or update file entry
	remoteFile := &RemoteTextFile{
		Metadata:    make(map[string]string),
		RepoName:    file.Release().Repository().Name(),
		ReleaseRef:  file.Release().Ref(),
		LocalPath:   localPath,
		LastUpdated: time.Now(),
		IsPatched:   false, // Will be set to true if/when patches are applied
		ContentHash: hash,
		Permalink:   file.WebViewPermalink(),
	}

	// Update state
	found := false
	for i, f := range s.file.RemoteTextFiles {
		if f.LocalPath == localPath {
			s.file.RemoteTextFiles[i] = *remoteFile
			found = true
			break
		}
	}
	if !found {
		s.file.RemoteTextFiles = append(s.file.RemoteTextFiles, *remoteFile)
	}

	s.logger.LogFileChange(FileChange{
		Type:        FileUpdated,
		Path:        localPath,
		Description: "Updated from remote",
	})

	return remoteFile, nil
}

// 🔨 PutGeneratedFile adds or updates a generated file in the state
func (s *State) PutGeneratedFile(ctx context.Context, file GeneratedFile) (*GeneratedFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", file.LocalPath).Msg("putting generated file")

	// Validate file exists
	if _, err := os.Stat(file.LocalPath); os.IsNotExist(err) {
		return nil, errors.Errorf("file does not exist: %s", file.LocalPath)
	} else if err != nil {
		return nil, errors.Errorf("checking file: %w", err)
	}

	// Validate reference file exists
	if _, err := os.Stat(file.ReferenceFile); os.IsNotExist(err) {
		return nil, errors.Errorf("reference file does not exist: %s", file.ReferenceFile)
	} else if err != nil {
		return nil, errors.Errorf("checking reference file: %w", err)
	}

	// Create or update file entry
	generatedFile := &GeneratedFile{
		LocalPath:     file.LocalPath,
		LastUpdated:   time.Now(),
		ReferenceFile: file.ReferenceFile,
	}

	// Update state
	found := false
	for i, f := range s.file.GeneratedFiles {
		if f.LocalPath == file.LocalPath {
			s.file.GeneratedFiles[i] = *generatedFile
			found = true
			break
		}
	}
	if !found {
		s.file.GeneratedFiles = append(s.file.GeneratedFiles, *generatedFile)
	}

	s.logger.LogFileChange(FileChange{
		Type:        FileUpdated,
		Path:        file.LocalPath,
		Description: "Updated generated file",
	})

	return generatedFile, nil
}

// 📦 PutArchiveFile adds or updates an archive file in the state
func (s *State) PutArchiveFile(ctx context.Context, file ArchiveFile) (*ArchiveFile, error) {
	s.logger.LogFileOperation(ctx, "putting archive file", file.LocalPath)

	// Validate file exists
	if _, err := os.Stat(file.LocalPath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Errorf("file does not exist: %s", file.LocalPath)
		}
		return nil, errors.Errorf("checking file: %w", err)
	}

	// Compute hash of archive
	hash, err := s.hashFile(file.LocalPath)
	if err != nil {
		return nil, errors.Errorf("computing hash: %w", err)
	}

	// Check if file already exists in state
	for i, existing := range s.file.ArchiveFiles {
		if existing.LocalPath == file.LocalPath {
			// Update existing entry
			s.file.ArchiveFiles[i] = file
			s.file.ArchiveFiles[i].ContentHash = hash
			s.logger.LogFileOperation(ctx, "Updated archive file", file.LocalPath)
			return &s.file.ArchiveFiles[i], nil
		}
	}

	// Add new entry
	file.ContentHash = hash
	s.file.ArchiveFiles = append(s.file.ArchiveFiles, file)
	s.logger.LogFileOperation(ctx, "Added archive file", file.LocalPath)
	return &s.file.ArchiveFiles[len(s.file.ArchiveFiles)-1], nil
}

// 📄 RawRemoteContent returns the raw content of a remote text file
func (f *RemoteTextFile) RawRemoteContent() (io.ReadCloser, error) {
	// For patched files, we need to read from the gzipped content
	if f.IsPatched && f.Patch != nil && f.Patch.RemoteContent != "" {
		// Decode base64 content
		gzipData, err := base64.StdEncoding.DecodeString(f.Patch.RemoteContent)
		if err != nil {
			return nil, errors.Errorf("decoding base64 content: %w", err)
		}

		// Create gzip reader
		gzipReader, err := gzip.NewReader(bytes.NewReader(gzipData))
		if err != nil {
			return nil, errors.Errorf("creating gzip reader: %w", err)
		}

		return gzipReader, nil
	}

	// For non-patched files, read directly from the local path
	file, err := os.Open(f.LocalPath)
	if err != nil {
		return nil, errors.Errorf("opening file: %w", err)
	}

	return file, nil
}

// 🔄 RawPatchContent returns the raw content of a patch file
func (f *RemoteTextFile) RawPatchContent() (io.ReadCloser, error) {
	if !f.IsPatched {
		return nil, errors.Errorf("file is not patched")
	}

	if f.Patch == nil || f.Patch.PatchPath == "" {
		return nil, errors.Errorf("no patch information available")
	}

	file, err := os.Open(f.Patch.PatchPath)
	if err != nil {
		return nil, errors.Errorf("opening patch file: %w", err)
	}

	return file, nil
}

// 🔄 ApplyModificationToRawRemoteContent applies a text modification to the file's content
func (f *RemoteTextFile) ApplyModificationToRawRemoteContent(ctx context.Context, mod config.TextReplacement) error {
	// Get raw content
	content, err := f.RawRemoteContent()
	if err != nil {
		return errors.Errorf("getting raw content: %w", err)
	}
	defer content.Close()

	// Read all content
	data, err := io.ReadAll(content)
	if err != nil {
		return errors.Errorf("reading content: %w", err)
	}

	// Apply text replacement
	modified := strings.ReplaceAll(string(data), mod.FromText, mod.ToText)

	// Create patch if needed
	if !f.IsPatched {
		// Compress original content
		var gzipBuf bytes.Buffer
		gzipWriter := gzip.NewWriter(&gzipBuf)
		if _, err := gzipWriter.Write(data); err != nil {
			return errors.Errorf("compressing content: %w", err)
		}
		if err := gzipWriter.Close(); err != nil {
			return errors.Errorf("closing gzip writer: %w", err)
		}

		// Convert .copy. to .patch. in the path
		patchPath := strings.Replace(f.LocalPath, ".copy.", ".patch.", 1)
		if patchPath == f.LocalPath {
			return errors.Errorf("file path must contain .copy.: %s", f.LocalPath)
		}

		// Store compressed content
		f.Patch = &Patch{
			RemoteContent: base64.StdEncoding.EncodeToString(gzipBuf.Bytes()),
			PatchPath:     patchPath,
		}
		f.IsPatched = true

		// Write patch file
		if err := os.WriteFile(patchPath, []byte(f.Patch.PatchDiff), 0644); err != nil {
			return errors.Errorf("writing patch file: %w", err)
		}
	}

	// Write modified content
	if err := os.WriteFile(f.LocalPath, []byte(modified), 0644); err != nil {
		return errors.Errorf("writing modified content: %w", err)
	}

	// Update content hash
	f.ContentHash = fmt.Sprintf("%x", sha256.Sum256([]byte(modified)))

	// Generate patch diff
	f.Patch.PatchDiff = fmt.Sprintf("--- %s\n+++ %s\n@@ -1,1 +1,1 @@\n-%s\n+%s\n",
		f.LocalPath, f.LocalPath, string(data), modified)

	return nil
}

// 🔍 ValidateLocalState checks that all files match their recorded state
func (s *State) ValidateLocalState(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate schema version
	if s.file.SchemaVersion != CurrentSchemaVersion {
		return errors.Errorf("invalid schema version: expected %s, got %s", CurrentSchemaVersion, s.file.SchemaVersion)
	}

	// Validate remote text files
	for _, f := range s.file.RemoteTextFiles {
		// Validate file suffix
		if !strings.Contains(f.LocalPath, ".copy.") && !strings.Contains(f.LocalPath, ".patch.") {
			return errors.Errorf("invalid file suffix: %s (must contain .copy. or .patch.)", f.LocalPath)
		}

		// Check file exists
		if _, err := os.Stat(f.LocalPath); os.IsNotExist(err) {
			return errors.Errorf("file does not exist: %s", f.LocalPath)
		}

		// Check content hash
		hash, err := s.hashFile(f.LocalPath)
		if err != nil {
			return errors.Errorf("computing hash for %s: %w", f.LocalPath, err)
		}
		if hash != f.ContentHash {
			return errors.Errorf("content hash mismatch for %s: expected %s, got %s", f.LocalPath, f.ContentHash, hash)
		}

		// Check patch file if patched
		if f.IsPatched && f.Patch != nil {
			if _, err := os.Stat(f.Patch.PatchPath); os.IsNotExist(err) {
				return errors.Errorf("patch file does not exist: %s", f.Patch.PatchPath)
			}
		}
	}

	// Validate repositories
	for _, r := range s.file.Repositories {
		if r.Release != nil && r.Release.Archive != nil {
			a := r.Release.Archive
			// Check archive exists
			if _, err := os.Stat(a.LocalPath); os.IsNotExist(err) {
				return errors.Errorf("archive file does not exist: %s", a.LocalPath)
			}

			// Check archive hash
			hash, err := s.hashFile(a.LocalPath)
			if err != nil {
				return errors.Errorf("computing hash for %s: %w", a.LocalPath, err)
			}
			if hash != a.Hash {
				return errors.Errorf("archive hash mismatch for %s: expected %s, got %s", a.LocalPath, a.Hash, hash)
			}
		}
	}

	return nil
}

// 🔍 IsConsistent checks if the current memory state matches the filesystem
func (s *State) IsConsistent(ctx context.Context) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Quick check - verify all files exist and have correct hashes
	for _, file := range s.file.RemoteTextFiles {
		if err := s.validateFile(ctx, file.LocalPath, file.ContentHash); err != nil {
			s.logger.LogValidation(false, "State inconsistency detected", err)
			return false, nil
		}
	}

	s.logger.LogValidation(true, "State is consistent", nil)
	return true, nil
}

// 🔐 validateFile checks if a file exists and optionally verifies its hash
func (s *State) validateFile(ctx context.Context, path string, expectedHash string) error {
	// Check file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("file does not exist: %s", path)
		}
		return errors.Errorf("checking file: %w", err)
	}

	// Check it's a regular file
	if !info.Mode().IsRegular() {
		return errors.Errorf("not a regular file: %s", path)
	}

	// If hash provided, verify it
	if expectedHash != "" {
		hash, err := s.hashFile(path)
		if err != nil {
			return errors.Errorf("hashing file: %w", err)
		}
		if hash != expectedHash {
			return errors.Errorf("hash mismatch for %s: expected %s, got %s", path, expectedHash, hash)
		}
	}

	return nil
}

// 🧹 CleanupOrphanedFiles removes files that are no longer referenced in the state
func (s *State) CleanupOrphanedFiles(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Build set of known paths
	knownPaths := make(map[string]bool)

	// Add remote text files and their patches
	for _, file := range s.file.RemoteTextFiles {
		knownPaths[file.LocalPath] = true
		if file.IsPatched && file.Patch != nil {
			knownPaths[file.Patch.PatchPath] = true
		}
	}

	// Add generated files
	for _, file := range s.file.GeneratedFiles {
		knownPaths[file.LocalPath] = true
	}

	// Add archives
	for _, repo := range s.file.Repositories {
		if repo.Release != nil && repo.Release.Archive != nil {
			knownPaths[repo.Release.Archive.LocalPath] = true
		}
	}

	// Walk directory to find unknown files
	baseDir := filepath.Dir(s.path)
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and the state file itself
		if info.IsDir() || path == s.path {
			return nil
		}

		// Check if file matches our patterns
		if strings.Contains(path, ".copy.") || strings.Contains(path, ".patch.") {
			if !knownPaths[path] {
				// Found an orphaned file
				s.logger.LogFileChange(FileChange{
					Type:        FileDeleted,
					Path:        path,
					Description: "Removing orphaned file",
				})
				if err := os.Remove(path); err != nil {
					return errors.Errorf("removing orphaned file: %s: %w", path, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return errors.Errorf("walking directory: %w", err)
	}

	return nil
}

// 🏠 Dir returns the directory where the state file is located
func (s *State) Dir() string {
	return filepath.Dir(s.path)
}

// TODO(dr.methodical): 🧪 Add tests for state loading/saving
// TODO(dr.methodical): 🧪 Add tests for file operations
// TODO(dr.methodical): 🧪 Add tests for text modifications
// TODO(dr.methodical): 📝 Add examples of state usage
