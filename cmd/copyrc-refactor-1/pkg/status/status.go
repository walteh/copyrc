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

package status

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/walteh/copyrc/cmd/copyrc-refactor-1/pkg/config"
	"gitlab.com/tozd/go/errors"
)

// 📊 FileStatus represents the current state of a file
type FileStatus int

const (
	StatusUnknown   FileStatus = iota
	StatusNew                  // File doesn't exist in destination
	StatusModified             // File exists but content differs
	StatusUnchanged            // File exists and content matches
	StatusDeleted              // File was deleted
	StatusLocal                // File is local only
	StatusManaged              // File is managed by copyrc
)

// String returns a string representation of FileStatus
func (s FileStatus) String() string {
	switch s {
	case StatusNew:
		return "new"
	case StatusModified:
		return "modified"
	case StatusUnchanged:
		return "unchanged"
	case StatusDeleted:
		return "deleted"
	case StatusLocal:
		return "local"
	case StatusManaged:
		return "managed"
	default:
		return "unknown"
	}
}

// 📄 FileInfo contains metadata about a file
type FileInfo struct {
	Path     string      // Relative path to the file
	Status   FileStatus  // Current status
	Size     int64       // File size in bytes
	Mode     os.FileMode // File permissions
	IsDir    bool        // Whether this is a directory
	Checksum string      // Content hash for diff detection
	Error    error       // Any error associated with this file
}

// 💾 FileManager handles all file system operations
type FileManager interface {
	// Core operations
	WriteFile(ctx context.Context, path string, content []byte) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
	DeleteFile(ctx context.Context, path string) error
	FileExists(ctx context.Context, path string) (bool, error)

	// Directory operations
	CreateDir(ctx context.Context, path string) error
	RemoveDir(ctx context.Context, path string) error

	// Atomic operations
	WriteFileAtomic(ctx context.Context, path string, content []byte) error

	// Backup operations
	BackupFile(ctx context.Context, path string) error
	RestoreFile(ctx context.Context, path string) error
}

// 📈 StatusReporter tracks file status and reports progress
type StatusReporter interface {
	// Status tracking
	TrackFile(ctx context.Context, path string, info FileInfo)
	GetFileInfo(ctx context.Context, path string) (FileInfo, error)
	ListFiles(ctx context.Context) ([]FileInfo, error)

	// Progress reporting
	StartOperation(ctx context.Context, total int)
	UpdateProgress(ctx context.Context, processed int)
	FinishOperation(ctx context.Context)
}

// 📊 Manager handles file operations and status tracking
type Manager struct {
	baseDir   string
	formatter FileFormatter
	progress  ProgressIndicator
	files     map[string]*FileEntry
	mu        sync.RWMutex
	total     int
	processed int
}

// 📄 FileEntry tracks the status and metadata of a file
type FileEntry struct {
	Status   FileStatus
	Metadata map[string]string
}

// 📝 LockFile represents the status lock file
type LockFile struct {
	LastUpdated    time.Time                 `json:"last_updated"`
	CommitHash     string                    `json:"commit_hash"`
	Config         *config.Config            `json:"config"`
	CopiedFiles    map[string]CopiedFileInfo `json:"copied_files"`
	GeneratedFiles map[string]GeneratedFile  `json:"generated_files"`
}

// 📄 CopiedFileInfo represents a copied file's status
type CopiedFileInfo struct {
	Path      string    `json:"path"`
	Hash      string    `json:"hash"`
	UpdatedAt time.Time `json:"updated_at"`
	Source    string    `json:"source"`
	Permalink string    `json:"permalink"`
}

