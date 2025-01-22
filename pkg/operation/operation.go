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

package operation

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/log"
	"github.com/walteh/copyrc/pkg/provider"
	"gitlab.com/tozd/go/errors"
)

// 🎯 Manager handles file operations
type Manager struct {
	cfg      *config.Config
	provider provider.Provider
	logger   *log.Logger
}

// 🏭 New creates a new operation manager
func New(cfg *config.Config, p provider.Provider, l *log.Logger) *Manager {
	return &Manager{
		cfg:      cfg,
		provider: p,
		logger:   l,
	}
}

// 📝 FileInfo represents information about a file
type FileInfo struct {
	Path         string // File path
	IsNew        bool   // Whether this is a new file
	IsModified   bool   // Whether the file was modified
	IsRemoved    bool   // Whether the file was removed
	IsManaged    bool   // Whether this is a managed file
	Replacements int    // Number of replacements made
}

// 🔍 ProcessFile processes a single file
func (m *Manager) ProcessFile(ctx context.Context, path string) error {
	// Check if file should be ignored
	if m.shouldIgnore(path) {
		return nil
	}

	// Get file info
	info, err := m.getFileInfo(ctx, path)
	if err != nil {
		return errors.Errorf("getting file info: %w", err)
	}

	// Log operation
	m.logger.LogFileOperation(ctx, log.FileOperation{
		Path:         path,
		Type:         m.getFileType(info),
		Status:       m.getFileStatus(info),
		IsNew:        info.IsNew,
		IsModified:   info.IsModified,
		IsRemoved:    info.IsRemoved,
		IsManaged:    info.IsManaged,
		Replacements: info.Replacements,
	})

	// Process file
	if err := m.processFileContent(ctx, path, info); err != nil {
		return errors.Errorf("processing file content: %w", err)
	}

	return nil
}

// 🔍 shouldIgnore checks if a file should be ignored
func (m *Manager) shouldIgnore(path string) bool {
	if m.cfg.Copy == nil || len(m.cfg.Copy.IgnoreFiles) == 0 {
		return false
	}

	for _, pattern := range m.cfg.Copy.IgnoreFiles {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}

	return false
}

// 🔍 getFileInfo gets information about a file
func (m *Manager) getFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	// Get file from provider
	rc, err := m.provider.GetFile(ctx, m.cfg.Provider, path)
	if err != nil {
		return nil, errors.Errorf("getting file from provider: %w", err)
	}
	defer rc.Close()

	// Read file content
	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, errors.Errorf("reading file content: %w", err)
	}

	// Get destination path
	destPath := filepath.Join(m.cfg.Destination, path)

	// Check if file exists
	destContent, err := os.ReadFile(destPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Errorf("reading destination file: %w", err)
	}

	// Create file info
	info := &FileInfo{
		Path:      path,
		IsManaged: strings.HasSuffix(path, ".copyrc.lock"),
	}

	// Determine file status
	if os.IsNotExist(err) {
		info.IsNew = true
	} else {
		info.IsModified = string(content) != string(destContent)
	}

	// Count replacements if needed
	if m.cfg.Copy != nil && len(m.cfg.Copy.Replacements) > 0 {
		for _, r := range m.cfg.Copy.Replacements {
			if r.File != nil && *r.File != path {
				continue
			}
			info.Replacements += strings.Count(string(content), r.Old)
		}
	}

	return info, nil
}

// 🔍 getFileType gets the type of a file
func (m *Manager) getFileType(info *FileInfo) string {
	if info.IsManaged {
		return "managed"
	}
	if info.Replacements > 0 {
		return "copy"
	}
	return "local"
}

// 🔍 getFileStatus gets the status of a file
func (m *Manager) getFileStatus(info *FileInfo) string {
	switch {
	case info.IsRemoved:
		return "REMOVED"
	case info.IsNew:
		return "NEW"
	case info.IsModified:
		return "UPDATED"
	default:
		return "no change"
	}
}

// 🔍 processFileContent processes the content of a file
func (m *Manager) processFileContent(ctx context.Context, path string, info *FileInfo) error {
	// Get file from provider
	rc, err := m.provider.GetFile(ctx, m.cfg.Provider, path)
	if err != nil {
		return errors.Errorf("getting file from provider: %w", err)
	}
	defer rc.Close()

	// Read file content
	content, err := io.ReadAll(rc)
	if err != nil {
		return errors.Errorf("reading file content: %w", err)
	}

	// Apply replacements
	if m.cfg.Copy != nil && len(m.cfg.Copy.Replacements) > 0 {
		for _, r := range m.cfg.Copy.Replacements {
			if r.File != nil && *r.File != path {
				continue
			}
			content = []byte(strings.ReplaceAll(string(content), r.Old, r.New))
		}
	}

	// Get destination path
	destPath := filepath.Join(m.cfg.Destination, path)

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return errors.Errorf("creating destination directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return errors.Errorf("writing file: %w", err)
	}

	return nil
}

// 🚀 ProcessFiles processes multiple files
func (m *Manager) ProcessFiles(ctx context.Context, paths []string) error {
	if m.cfg.Async {
		return m.processFilesAsync(ctx, paths)
	}
	return m.processFilesSync(ctx, paths)
}

// 🔄 processFilesSync processes files synchronously
func (m *Manager) processFilesSync(ctx context.Context, paths []string) error {
	for _, path := range paths {
		if err := m.ProcessFile(ctx, path); err != nil {
			return errors.Errorf("processing file %s: %w", path, err)
		}
	}
	return nil
}

// ⚡️ processFilesAsync processes files asynchronously
func (m *Manager) processFilesAsync(ctx context.Context, paths []string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(paths))

	for _, path := range paths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			if err := m.ProcessFile(ctx, p); err != nil {
				errCh <- errors.Errorf("processing file %s: %w", p, err)
			}
		}(path)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errCh)

	// Check for errors
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
