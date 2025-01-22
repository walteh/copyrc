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

package log

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	// Disable color for testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	tests := []struct {
		name     string
		op       func(t *testing.T, logger *Logger)
		wantLogs []string
	}{
		{
			name: "log_file_operation",
			op: func(t *testing.T, logger *Logger) {
				logger.LogFileOperation(context.Background(), FileOperation{
					Path:         "test.txt",
					Type:         "managed",
					Status:       "NEW",
					IsNew:        true,
					IsManaged:    true,
					Replacements: 2,
				})
			},
			wantLogs: []string{
				"    ✓ test.txt                          managed         NEW            ",
			},
		},
		{
			name: "log_repo_operation",
			op: func(t *testing.T, logger *Logger) {
				logger.StartRepoOperation(context.Background(), RepoOperation{
					Name:        "github.com/test/repo",
					Ref:         "main",
					Destination: "/tmp/test",
				})
			},
			wantLogs: []string{
				"[syncing /tmp/test]",
				"◆ github.com/test/repo • main",
			},
		},
		{
			name: "log_messages",
			op: func(t *testing.T, logger *Logger) {
				logger.Info("info message")
				logger.Warning("warning message")
				logger.Error("error message")
				logger.Success("success message")
			},
			wantLogs: []string{
				"ℹ️  info message",
				"⚠️  warning message",
				"❌ error message",
				"✅ success message",
			},
		},
		{
			name: "log_formatted_messages",
			op: func(t *testing.T, logger *Logger) {
				logger.Infof("info %s", "test")
				logger.Warningf("warning %s", "test")
				logger.Errorf("error %s", "test")
				logger.Successf("success %s", "test")
			},
			wantLogs: []string{
				"ℹ️  info test",
				"⚠️  warning test",
				"❌ error test",
				"✅ success test",
			},
		},
		{
			name: "log_header",
			op: func(t *testing.T, logger *Logger) {
				logger.Header("syncing repository files")
			},
			wantLogs: []string{
				"copyrc • syncing repository files",
			},
		},
		{
			name: "log_newline",
			op: func(t *testing.T, logger *Logger) {
				logger.Info("first")
				logger.LogNewline()
				logger.Info("second")
			},
			wantLogs: []string{
				"ℹ️  first",
				"",
				"ℹ️  second",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create buffer for console output
			buf := &bytes.Buffer{}
			logger := New(buf, zerolog.InfoLevel)

			// Perform operation
			tt.op(t, logger)

			// Check output
			output := strings.TrimSpace(buf.String())
			lines := strings.Split(output, "\n")

			require.Equal(t, len(tt.wantLogs), len(lines), "number of log lines should match")
			for i, want := range tt.wantLogs {
				assert.Equal(t, want, strings.TrimSpace(lines[i]), "log line %d should match", i)
			}
		})
	}
}

func TestLoggerContext(t *testing.T) {
	// Create logger
	logger := New(io.Discard, zerolog.InfoLevel)

	// Add to context
	ctx := context.Background()
	ctx = NewContext(ctx, logger)

	// Get from context
	got := FromContext(ctx)
	assert.Same(t, logger, got, "logger from context should be the same instance")

	// Check panic on missing logger
	assert.Panics(t, func() {
		FromContext(context.Background())
	}, "FromContext should panic when logger is missing")
}

func TestFileOperationFormatting(t *testing.T) {
	// Disable color for testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	tests := []struct {
		name string
		op   FileOperation
		want string
	}{
		{
			name: "new_managed_file",
			op: FileOperation{
				Path:      "test.txt",
				Type:      "managed",
				Status:    "NEW",
				IsNew:     true,
				IsManaged: true,
			},
			want: "    ✓ test.txt                          managed         NEW            ",
		},
		{
			name: "modified_copy_file",
			op: FileOperation{
				Path:       "test.txt",
				Type:       "copy",
				Status:     "UPDATED",
				IsModified: true,
			},
			want: "    ⟳ test.txt                          copy            UPDATED        ",
		},
		{
			name: "removed_file",
			op: FileOperation{
				Path:      "test.txt",
				Type:      "local",
				Status:    "REMOVED",
				IsRemoved: true,
			},
			want: "    ✗ test.txt                          local           REMOVED        ",
		},
		{
			name: "unchanged_file",
			op: FileOperation{
				Path:   "test.txt",
				Type:   "copy",
				Status: "no change",
			},
			want: "    • test.txt                          copy            no change      ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create buffer for console output
			buf := &bytes.Buffer{}
			logger := New(buf, zerolog.InfoLevel)

			// Log operation
			logger.LogFileOperation(context.Background(), tt.op)

			// Check output
			output := strings.TrimSpace(buf.String())
			assert.Equal(t, tt.want, output, "formatted output should match")
		})
	}
}
