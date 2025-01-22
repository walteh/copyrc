package main

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertLogOutput is a helper function that verifies log output formatting
func assertLogOutput(t *testing.T, line string) {
	t.Helper()

	// Trim any trailing whitespace while preserving internal spacing
	line = strings.TrimRight(line, " \n\r\t")
	t.Logf("Checking line: %q", line)

	// Basic format checks
	assert.True(t, len(line) >= 40, "line should be properly padded: %q", line)

	// Check bullet/dash/symbol prefix
	firstRune := []rune(strings.TrimLeft(line, " "))[0]
	validRunes := []rune{'•', '-', '✓', '⟳'}
	isValid := false
	for _, r := range validRunes {
		if firstRune == r {
			isValid = true
			break
		}
	}
	assert.True(t, isValid, "line should start with bullet, dash, or status symbol: %q (got %U)", line, firstRune)

	// Split the line into fields, preserving spacing
	parts := strings.Fields(line)
	require.GreaterOrEqual(t, len(parts), 3, "line should have at least 3 parts: %q", line)

	// Find the type field - it should be one of managed/local/copy
	// The type is usually after the filename and before any status
	var foundType bool
	for i, part := range parts {
		if i > 0 && (part == "managed" || part == "local" || part == "copy" || strings.HasPrefix(part, "copy")) {
			foundType = true
			break
		}
	}
	assert.True(t, foundType, "line should contain a valid type (managed/local/copy): %q", line)

	// Check alignment of type field
	// The type field should start at a consistent column
	typeStart := -1
	for _, typ := range []string{"managed", "local", "copy"} {
		if idx := strings.LastIndex(line, typ); idx != -1 {
			typeStart = idx
			break
		}
	}
	assert.True(t, typeStart >= 35, "type field should be properly aligned (>= column 35): %q", line)
}

func TestLogFileOperation(t *testing.T) {
	tests := []struct {
		name     string
		opts     FileInfo
		expected string
	}{
		{
			name: "managed_file_new",
			opts: FileInfo{
				Name:      ".copyrc.lock",
				IsNew:     true,
				IsManaged: true,
			},
			expected: "    ✓ .copyrc.lock                        copy            NEW            \n",
		},
		{
			name: "managed_file_updated",
			opts: FileInfo{
				Name:       ".copyrc.lock",
				IsModified: true,
				IsManaged:  true,
			},
			expected: "    ⟳ .copyrc.lock                        copy            UPDATED        \n",
		},
		{
			name: "managed_file_unchanged",
			opts: FileInfo{
				Name:      ".copyrc.lock",
				IsManaged: true,
			},
			expected: "    • .copyrc.lock                        copy                           \n",
		},
		{
			name: "local_file",
			opts: FileInfo{
				Name:        "custom.patch.go",
				IsUntracked: true,
			},
			expected: "    - custom.patch.go                     local                          \n",
		},
		{
			name: "copy_file_with_replacements",
			opts: FileInfo{
				Name:         "README.copy.md",
				Replacements: 7,
			},
			expected: "    • README.copy.md                      copy [7]                       \n",
		},
		{
			name: "copy_file_updated",
			opts: FileInfo{
				Name:         "main.copy.go",
				IsModified:   true,
				Replacements: 11,
			},
			expected: "    ⟳ main.copy.go                        copy [11]       UPDATED        \n",
		},
		{
			name: "embed_gen_file",
			opts: FileInfo{
				Name:      "embed.gen.go",
				IsManaged: true,
			},
			expected: "    • embed.gen.go                        copy                           \n",
		},
		{
			name: "tarball_file",
			opts: FileInfo{
				Name: "nvim-lspconfig.tar.gz",
			},
			expected: "    • nvim-lspconfig.tar.gz               copy                           \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := newTestLogger(t)
			ctx = NewLoggerInContext(ctx, logger)

			// Reset the processed files map
			processedFiles = sync.Map{}

			// Call logFileOperation
			logFileOperation(ctx, tt.opts)

			output := logger.CopyOfCurrentConsoleOutputInTest()
			t.Logf("Output: %q", output)

			// Verify output
			assert.Equal(t, tt.expected, output, "output should match expected format")
			assertLogOutput(t, strings.TrimSpace(output))
		})
	}
}

func TestLogFileOperationAlignment(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger(t)
	ctx = NewLoggerInContext(ctx, logger)

	// Reset the processed files map
	processedFiles = sync.Map{}

	// Log multiple files to test alignment
	files := []FileInfo{
		{
			Name:        "short.go",
			IsUntracked: true,
		},
		{
			Name:         "very_long_filename.copy.go",
			IsModified:   true,
			Replacements: 100,
		},
		{
			Name:      ".copyrc.lock",
			IsNew:     true,
			IsManaged: true,
		},
	}

	for _, f := range files {
		logFileOperation(ctx, f)
	}

	output := logger.CopyOfCurrentConsoleOutputInTest()
	t.Logf("Output:\n%s", output)

	// Check alignment properties
	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.Len(t, lines, len(files), "should have correct number of lines")

	// Check each line's format
	for _, line := range lines {
		assertLogOutput(t, line)
	}
}

func TestLogSimple(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger(t)
	ctx = NewLoggerInContext(ctx, logger)

	// Reset the processed files map
	processedFiles = sync.Map{}

	// Log the same file twice
	opts := FileInfo{
		Name:         "test.go",
		IsModified:   true,
		Replacements: 2,
	}

	// First call should log
	logFileOperation(ctx, opts)
	firstOutput := logger.CopyOfCurrentConsoleOutputInTest()
	assert.NotEmpty(t, firstOutput, "first call should produce output")
	assertLogOutput(t, strings.TrimSpace(firstOutput))
}
