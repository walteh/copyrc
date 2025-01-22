package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 🧪 TestDefaultFileFormatter tests the default file formatter implementation
func TestDefaultFileFormatter(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		fileType    string
		status      string
		isNew       bool
		isModified  bool
		isRemoved   bool
		want        string
		description string
	}{
		{
			name:        "new_file",
			path:        "test.txt",
			fileType:    "file",
			status:      "ok",
			isNew:       true,
			isModified:  false,
			isRemoved:   false,
			want:        "✨ Created test.txt",
			description: "should show creation symbol for new files",
		},
		{
			name:        "modified_file",
			path:        "config.yaml",
			fileType:    "file",
			status:      "ok",
			isNew:       false,
			isModified:  true,
			isRemoved:   false,
			want:        "📝 Modified config.yaml",
			description: "should show modification symbol for changed files",
		},
		{
			name:        "removed_file",
			path:        "old.txt",
			fileType:    "file",
			status:      "ok",
			isNew:       false,
			isModified:  false,
			isRemoved:   true,
			want:        "🗑️  Removed old.txt",
			description: "should show removal symbol for deleted files",
		},
		{
			name:        "unchanged_file",
			path:        "stable.txt",
			fileType:    "file",
			status:      "ok",
			isNew:       false,
			isModified:  false,
			isRemoved:   false,
			want:        "👍 Unchanged stable.txt",
			description: "should show unchanged symbol for stable files",
		},
		{
			name:        "error_status",
			path:        "error.txt",
			fileType:    "file",
			status:      "error",
			isNew:       false,
			isModified:  false,
			isRemoved:   false,
			want:        "❌ Failed error.txt",
			description: "should show error symbol for failed operations",
		},
		// 🧪 Edge cases for medical-grade coverage
		{
			name:        "empty_path",
			path:        "",
			fileType:    "file",
			status:      "ok",
			isNew:       true,
			isModified:  false,
			isRemoved:   false,
			want:        "✨ Created ",
			description: "should handle empty path gracefully",
		},
		{
			name:        "multiple_states",
			path:        "conflict.txt",
			fileType:    "file",
			status:      "ok",
			isNew:       true,
			isModified:  true,
			isRemoved:   false,
			want:        "✨ Created conflict.txt", // New takes precedence
			description: "should handle multiple states with precedence",
		},
	}

	formatter := NewDefaultFileFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.FormatFileOperation(tt.path, tt.fileType, tt.status, tt.isNew, tt.isModified, tt.isRemoved)
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

// 🧪 TestProgressFormatting tests progress message formatting
func TestProgressFormatting(t *testing.T) {
	tests := []struct {
		name        string
		current     int
		total       int
		want        string
		description string
	}{
		{
			name:        "zero_progress",
			current:     0,
			total:       100,
			want:        "⏳ Progress: 0/100 (0%)",
			description: "should show 0% progress",
		},
		{
			name:        "half_progress",
			current:     50,
			total:       100,
			want:        "⏳ Progress: 50/100 (50%)",
			description: "should show 50% progress",
		},
		{
			name:        "complete",
			current:     100,
			total:       100,
			want:        "✅ Progress: 100/100 (100%)",
			description: "should show completion symbol at 100%",
		},
		// 🧪 Edge cases for medical-grade coverage
		{
			name:        "zero_total",
			current:     0,
			total:       0,
			want:        "✅ Progress: 0/0 (0%)",
			description: "should handle zero total gracefully",
		},
		{
			name:        "zero_total_with_current",
			current:     5,
			total:       0,
			want:        "✅ Progress: 5/0 (100%)",
			description: "should handle zero total with positive current",
		},
		{
			name:        "current_exceeds_total",
			current:     150,
			total:       100,
			want:        "✅ Progress: 150/100 (150%)",
			description: "should handle current exceeding total",
		},
		{
			name:        "negative_values",
			current:     -10,
			total:       100,
			want:        "⏳ Progress: -10/100 (-10%)",
			description: "should handle negative values",
		},
	}

	formatter := NewDefaultFileFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.FormatProgress(tt.current, tt.total)
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

// 🧪 TestErrorFormatting tests error message formatting
func TestErrorFormatting(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		want        string
		description string
	}{
		{
			name:        "simple_error",
			err:         assert.AnError,
			want:        "❌ Error: assert.AnError general error for testing",
			description: "should format simple errors",
		},
		{
			name:        "nil_error",
			err:         nil,
			want:        "",
			description: "should return empty string for nil errors",
		},
	}

	formatter := NewDefaultFileFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.FormatError(tt.err)
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}
