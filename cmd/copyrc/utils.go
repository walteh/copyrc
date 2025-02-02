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

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
	"gitlab.com/tozd/go/errors"
)

// WriteFileOpts represents options for writing a file
type WriteFileOpts struct {
	Destination Destination // Destination path
	// Required fields
	Path       string // Path to write the file to
	SourcePath string // Source path to write the file to
	Contents   []byte // Contents to write to the file
	// FileType FileType // Type of file (managed/local/copy)

	// Optional fields
	StatusFile       *StatusFile // Full status file for checking existing entries
	StatusMutex      *sync.Mutex // Mutex for status file access
	ReplacementCount int         // Number of replacements made in the file
	EnsureNewline    bool        // Ensure contents end with a newline
	RepoSourceInfo   string      // Source info for status entry
	Permalink        string      // Permalink for status entry
	Changes          []string    // Changes made to the file
	IsStatusFile     bool        // Whether this is a status file
	IsUntracked      bool        // Whether this is an untracked file
	IsManaged        bool        // Whether this is a managed file
}

// writeFile handles all file writing scenarios including status updates and logging.
// It returns true if the file was written, false if no changes were needed.
func writeFile(ctx context.Context, opts WriteFileOpts) (bool, error) {
	// Validate required fields
	if opts.Path == "" {
		return false, errors.New("path is required")
	}

	fileName := strings.TrimPrefix(opts.Path, opts.Destination.Path+"/")

	if opts.IsUntracked {
		logFileOperation(ctx, FileInfo{
			Name:        opts.Path,
			IsUntracked: true,
		})

		return false, nil
	}

	if opts.StatusFile == nil && !opts.IsUntracked {
		return false, errors.New("status file is required")
	}

	if opts.IsStatusFile {
		opts.IsManaged = true
	}

	isCustomized := false
	// if the file contains the string "copyrc:customized"

	// Get base name for status entries

	// Try to read existing file
	existing, err := os.ReadFile(opts.Path)
	if err != nil && !os.IsNotExist(err) {
		return false, errors.Errorf("reading file: %w", err)
	}

	existingHashData := sha256.New()
	existingHashData.Write(existing)
	existingHash := base64.URLEncoding.EncodeToString(existingHashData.Sum(nil))

	if len(existing) == 0 {
		// try the path without the .copy. in the name
		existing, err = os.ReadFile(opts.SourcePath)
		if err != nil && !os.IsNotExist(err) {
			return false, errors.Errorf("reading file: %w", err)
		}
	}
	var rcount = 0
	var hasEntry bool = false
	var remoteHash string
	var customizations string = ""
	if opts.StatusFile != nil {
		if opts.IsManaged {
			_, hasEntryd := opts.StatusFile.GeneratedFiles[fileName]
			hasEntry = hasEntryd
		} else {
			entry, hasEntryd := opts.StatusFile.CoppiedFiles[fileName]
			hasEntry = hasEntryd
			if hasEntry {
				remoteHash = entry.RemoteHash
				customizations = entry.DiffDelta
				rcount = len(entry.Changes)
			}
		}
	}

	var encodedCustomizations string
	if (remoteHash != "" && existingHash != remoteHash) || customizations != "" {
		isCustomized = true
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(existing), string(opts.Contents), false)
		encodedCustomizations = dmp.DiffToDelta(diffs)
		rcount = len(diffs)
	}

	if hasEntry && len(opts.Contents) == 0 {
		logFileOperation(ctx, FileInfo{
			Name:         fileName,
			IsCustomized: isCustomized,
			IsManaged:    opts.IsManaged,
			Replacements: rcount,
		})
		return false, nil
	}

	if len(opts.Contents) == 0 {
		return false, errors.Errorf("contents are required for %s", opts.Path)
	}

	// Ensure newline at end of file if requested
	contents := opts.Contents
	if opts.EnsureNewline && !bytes.HasSuffix(contents, []byte("\n")) {
		contents = append(contents, '\n')
	}

	logger := loggerFromContext(ctx)
	logger.zlog.Debug().Msgf("👀 Writing file %s with contents length %d, curr len: %d, equal: %t", opts.Path, len(contents), len(existing), bytes.Equal(existing, contents))

	// If file exists and content is the same, and we have an existing status entry, no need to write
	if err == nil && bytes.Equal(existing, contents) && (hasEntry || opts.IsStatusFile) {
		// Log the unchanged status
		logFileOperation(ctx, FileInfo{
			Name:         fileName,
			IsModified:   false,
			IsManaged:    opts.IsManaged,
			Replacements: opts.ReplacementCount,
		})

		return false, nil
	}

	if opts.IsStatusFile {
		opts.StatusFile.LastUpdated = time.Now()

		// Marshal status data
		data, err := json.MarshalIndent(opts.StatusFile, "", "\t")
		if err != nil {
			return false, errors.Errorf("marshaling status: %w", err)
		}
		opts.Contents = data
	}

	if !isCustomized {
		// Ensure the directory exists
		if err := os.MkdirAll(filepath.Dir(opts.Path), 0755); err != nil {
			return false, errors.Errorf("creating directory for %s: %w", opts.Path, err)
		}

		if err := os.WriteFile(opts.Path, contents, 0644); err != nil {
			return false, errors.Errorf("writing file %s: %w", opts.Path, err)
		}
	}

	hashData := sha256.New()
	hashData.Write(contents)
	hash := base64.URLEncoding.EncodeToString(hashData.Sum(nil))

	// Update status entries
	now := time.Now().UTC()
	if opts.StatusFile != nil && opts.StatusMutex != nil {
		opts.StatusMutex.Lock()
		if opts.IsManaged {
			entry, ok := opts.StatusFile.GeneratedFiles[fileName]
			if !ok {
				entry = GeneratedFileEntry{
					File: fileName,
				}
			}
			entry.LastUpdated = now
			opts.StatusFile.GeneratedFiles[fileName] = entry
		} else {
			entry, ok := opts.StatusFile.CoppiedFiles[fileName]
			if !ok {
				entry = StatusEntry{
					File: fileName,
				}
			}
			entry.LastUpdated = now
			entry.Source = opts.RepoSourceInfo
			entry.Permalink = opts.Permalink
			entry.Changes = opts.Changes
			entry.DiffDelta = encodedCustomizations
			entry.RemoteHash = hash
			opts.StatusFile.CoppiedFiles[fileName] = entry
		}
		opts.StatusMutex.Unlock()
	}

	// Log the operation based on what changed
	logFileOperation(ctx, FileInfo{
		Name:         fileName,
		IsNew:        len(existing) == 0 && len(contents) > 0,
		IsModified:   true,
		IsCustomized: isCustomized,
		IsManaged:    opts.IsManaged,
		Replacements: opts.ReplacementCount,
	})

	return true, nil
}
