package text

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleTextReplacer_ReplaceText(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		rules        []ReplacementRule
		want         string
		wantCount    int
		wantError    string
		wantModified bool
	}{
		{
			name:    "simple_replacement",
			content: "Hello World",
			rules: []ReplacementRule{
				{FromText: "World", ToText: "Universe"},
			},
			want:         "Hello Universe",
			wantCount:    1,
			wantModified: true,
		},
		{
			name:    "multiple_replacements",
			content: "Hello World World",
			rules: []ReplacementRule{
				{FromText: "World", ToText: "Universe"},
			},
			want:         "Hello Universe Universe",
			wantCount:    2,
			wantModified: true,
		},
		{
			name:    "multiple_rules",
			content: "Hello World",
			rules: []ReplacementRule{
				{FromText: "Hello", ToText: "Hi"},
				{FromText: "World", ToText: "Universe"},
			},
			want:         "Hi Universe",
			wantCount:    2,
			wantModified: true,
		},
		{
			name:    "no_match",
			content: "Hello World",
			rules: []ReplacementRule{
				{FromText: "Goodbye", ToText: "Hi"},
			},
			want:         "Hello World",
			wantCount:    0,
			wantModified: false,
		},
		{
			name:    "empty_content",
			content: "",
			rules: []ReplacementRule{
				{FromText: "World", ToText: "Universe"},
			},
			want:         "",
			wantCount:    0,
			wantModified: false,
		},
		{
			name:         "empty_rules",
			content:      "Hello World",
			rules:        []ReplacementRule{},
			want:         "Hello World",
			wantCount:    0,
			wantModified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replacer := NewSimpleTextReplacer()
			result, err := replacer.ReplaceText(
				context.Background(),
				strings.NewReader(tt.content),
				tt.rules,
			)

			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.content, string(result.OriginalContent))
			assert.Equal(t, tt.want, string(result.ModifiedContent))
			assert.Equal(t, tt.wantCount, result.ReplacementCount)
			assert.Equal(t, tt.wantModified, result.WasModified)
		})
	}
}

func TestSimpleTextReplacer_ValidateRules(t *testing.T) {
	tests := []struct {
		name      string
		rules     []ReplacementRule
		wantError string
	}{
		{
			name: "valid_rules",
			rules: []ReplacementRule{
				{
					FromText:       "foo",
					ToText:         "bar",
					FileFilterGlob: "*.txt",
				},
			},
		},
		{
			name: "missing_from_text",
			rules: []ReplacementRule{
				{
					ToText:         "bar",
					FileFilterGlob: "*.txt",
				},
			},
			wantError: "from_text is required",
		},
		{
			name: "missing_file_filter",
			rules: []ReplacementRule{
				{
					FromText: "foo",
					ToText:   "bar",
				},
			},
			wantError: "file_filter_glob is required",
		},
		{
			name:  "empty_rules",
			rules: []ReplacementRule{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replacer := NewSimpleTextReplacer()
			err := replacer.ValidateRules(tt.rules)

			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TODO(dr.methodical): 🧪 Add benchmark tests for large content
// TODO(dr.methodical): 🧪 Add tests for concurrent usage
// TODO(dr.methodical): 🧪 Add tests for context cancellation
