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
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/provider"
	"gitlab.com/tozd/go/errors"
)

// 📝 FileEntry represents a file in the status file
type FileEntry struct {
	Hash         string            `json:"hash"`         // Content hash
	Replacements map[string]string `json:"replacements"` // Applied replacements
	Source       string            `json:"source"`       // Source URL
}

// 📝 Status represents the complete status file
type Status struct {
	Files      map[string]FileEntry `json:"files"`       // File entries
	CommitHash string               `json:"commit_hash"` // Repository commit hash
	Config     *config.Config       `json:"config"`      // Configuration snapshot
}

// 🎯 Manager handles status tracking
type Manager struct {
	mu     sync.RWMutex
	status *Status
	path   string
}

// 🏭 New creates a new status manager
func New(path string) (*Manager, error) {
	mgr := &Manager{
		path: path,
	}

	// Load existing status
	if err := mgr.load(); err != nil {
		return nil, errors.Errorf("loading status: %w", err)
	}

	return mgr, nil
}

// 🔍 load loads the status from disk
func (m *Manager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create empty status if file doesn't exist
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		m.status = &Status{
			Files: make(map[string]FileEntry),
		}
		return nil
	}

	// Read status file
	data, err := os.ReadFile(m.path)
	if err != nil {
		return errors.Errorf("reading status file: %w", err)
	}

	// Parse JSON
	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return errors.Errorf("parsing status file: %w", err)
	}

	m.status = &status
	return nil
}

// 💾 save saves the status to disk
func (m *Manager) save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return errors.Errorf("creating status directory: %w", err)
	}

	// Marshal JSON
	data, err := json.MarshalIndent(m.status, "", "  ")
	if err != nil {
		return errors.Errorf("marshaling status: %w", err)
	}

	// Write file
	if err := os.WriteFile(m.path, data, 0644); err != nil {
		return errors.Errorf("writing status file: %w", err)
	}

	return nil
}

// 🔄 Update updates the status for a file
func (m *Manager) Update(ctx context.Context, path string, entry FileEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.status.Files[path] = entry
	return m.save()
}

// 🔄 UpdateCommitHash updates the commit hash
func (m *Manager) UpdateCommitHash(ctx context.Context, hash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.status.CommitHash = hash
	return m.save()
}

// 🔄 UpdateConfig updates the configuration snapshot
func (m *Manager) UpdateConfig(ctx context.Context, cfg *config.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.status.Config = cfg
	return m.save()
}

// 🔍 Get gets the status for a file
func (m *Manager) Get(path string) (FileEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.status.Files[path]
	return entry, ok
}

// 🔍 GetCommitHash gets the commit hash
func (m *Manager) GetCommitHash() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.status.CommitHash
}

// 🔍 GetConfig gets the configuration snapshot
func (m *Manager) GetConfig() *config.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.status.Config
}

// 🔍 CheckStatus checks if files are up to date
func (m *Manager) CheckStatus(ctx context.Context, cfg *config.Config, p provider.Provider) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check commit hash
	remoteHash, err := p.GetCommitHash(ctx, cfg.Provider)
	if err != nil {
		return errors.Errorf("getting remote commit hash: %w", err)
	}

	if remoteHash != m.status.CommitHash {
		return errors.Errorf("commit hash mismatch: local=%s remote=%s", m.status.CommitHash, remoteHash)
	}

	return nil
}

// 🧹 Clean removes all status information
func (m *Manager) Clean() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.status = &Status{
		Files: make(map[string]FileEntry),
	}
	return m.save()
}

// 🔍 List lists all tracked files
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files := make([]string, 0, len(m.status.Files))
	for file := range m.status.Files {
		files = append(files, file)
	}
	return files
}
