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
	"time"

	"github.com/rs/zerolog"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/remote"
	"github.com/walteh/copyrc/pkg/text"
	"gitlab.com/tozd/go/errors"
)

// CurrentSchemaVersion is the current version of the state file schema
const CurrentSchemaVersion = "1.0.0"

// 🆕 New creates a new state manager for the given directory
func New(ctx context.Context, cfg config.Config) (*State, error) {
	zerolog.Ctx(ctx).Debug().Str("path", cfg.GetLocation()).Msg("creating state manager")
	path := filepath.Join(filepath.Dir(cfg.GetLocation()), ".copyrc.lock")
	st := &State{
		path: path,
		file: &StateFile{
			SchemaVersion:           CurrentSchemaVersion,
			LastUpdated:             time.Now(),
			Config:                  config.NewStaticConfig(ctx, cfg),
			RemoteTextFiles:         make([]RemoteTextFile, 0),
			GeneratedFiles:          make([]GeneratedFile, 0),
			Repositories:            make([]Repository, 0),
			ArchiveFiles:            make([]ArchiveFile, 0),
			RawGzipData:             make([]byte, 0),
			decompressedRawGzipData: make(map[string][]byte),
		},
		logger: NewUserLogger(context.Background()),
	}

	return st, nil
}

// 💾 Load reads the state from disk, creating a new one if it doesn't exist
func LoadExisting(ctx context.Context, cfg config.Config) (*State, error) {
	s, err := New(ctx, cfg)
	if err != nil {
		return nil, errors.Errorf("creating state manager: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if file exists
	_, err = os.Stat(s.path)
	if err != nil {
		return nil, errors.Errorf("checking state file: %w", err)
	}

	// Read and parse state file
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, errors.Errorf("reading state file: %w", err)
	}

	if err := json.Unmarshal(data, s.file); err != nil {
		return nil, errors.Errorf("parsing state file: %w", err)
	}

	if err := s.DecompressRawGzipData(ctx); err != nil {
		return nil, errors.Errorf("decompressing raw gzip data: %w", err)
	}

	// Load configuration from parent director

	s.logger.LogStateChange("Loaded existing state")
	return s, nil
}

func (s *State) CompressRawGzipData(ctx context.Context) error {

	// if already compressed, return

	if s.file.decompressedRawGzipData == nil {
		return nil
	}

	// json marshal the data
	data, err := json.Marshal(s.file.decompressedRawGzipData)
	if err != nil {
		return errors.Errorf("marshaling raw gzip data: %w", err)
	}

	// gzip the data
	buf := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(buf)
	gz.Write(data)
	gz.Close()

	s.file.RawGzipData = buf.Bytes()

	return nil
}

func (s *State) DecompressRawGzipData(ctx context.Context) error {

	// ungzip the data
	buf := bytes.NewBuffer(s.file.RawGzipData)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return errors.Errorf("creating gzip reader: %w", err)
	}
	defer gz.Close()

	dec := json.NewDecoder(gz)
	var rawGzipData map[string][]byte
	if err := dec.Decode(&rawGzipData); err != nil {
		return errors.Errorf("unmarshaling raw gzip data: %w", err)
	}

	s.file.decompressedRawGzipData = rawGzipData

	return nil
}