// 📄 GeneratedFile represents a generated file's status
type GeneratedFile struct {
	Path      string    `json:"path"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 🆕 NewManager creates a new Manager instance
func NewManager(baseDir string, formatter FileFormatter) *Manager {
	return &Manager{
		baseDir:   filepath.Clean(baseDir),
		formatter: formatter,
		progress:  NewDefaultProgressIndicator(),
		files:     make(map[string]*FileEntry),
	}
}

// 🔄 UpdateStatus updates the status of a file
func (m *Manager) UpdateStatus(ctx context.Context, path string, status FileStatus, entry *FileEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.files[path] = entry
	msg := m.formatter.FormatFileStatus(path, status, entry.Metadata)
	fmt.Println(msg)
}

// 🔒 getAbsPath returns the absolute path for a given relative path
func (m *Manager) getAbsPath(path string) string {
	return filepath.Join(m.baseDir, path)
}

// 🔍 calculateChecksum generates a SHA-256 hash of the content
func calculateChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// FileManager interface implementation

func (m *Manager) WriteFile(ctx context.Context, path string, content []byte) error {
	absPath := m.getAbsPath(path)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return errors.Errorf("creating parent directories: %w", err)
	}

	// Write file atomically
	return m.WriteFileAtomic(ctx, path, content)
}

func (m *Manager) WriteFileAtomic(ctx context.Context, path string, content []byte) error {
	absPath := m.getAbsPath(path)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return errors.Errorf("creating parent directories: %w", err)
	}

	// Create temporary file
	tmpPath := absPath + ".tmp"
	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return errors.Errorf("writing temporary file: %w", err)
	}

	// Rename temporary file to target
	if err := os.Rename(tmpPath, absPath); err != nil {
		os.Remove(tmpPath) // Clean up temp file
		return errors.Errorf("renaming temporary file: %w", err)
	}

	return nil
}

func (m *Manager) ReadFile(ctx context.Context, path string) ([]byte, error) {
	content, err := os.ReadFile(m.getAbsPath(path))
	if err != nil {
		return nil, errors.Errorf("reading file: %w", err)
	}
	return content, nil
}

func (m *Manager) DeleteFile(ctx context.Context, path string) error {
	if err := os.Remove(m.getAbsPath(path)); err != nil {
		return errors.Errorf("deleting file: %w", err)
	}
	return nil
}

func (m *Manager) FileExists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "checking file existence")
	}
	return true, nil
}

func (m *Manager) CreateDir(ctx context.Context, path string) error {
	if err := os.MkdirAll(m.getAbsPath(path), 0755); err != nil {
		return errors.Errorf("creating directory: %w", err)
	}
	return nil
}

func (m *Manager) RemoveDir(ctx context.Context, path string) error {
	if err := os.RemoveAll(m.getAbsPath(path)); err != nil {
		return errors.Errorf("removing directory: %w", err)
	}
	return nil
}

func (m *Manager) BackupFile(ctx context.Context, path string) error {
	absPath := m.getAbsPath(path)
	backupPath := absPath + ".bak"

	// Only backup if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.Errorf("checking file existence: %w", err)
	}

	// Copy file to backup
	if err := copyFile(absPath, backupPath); err != nil {
		return errors.Errorf("creating backup: %w", err)
	}

	return nil
}

func (m *Manager) RestoreFile(ctx context.Context, path string) error {
	absPath := m.getAbsPath(path)
	backupPath := absPath + ".bak"

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return errors.Errorf("backup file does not exist")
	} else if err != nil {
		return errors.Errorf("checking backup existence: %w", err)
	}

	// Restore from backup
	if err := copyFile(backupPath, absPath); err != nil {
		return errors.Errorf("restoring from backup: %w", err)
	}

	// Remove backup
	if err := os.Remove(backupPath); err != nil {
		return errors.Errorf("removing backup: %w", err)
	}

	return nil
}

// StatusReporter interface implementation

func (m *Manager) TrackFile(ctx context.Context, path string, info FileInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &FileEntry{
		Status: info.Status,
		Metadata: map[string]string{
			"size": fmt.Sprintf("%d", info.Size),
			"mode": info.Mode.String(),
		},
	}
	if info.Error != nil {
		entry.Metadata["error"] = info.Error.Error()
	}
	m.files[path] = entry

	msg := m.formatter.FormatFileOperation(
		path,
		"file",
		"ok",
		info.Status == StatusNew,
		info.Status == StatusModified,
		info.Status == StatusDeleted,
	)
	if info.Error != nil {
		msg = m.formatter.FormatError(info.Error)
	}
	fmt.Println(msg)
}

func (m *Manager) GetFileInfo(ctx context.Context, path string) (FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.files[path]
	if !ok {
		return FileInfo{}, errors.Errorf("file not tracked: %s", path)
	}

	size := int64(0)
	if s, ok := entry.Metadata["size"]; ok {
		size, _ = strconv.ParseInt(s, 10, 64)
	}

	mode := os.FileMode(0644) // Default mode
	if m, ok := entry.Metadata["mode"]; ok {
		if parsed, err := strconv.ParseUint(m, 8, 32); err == nil {
			mode = os.FileMode(parsed)
		}
	}

	var err error
	if e, ok := entry.Metadata["error"]; ok {
		err = errors.New(e)
	}

	return FileInfo{
		Path:   path,
		Status: entry.Status,
		Size:   size,
		Mode:   mode,
		Error:  err,
	}, nil
}

func (m *Manager) ListFiles(ctx context.Context) ([]FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files := make([]FileInfo, 0, len(m.files))
	for path := range m.files {
		if info, err := m.GetFileInfo(ctx, path); err == nil {
			files = append(files, info)
		}
	}
	return files, nil
}

// StartOperation starts tracking progress for an operation
func (m *Manager) StartOperation(ctx context.Context, total int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.total = total
	m.processed = 0
	msg := m.formatter.FormatProgress(0, total)
	fmt.Println(msg)
}

// UpdateProgress updates the progress of the current operation
func (m *Manager) UpdateProgress(ctx context.Context, processed int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processed = processed
	msg := m.formatter.FormatProgress(processed, m.total)
	fmt.Println(msg)
}

// FinishOperation completes the current operation
func (m *Manager) FinishOperation(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg := m.formatter.FormatProgress(m.total, m.total)
	fmt.Println(msg)
}

// Helper functions

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return errors.Errorf("opening source file: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return errors.Errorf("creating destination file: %w", err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return errors.Errorf("copying file: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst, creating parent directories if needed
func (m *Manager) CopyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return errors.Errorf("creating parent directories: %w", err)
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return errors.Errorf("creating destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return errors.Errorf("copying file content: %w", err)
	}

	return nil
}

// 💾 loadLockFile loads the lock file from disk
func (m *Manager) loadLockFile(ctx context.Context) (*LockFile, error) {
	lockPath := filepath.Join(m.baseDir, ".copyrc.lock")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &LockFile{
				CopiedFiles:    make(map[string]CopiedFileInfo),
				GeneratedFiles: make(map[string]GeneratedFile),
			}, nil
		}
		return nil, errors.Errorf("reading lock file: %w", err)
	}

	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, errors.Errorf("parsing lock file: %w", err)
	}

	if lock.CopiedFiles == nil {
		lock.CopiedFiles = make(map[string]CopiedFileInfo)
	}
	if lock.GeneratedFiles == nil {
		lock.GeneratedFiles = make(map[string]GeneratedFile)
	}

	return &lock, nil
}

// 💾 saveLockFile saves the lock file to disk
func (m *Manager) saveLockFile(ctx context.Context, lock *LockFile) error {
	lockPath := filepath.Join(m.baseDir, ".copyrc.lock")

	// Update timestamp
	lock.LastUpdated = time.Now()

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return errors.Errorf("marshaling lock file: %w", err)
	}

	// Write atomically using temp file
	tempPath := lockPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return errors.Errorf("writing temp lock file: %w", err)
	}

	if err := os.Rename(tempPath, lockPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return errors.Errorf("renaming temp lock file: %w", err)
	}

	return nil
}

// 🔄 UpdateLockFile updates the lock file with the latest status
func (m *Manager) UpdateLockFile(ctx context.Context, commitHash string, cfg *config.Config) error {
	lock, err := m.loadLockFile(ctx)
	if err != nil {
		return errors.Errorf("loading lock file: %w", err)
	}

	// Update metadata
	lock.CommitHash = commitHash
	lock.Config = cfg

	// Update copied files
	for path, entry := range m.files {
		if entry.Status == StatusDeleted {
			delete(lock.CopiedFiles, path)
			continue
		}

		info, err := m.GetFileInfo(ctx, path)
		if err != nil {
			return errors.Errorf("getting file info: %w", err)
		}

		lock.CopiedFiles[path] = CopiedFileInfo{
			Path:      info.Path,
			Hash:      calculateChecksum([]byte(path)), // TODO: Use actual file content hash
			UpdatedAt: time.Now(),
			Source:    "github.com/org/repo", // TODO: Get from provider
			Permalink: "https://...",         // TODO: Get from provider
		}
	}

	return m.saveLockFile(ctx, lock)
}

// 🏷️ MarkFileIgnored marks a file as ignored with metadata
func (m *Manager) MarkFileIgnored(ctx context.Context, path string, reason string) {
	m.UpdateStatus(ctx, path, StatusUnchanged, &FileEntry{
		Status: StatusUnchanged,
		Metadata: map[string]string{
			"reason":  reason,
			"ignored": "true",
		},
	})
}

// 🔄 UpdateFileContent updates file content and status atomically
func (m *Manager) UpdateFileContent(ctx context.Context, path string, content []byte, metadata map[string]string) error {
	// Calculate content hash
	hash := calculateChecksum(content)

	// Write file atomically
	if err := m.WriteFileAtomic(ctx, path, content); err != nil {
		return errors.Errorf("writing file content: %w", err)
	}

	// Merge metadata
	allMetadata := map[string]string{
		"hash": hash,
		"size": fmt.Sprintf("%d", len(content)),
		"mode": "0644",
	}
	for k, v := range metadata {
		allMetadata[k] = v
	}

	// Update status
	m.UpdateStatus(ctx, path, StatusModified, &FileEntry{
		Status:   StatusModified,
		Metadata: allMetadata,
	})

	return nil
}
