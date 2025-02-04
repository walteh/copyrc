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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gitlab.com/tozd/go/errors"
)

// 🧪 Mock provider for testing
type MockProvider struct {
	files      map[string][]byte
	commitHash string
	ref        string
	org        string
	repo       string
	path       string
	t          *testing.T
}

func NewMockProvider(t *testing.T) *MockProvider {
	return &MockProvider{
		files:      make(map[string][]byte), // Create a new map for each instance
		commitHash: "abc123",
		ref:        "main",
		org:        "org",
		repo:       "repo",
		path:       "path/to/files",
		t:          t,
	}
}

// Helper methods for testing
func (m *MockProvider) AddFile(name string, content []byte) {
	m.files[name] = content
}

func (m *MockProvider) ClearFiles() {
	m.files = make(map[string][]byte)
}

func (m *MockProvider) ListFiles(ctx context.Context, args Source, recursive bool) ([]ProviderFile, error) {
	// Return all files in the map
	files := make([]ProviderFile, 0, len(m.files))
	for f := range m.files {
		files = append(files, ProviderFile{
			Path: f,
		})
	}
	logger := loggerFromContext(ctx)
	logger.zlog.Debug().Msgf("🧪 Mock provider listing %d files: %v", len(files), files)
	return files, nil
}

func (m *MockProvider) GetCommitHash(ctx context.Context, args Source) (string, error) {
	return m.commitHash, nil
}

func (m *MockProvider) GetPermalink(ctx context.Context, args Source, commitHash string, file string) (string, error) {
	// Remove the path prefix if it exists
	cleanFile := strings.TrimPrefix(file, m.path+"/")
	if m.files[cleanFile] == nil {
		return "", errors.Errorf("file not found: %s", file)
	}

	// Return a mock permalink that includes the file path
	return "mock://" + cleanFile, nil
}

func (m *MockProvider) GetSourceInfo(ctx context.Context, args Source, commitHash string) (string, error) {
	return "mock@" + commitHash, nil
}

func (m *MockProvider) GetFullRepo() string {
	return "github.com/" + m.org + "/" + m.repo
}

// GetArchiveUrl returns a mock URL for testing
func (m *MockProvider) GetArchiveUrl(ctx context.Context, args Source) (string, error) {
	// Create a temporary file with the archive data
	data, err := m.GetArchiveData()
	if err != nil {
		return "", errors.Errorf("creating archive data: %w", err)
	}

	// Create a temporary file
	f, err := os.CreateTemp("", "mock-archive-*.tar.gz")
	if err != nil {
		return "", errors.Errorf("creating temp file: %w", err)
	}
	defer f.Close()

	// Write the archive data
	if _, err := f.Write(data); err != nil {
		return "", errors.Errorf("writing archive data: %w", err)
	}

	// Return a file:// URL to the temporary file
	return "file://" + f.Name(), nil
}

// GetArchiveData returns a mock archive for testing
func (m *MockProvider) GetArchiveData() ([]byte, error) {
	// Create a tar.gz archive in memory
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add each file to the archive
	for name, content := range m.files {
		// Create tar header
		header := &tar.Header{
			Name:    filepath.Join(m.path, name), // Include the path prefix
			Size:    int64(len(content)),
			Mode:    0644,
			ModTime: time.Now(),
		}

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return nil, errors.Errorf("writing tar header: %w", err)
		}

		// Write content
		if _, err := tw.Write(content); err != nil {
			return nil, errors.Errorf("writing tar content: %w", err)
		}
	}

	// Close writers
	if err := tw.Close(); err != nil {
		return nil, errors.Errorf("closing tar writer: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, errors.Errorf("closing gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// GetFile returns the content of a file from the mock provider
func (m *MockProvider) GetFile(ctx context.Context, args Source, file string) ([]byte, error) {
	// Remove the path prefix if it exists
	cleanFile := strings.TrimPrefix(file, m.path+"/")
	content, ok := m.files[cleanFile]
	if !ok {
		return nil, errors.Errorf("file not found: %s", file)
	}

	// Return the content directly
	return content, nil
}

func (m *MockProvider) GetLicense(ctx context.Context, args Source, commitHash string) (LicenseEntry, error) {
	return LicenseEntry{
		SPDX:      "mock license",
		Permalink: "mock permalink",
		Name:      "mock name",
	}, nil
}