// 💾 Save writes the current state to disk with file locking
func (s *State) Save(ctx context.Context) error {

	// Validate local state

	// write all files to disk
	if err := s.WriteAllFiles(ctx); err != nil {
		return errors.Errorf("writing all files: %w", err)
	}

	// cleanup old files
	if err := s.CleanupOrphanedFiles(ctx); err != nil {
		return errors.Errorf("cleaning up old files: %w", err)
	}

	if err := s.CompressRawGzipData(ctx); err != nil {
		return errors.Errorf("compressing raw gzip data: %w", err)
	}

	// validate local state
	if err := s.ValidateLocalState(ctx); err != nil {
		return errors.Errorf("validating local state: %w", err)
	}

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

	zerolog.Ctx(ctx).Debug().Interface("state", s.file).Msg("state")

	// Marshal state
	data, err := json.MarshalIndent(s.file, "", "\t")
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
func LoadState(ctx context.Context, cfg config.Config) (*State, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", cfg.GetLocation()).Msg("loading state")

	state, err := New(ctx, cfg)
	if err != nil {
		return nil, errors.Errorf("creating state manager: %w", err)
	}

	// existing, err := LoadExisting(ctx, cfg)
	// if err != nil {
	// 	return nil, errors.Errorf("loading existing state: %w", err)
	// }

	// if err := state.Load(ctx); err != nil {
	// 	return nil, errors.Errorf("loading state: %w", err)
	// }

	zerolog.Ctx(ctx).Debug().Interface("state", state.file).Msg("state")

	return state, nil
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

type PutRemoteTextFileOptions struct {
	IsIgnored bool
}

// 🔄 PutRemoteTextFile adds or updates a remote text file in the state
func (s *State) PutRemoteTextFile(ctx context.Context, file remote.RawTextFile, localPath string, opts PutRemoteTextFileOptions) (*RemoteTextFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("path", localPath).Msg("putting remote text file")

	// Validate localPath is a directory
	info, err := os.Stat(localPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Errorf("checking directory: %w", err)
		}
		// Create directory if it doesn't exist
		if err := os.MkdirAll(localPath, 0755); err != nil {
			return nil, errors.Errorf("creating directory %s: %w", localPath, err)
		}
	} else if !info.IsDir() {
		return nil, errors.Errorf("path is not a directory: %s", localPath)
	}

	// Construct file path with .copy. suffix
	fileName := filepath.Base(file.Path())
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)
	destFileName := baseName + ".copy" + ext
	destPath := filepath.Join(localPath, destFileName)

	patchPath := strings.Replace(destPath, ".copy.", ".patch.", 1)

	if !strings.Contains(patchPath, ".patch.") && strings.HasSuffix(patchPath, ".copy") {
		patchPath = strings.Replace(patchPath, ".copy", ".patch", 1)
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

	// if !opts.IsIgnored {
	// Write content to file
	// if err := os.WriteFile(destPath, data, 0644); err != nil {
	// 	return nil, errors.Errorf("writing file content: %w", err)
	// }
	// }
	var patch *Patch
	// check if patch file exists
	if _, err := os.Stat(patchPath); err == nil {
		zerolog.Ctx(ctx).Debug().Str("path", patchPath).Msg("patch file exists")
		patch = &Patch{
			PatchPath: patchPath,
		}

		patchContent, err := os.ReadFile(patchPath)
		if err != nil {
			return nil, errors.Errorf("reading patch file: %w", err)
		}

		dmp := diffmatchpatch.New()

		diffs := dmp.DiffMain(string(data), string(patchContent), false)

		patch.PatchDiff = string(dmp.DiffToDelta(diffs))

		patchHash, err := s.hashFile(patchPath)
		if err != nil {
			return nil, errors.Errorf("hashing patch file: %w", err)
		}
		patch.PatchHash = patchHash

		// delete patch file
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
		LocalPath:   destPath,
		LastUpdated: time.Now(),
		IsPatched:   patch != nil, // we check this on write
		Patch:       patch,
		ContentHash: hash,
		Permalink:   file.WebViewPermalink(),
		IsIgnored:   opts.IsIgnored,
		content:     data,
	}

	// Update state
	s.file.RemoteTextFiles = append(s.file.RemoteTextFiles, *remoteFile)

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
	// Create replacer
	replacer := text.NewSimpleTextReplacer()

	// Convert config rule to text rule
	rule := text.ReplacementRule{
		FromText:       mod.FromText,
		ToText:         mod.ToText,
		FileFilterGlob: mod.FileFilterGlob,
	}

	// Get raw content
	content, err := f.RawRemoteContent()
	if err != nil {
		return errors.Errorf("getting raw content: %w", err)
	}
	defer content.Close()

	// Apply replacement
	result, err := replacer.ReplaceText(ctx, content, []text.ReplacementRule{rule})
	if err != nil {
		return errors.Errorf("applying text replacement: %w", err)
	}

	// Create patch if needed
	if result.WasModified && !f.IsPatched {
		// Compress original content
		var gzipBuf bytes.Buffer
		gzipWriter := gzip.NewWriter(&gzipBuf)
		if _, err := gzipWriter.Write(result.OriginalContent); err != nil {
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

		// Create patch file
		if err := os.WriteFile(patchPath, result.OriginalContent, 0644); err != nil {
			return errors.Errorf("writing patch file: %w", err)
		}

		// Update file state
		f.IsPatched = true
		f.Patch = &Patch{
			PatchPath:     patchPath,
			RemoteContent: base64.StdEncoding.EncodeToString(gzipBuf.Bytes()),
		}
	}

	// Write modified content
	if err := os.WriteFile(f.LocalPath, result.ModifiedContent, 0644); err != nil {
		return errors.Errorf("writing modified content: %w", err)
	}

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
			if !strings.HasSuffix(f.LocalPath, ".copy") && !strings.HasSuffix(f.LocalPath, ".patch") {
				return errors.Errorf("invalid file suffix: %s (must contain .copy. or .patch.)", f.LocalPath)
			}
		}

		if f.IsIgnored || f.IsPatched {
			// Check file exists
			if _, err := os.Stat(f.LocalPath); !os.IsNotExist(err) {
				return errors.Errorf("file exists: %s", f.LocalPath)
			}
		} else {
			// Check file exists
			if _, err := os.Stat(f.LocalPath); os.IsNotExist(err) {
				return errors.Errorf("file does not exist: %s", f.LocalPath)
			}

			hash, err := s.hashFile(f.LocalPath)
			if err != nil {
				return errors.Errorf("computing hash for %s: %w", f.LocalPath, err)
			}
			if hash != f.ContentHash {
				return errors.Errorf("content hash mismatch for %s: expected %s, got %s", f.LocalPath, f.ContentHash, hash)
			}
		}

		if f.IsPatched {
			if _, err := os.Stat(f.Patch.PatchPath); os.IsNotExist(err) {
				return errors.Errorf("patch file does not exist: %s", f.Patch.PatchPath)
			}
		}

		// Check content hash

		// Check patch file if patched

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
		if file.IsIgnored {
			continue
		}
		if file.IsPatched {
			err := s.validateFile(ctx, file.Patch.PatchPath, file.Patch.PatchHash)
			if err != nil {
				s.logger.LogValidation(false, "State inconsistency detected", err)
				return false, nil
			}
		} else {
			if err := s.validateFile(ctx, file.LocalPath, file.ContentHash); err != nil {
				s.logger.LogValidation(false, "State inconsistency detected", err)
				return false, nil
			}
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

	// delete all ignored files if they exist or not
	for _, file := range s.file.RemoteTextFiles {
		if file.IsIgnored {
			if _, err := os.Stat(file.LocalPath); err == nil {
				if err := os.Remove(file.LocalPath); err != nil {
					return errors.Errorf("removing ignored file: %s: %w", file.LocalPath, err)
				}
			}
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

// 🔄 WriteAllFiles writes all files to disk
func (s *State) WriteAllFiles(ctx context.Context) error {
	for _, file := range s.file.RemoteTextFiles {
		if file.IsIgnored || file.IsPatched {
			if s.file.decompressedRawGzipData == nil {
				s.file.decompressedRawGzipData = make(map[string][]byte)
			}
			s.file.decompressedRawGzipData[file.LocalPath] = file.content
			continue
		}
		if err := os.WriteFile(file.LocalPath, file.content, 0644); err != nil {
			return errors.Errorf("writing file: %w", err)
		}
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

// ConfigHash returns a hash of the current configuration
func (s *State) ConfigHash() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// If no config is stored, return empty hash
	if s.file.Config == nil {
		return ""
	}

	return s.file.Config.GetHash()

	// // Marshal config to JSON for consistent hashing
	// data, err := json.Marshal(s.file.Config)
	// if err != nil {
	// 	// Log error but return empty hash
	// 	zerolog.Ctx(context.Background()).Error().Err(err).Msg("failed to marshal config for hashing")
	// 	return ""
	// }

	// // Compute SHA-256 hash
	// h := sha256.New()
	// h.Write(data)
	// return fmt.Sprintf("%x", h.Sum(nil))
}

// Reset resets the state to empty
func (s *State) Reset(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.file = &StateFile{
		SchemaVersion: CurrentSchemaVersion,
		LastUpdated:   time.Now(),
	}

	s.logger.LogStateChange("Reset state")
	return nil
}
