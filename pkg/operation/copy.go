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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/walteh/copyrc/pkg/status"
	"gitlab.com/tozd/go/errors"
)

// 📦 NewCopyOperation creates a new copy operation
func NewCopyOperation(opts Options) Operation {
	return &copyOperation{
		BaseOperation: NewBaseOperation(opts),
	}
}

// 📦 copyOperation implements the copy operation
type copyOperation struct {
	BaseOperation
}

// 🏃 Execute runs the copy operation
func (op *copyOperation) Execute(ctx context.Context) error {
	// Get commit hash
	commitHash, err := op.Provider.GetCommitHash(ctx, op.Config.Provider)
	if err != nil {
		return errors.Errorf("getting commit hash: %w", err)
	}

	// Get list of files
	files, err := op.Provider.ListFiles(ctx, op.Config.Provider)
	if err != nil {
		return errors.Errorf("listing files: %w", err)
	}

	// Start tracking progress
	op.StatusMgr.StartOperation(ctx, len(files))
	defer op.StatusMgr.FinishOperation(ctx)

	// Process each file
	for i, file := range files {
		if err := op.processFile(ctx, file); err != nil {
			return errors.Errorf("processing file %s: %w", file, err)
		}
		op.StatusMgr.UpdateProgress(ctx, i+1)
	}

	// Update lock file
	if err := op.StatusMgr.UpdateLockFile(ctx, commitHash, op.Config); err != nil {
		return errors.Errorf("updating lock file: %w", err)
	}

	return nil
}

// 📄 processFile processes a single file
func (op *copyOperation) processFile(ctx context.Context, file string) error {
	// Get commit hash for permalink
	commitHash, err := op.Provider.GetCommitHash(ctx, op.Config.Provider)
	if err != nil {
		return errors.Errorf("getting commit hash: %w", err)
	}

	// Check if file should be ignored
	if op.shouldIgnore(file) {
		op.StatusMgr.UpdateStatus(ctx, file, status.StatusUnchanged, &status.FileEntry{
			Status: status.StatusUnchanged,
			Metadata: map[string]string{
				"reason": "ignored by pattern",
			},
		})
		return nil
	}

	// Get file content
	content, permalink, err := op.getFileContent(ctx, file, commitHash)
	if err != nil {
		return errors.Errorf("getting file content: %w", err)
	}

	// Apply replacements
	if op.Config.Copy != nil && len(op.Config.Copy.Replacements) > 0 {
		content = op.applyReplacements(content, file)
	}

	// Add file header
	content = op.addFileHeader(file, content)

	// Check if file exists and get current status
	fileStatus := status.StatusNew
	currentContent, err := op.StatusMgr.ReadFile(ctx, file)
	if err == nil {
		// File exists, check if content differs
		if bytes.Equal(currentContent, content) {
			fileStatus = status.StatusUnchanged
		} else {
			fileStatus = status.StatusModified
		}
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(filepath.Join(op.Config.Destination, file)), 0755); err != nil {
		return errors.Errorf("creating parent directories: %w", err)
	}

	// Write file atomically using status manager
	if err := op.StatusMgr.WriteFileAtomic(ctx, file, content); err != nil {
		return errors.Errorf("writing file: %w", err)
	}

	// Update status with metadata
	op.StatusMgr.UpdateStatus(ctx, file, fileStatus, &status.FileEntry{
		Status: fileStatus,
		Metadata: map[string]string{
			"commit_hash": commitHash,
			"permalink":   permalink,
			"size":        fmt.Sprintf("%d", len(content)),
			"mode":        "0644",
		},
	})

	return nil
}

// 📥 getFileContent gets the content of a file from the provider
func (op *copyOperation) getFileContent(ctx context.Context, file, commitHash string) ([]byte, string, error) {
	// Get file permalink
	permalink, err := op.Provider.GetPermalink(ctx, op.Config.Provider, commitHash, file)
	if err != nil {
		return nil, "", errors.Errorf("getting permalink: %w", err)
	}

	// Get file content from provider
	reader, err := op.Provider.GetFile(ctx, op.Config.Provider, file)
	if err != nil {
		return nil, "", errors.Errorf("getting file: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", errors.Errorf("reading file: %w", err)
	}

	return content, permalink, nil
}

// 🔄 applyReplacements applies text replacements to file content
func (op *copyOperation) applyReplacements(content []byte, file string) []byte {
	str := string(content)
	for _, r := range op.Config.Copy.Replacements {
		// Skip if this replacement is for a specific file and it's not this file
		if r.File != nil && *r.File != file {
			continue
		}
		str = strings.ReplaceAll(str, r.Old, r.New)
	}
	return []byte(str)
}

// 🔍 shouldIgnore checks if a file should be ignored
func (op *copyOperation) shouldIgnore(path string) bool {
	if op.Config.Copy == nil || len(op.Config.Copy.IgnorePatterns) == 0 {
		return false
	}

	for _, pattern := range op.Config.Copy.IgnorePatterns {
		matched, err := doublestar.Match(pattern, path)
		if err != nil {
			op.Logger.Debug().Str("pattern", pattern).Str("path", path).Err(err).Msg("error matching pattern")
			continue
		}
		if matched {
			op.Logger.Debug().Str("file", path).Str("pattern", pattern).Msg("file ignored by pattern")
			return true
		}
	}

	return false
}

// 📝 addFileHeader adds a header to the file content
func (op *copyOperation) addFileHeader(file string, content []byte) []byte {
	var buf bytes.Buffer

	// Get file extension
	ext := filepath.Ext(file)

	// Add standard header based on file type
	switch ext {
	case ".go", ".js", ".ts", ".jsx", ".tsx", ".css":
		fmt.Fprintf(&buf, "// 📦 Generated by copyrc. DO NOT EDIT.\n")
		fmt.Fprintf(&buf, "// 🔗 Source: %s\n", op.Config.Provider.Repo)
		fmt.Fprintf(&buf, "// ℹ️ See .copyrc.lock for more details.\n\n")
	case ".py", ".rb", ".pl", ".sh", ".yaml", ".yml":
		fmt.Fprintf(&buf, "# 📦 Generated by copyrc. DO NOT EDIT.\n")
		fmt.Fprintf(&buf, "# 🔗 Source: %s\n", op.Config.Provider.Repo)
		fmt.Fprintf(&buf, "# ℹ️ See .copyrc.lock for more details.\n\n")
	case ".md", ".xml", ".html":
		fmt.Fprintf(&buf, "<!--\n")
		fmt.Fprintf(&buf, "📦 Generated by copyrc. DO NOT EDIT.\n")
		fmt.Fprintf(&buf, "🔗 Source: %s\n", op.Config.Provider.Repo)
		fmt.Fprintf(&buf, "ℹ️ See .copyrc.lock for more details.\n")
		fmt.Fprintf(&buf, "-->\n\n")
	}

	// Add original content
	buf.Write(content)
	return buf.Bytes()
}
