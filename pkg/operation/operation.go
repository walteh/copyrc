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
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/provider"
	"github.com/walteh/copyrc/pkg/status"
	"gitlab.com/tozd/go/errors"
)

// 🎯 Manager handles file operations
type Manager struct {
	cfg      *config.Config
	provider provider.Provider
	logger   *zerolog.Logger
}

// 🏭 New creates a new operation manager
func New(cfg *config.Config, provider provider.Provider, logger *zerolog.Logger) *Manager {
	return &Manager{
		cfg:      cfg,
		provider: provider,
		logger:   logger,
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
	Type         string // File type
	Status       string // File status
}

// 📂 ProcessFile processes a single file
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
	m.logFileOperation(info)

	// Get file content
	rc, err := m.provider.GetFile(ctx, m.cfg.Provider, path)
	if err != nil {
		return errors.Errorf("getting file: %w", err)
	}
	defer rc.Close()

	// Create destination directory
	destPath := filepath.Join(m.cfg.Destination, path)
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.Errorf("creating destination directory: %w", err)
	}

	// Process content
	if err := m.processFileContent(rc, destPath); err != nil {
		return errors.Errorf("processing file content: %w", err)
	}

	return nil
}

// 📂 ProcessFiles processes multiple files
func (m *Manager) ProcessFiles(ctx context.Context, files []string) error {
	if m.cfg.Async {
		var wg sync.WaitGroup
		errCh := make(chan error, len(files))

		for _, file := range files {
			wg.Add(1)
			go func(f string) {
				defer wg.Done()
				if err := m.ProcessFile(ctx, f); err != nil {
					errCh <- errors.Errorf("processing file %s: %w", f, err)
				}
			}(file)
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			if err != nil {
				return err
			}
		}
	} else {
		for _, file := range files {
			if err := m.ProcessFile(ctx, file); err != nil {
				return errors.Errorf("processing file %s: %w", file, err)
			}
		}
	}

	return nil
}

// 🔍 shouldIgnore checks if a file should be ignored
func (m *Manager) shouldIgnore(path string) bool {
	if m.cfg.Copy == nil || len(m.cfg.Copy.IgnorePatterns) == 0 {
		return false
	}

	for _, pattern := range m.cfg.Copy.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}

	return false
}

// 📄 getFileInfo gets information about a file
func (m *Manager) getFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	info := &FileInfo{
		Path: path,
	}

	// Get file type
	info.Type = m.getFileType(path)

	// Get file status
	status, err := m.getFileStatus(ctx, path)
	if err != nil {
		return nil, errors.Errorf("getting file status: %w", err)
	}
	info.Status = status

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

	// Check if file exists
	destContent, err := os.ReadFile(filepath.Join(m.cfg.Destination, path))
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Errorf("reading destination file: %w", err)
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

// 📝 logFileOperation logs a file operation
func (m *Manager) logFileOperation(info *FileInfo) {
	// Format the operation string
	formatted := status.FormatFileOperation(
		info.Path,
		info.Type,
		info.Status,
		info.IsNew,
		info.IsModified,
		info.IsRemoved,
	)

	// Log with structured data
	m.logger.Info().
		Str("file", info.Path).
		Str("type", info.Type).
		Str("status", info.Status).
		Bool("is_new", info.IsNew).
		Bool("is_modified", info.IsModified).
		Bool("is_removed", info.IsRemoved).
		Int("replacements", info.Replacements).
		Msg(formatted)
}

// 🔍 getFileType determines the type of a file
func (m *Manager) getFileType(path string) string {
	if m.cfg.GoEmbed {
		return "embed"
	}
	if strings.HasSuffix(path, ".gen.go") {
		return "generated"
	}
	if m.cfg.Copy != nil && len(m.cfg.Copy.Replacements) > 0 {
		return "managed"
	}
	return "copy"
}

// 🔍 getFileStatus determines the status of a file
func (m *Manager) getFileStatus(ctx context.Context, path string) (string, error) {
	// TODO: Implement file status checking
	return "NEW", nil
}

// 📄 processFileContent processes the content of a file
func (m *Manager) processFileContent(rc io.ReadCloser, destPath string) error {
	// Read content
	content, err := io.ReadAll(rc)
	if err != nil {
		return errors.Errorf("reading content: %w", err)
	}

	// Apply replacements
	if m.cfg.Copy != nil {
		for _, r := range m.cfg.Copy.Replacements {
			// Skip if this replacement is for a specific file and it's not this one
			if r.File != nil && *r.File != filepath.Base(destPath) {
				continue
			}
			content = bytes.ReplaceAll(content, []byte(r.Old), []byte(r.New))
		}
	}

	// Write content
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return errors.Errorf("writing file: %w", err)
	}

	return nil
}
