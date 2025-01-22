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
	"os"
	"path/filepath"
	"sort"
	"time"

	"gitlab.com/tozd/go/errors"
)

// 📝 LogConfig represents logging configuration
type LogConfig struct {
	Level      string            `json:"level"`       // Log level (trace, debug, info, warn, error)
	Format     string            `json:"format"`      // Log format (text, json)
	File       string            `json:"file"`        // Log file path
	Rotation   bool              `json:"rotation"`    // Whether to rotate logs
	MaxSize    int               `json:"max_size"`    // Maximum size in MB before rotating
	MaxAge     int               `json:"max_age"`     // Maximum age in days before deleting
	MaxBackups int               `json:"max_backups"` // Maximum number of old log files to retain
	Compress   bool              `json:"compress"`    // Whether to compress rotated files
	Fields     map[string]string `json:"fields"`      // Additional fields to add to all logs
}

// 📝 LogState represents the current logging state
type LogState struct {
	Config      LogConfig `json:"config"`       // Current logging configuration
	LastRotate  time.Time `json:"last_rotate"`  // Last rotation time
	LastCleanup time.Time `json:"last_cleanup"` // Last cleanup time
	FileSize    int64     `json:"file_size"`    // Current log file size
	LogFiles    []string  `json:"log_files"`    // List of log files
}

// 🔄 UpdateLogConfig updates the logging configuration
func (m *Manager) UpdateLogConfig(ctx context.Context, cfg LogConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create log directory if needed
	if cfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0755); err != nil {
			return errors.Errorf("creating log directory: %w", err)
		}
	}

	// Update status
	if m.status.LogState == nil {
		m.status.LogState = &LogState{
			LastRotate:  time.Now(),
			LastCleanup: time.Now(),
			LogFiles:    []string{},
		}
	}
	m.status.LogState.Config = cfg

	return m.save()
}

// 🔍 GetLogConfig gets the current logging configuration
func (m *Manager) GetLogConfig() *LogConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.status.LogState == nil {
		return nil
	}
	cfg := m.status.LogState.Config
	return &cfg
}

// 📝 LogRotate performs log rotation if needed
func (m *Manager) LogRotate(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status.LogState == nil || m.status.LogState.Config.File == "" {
		return nil
	}

	cfg := m.status.LogState.Config
	if !cfg.Rotation {
		return nil
	}

	// Check if rotation is needed
	info, err := os.Stat(cfg.File)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Errorf("checking log file: %w", err)
	}

	if info.Size() < int64(cfg.MaxSize*1024*1024) {
		return nil
	}

	// Rotate file
	rotated := cfg.File + "." + time.Now().Format("2006-01-02-15-04-05")
	if err := os.Rename(cfg.File, rotated); err != nil {
		return errors.Errorf("rotating log file: %w", err)
	}

	// Update state
	m.status.LogState.LastRotate = time.Now()
	m.status.LogState.FileSize = 0
	m.status.LogState.LogFiles = append(m.status.LogState.LogFiles, filepath.Base(rotated))

	// Cleanup old files if needed
	if cfg.MaxBackups > 0 || cfg.MaxAge > 0 {
		if err := m.cleanupLogs(ctx); err != nil {
			return errors.Errorf("cleaning up logs: %w", err)
		}
	}

	return m.save()
}

// 🧹 cleanupLogs removes old log files
func (m *Manager) cleanupLogs(ctx context.Context) error {
	if m.status.LogState == nil {
		return nil
	}

	cfg := m.status.LogState.Config
	if !cfg.Rotation {
		return nil
	}

	dir := filepath.Dir(cfg.File)
	base := filepath.Base(cfg.File)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return errors.Errorf("reading log directory: %w", err)
	}

	var newFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Base(entry.Name()) == base {
			continue
		}

		matched, err := filepath.Match(base+".*", entry.Name())
		if err != nil {
			continue
		}
		if !matched {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if file is too old
		if cfg.MaxAge > 0 {
			age := time.Since(info.ModTime())
			if age.Hours() > float64(cfg.MaxAge*24) {
				os.Remove(path)
				continue
			}
		}

		newFiles = append(newFiles, entry.Name())
	}

	// Sort files by modification time (newest first)
	sort.Slice(newFiles, func(i, j int) bool {
		iPath := filepath.Join(dir, newFiles[i])
		jPath := filepath.Join(dir, newFiles[j])

		iInfo, err := os.Stat(iPath)
		if err != nil {
			return false
		}
		jInfo, err := os.Stat(jPath)
		if err != nil {
			return true
		}

		return iInfo.ModTime().After(jInfo.ModTime())
	})

	// Keep only the configured number of backups
	if cfg.MaxBackups > 0 && len(newFiles) > cfg.MaxBackups {
		for _, file := range newFiles[cfg.MaxBackups:] {
			path := filepath.Join(dir, file)
			os.Remove(path)
		}
		newFiles = newFiles[:cfg.MaxBackups]
	}

	// Update state
	m.status.LogState.LastCleanup = time.Now()
	m.status.LogState.LogFiles = newFiles
	return m.save()
}
