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
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"
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

// 🔧 Manager implements both FileManager and StatusReporter
type Manager struct {
	baseDir   string          // Base directory for all operations
	logger    *zerolog.Logger // Logger for status updates
	formatter FileFormatter   // Formatter for status messages

	// Status tracking
	mu    sync.RWMutex
	files map[string]FileInfo

	// Progress tracking
	total     int
	processed int
}

// 🏭 New creates a new status manager
func New(baseDir string, logger *zerolog.Logger) *Manager {
	return &Manager{
		baseDir:   filepath.Clean(baseDir),
		logger:    logger,
		formatter: NewDefaultFileFormatter(),
		files:     make(map[string]FileInfo),
	}
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
	tempPath := absPath + ".tmp"

	// Write to temp file
	if err := os.WriteFile(tempPath, content, 0644); err != nil {
		return errors.Errorf("writing temp file: %w", err)
	}

	// Rename temp file to target (atomic operation)
	if err := os.Rename(tempPath, absPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return errors.Errorf("renaming temp file: %w", err)
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
	_, err := os.Stat(m.getAbsPath(path))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.Errorf("checking file existence: %w", err)
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

	m.files[path] = info
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
	m.logger.Info().Str("path", path).Msg(msg)
}

func (m *Manager) GetFileInfo(ctx context.Context, path string) (FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, ok := m.files[path]
	if !ok {
		return FileInfo{}, errors.Errorf("file not tracked: %s", path)
	}
	return info, nil
}

func (m *Manager) ListFiles(ctx context.Context) ([]FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files := make([]FileInfo, 0, len(m.files))
	for _, info := range m.files {
		files = append(files, info)
	}
	return files, nil
}

func (m *Manager) StartOperation(ctx context.Context, total int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.total = total
	m.processed = 0
	msg := m.formatter.FormatProgress(0, total)
	m.logger.Info().Int("total", total).Msg(msg)
}

func (m *Manager) UpdateProgress(ctx context.Context, processed int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processed = processed
	msg := m.formatter.FormatProgress(processed, m.total)
	m.logger.Info().
		Int("processed", processed).
		Int("total", m.total).
		Msg(msg)
}

func (m *Manager) FinishOperation(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg := m.formatter.FormatProgress(m.total, m.total)
	m.logger.Info().
		Int("processed", m.total).
		Int("total", m.total).
		Msg(msg)
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
