package text

import (
	"context"
	"io"
)

// ReplacementRule defines a single text replacement operation
type ReplacementRule struct {
	// FromText is the text to replace
	FromText string

	// ToText is the replacement text
	ToText string

	// FileFilterGlob is a glob pattern to filter which files to apply the replacement to
	FileFilterGlob string
}

// ReplacementResult contains the results of a text replacement operation
type ReplacementResult struct {
	// WasModified indicates if any replacements were made
	WasModified bool

	// ReplacementCount is the number of replacements made
	ReplacementCount int

	// OriginalContent is the content before replacements
	OriginalContent []byte

	// ModifiedContent is the content after replacements
	ModifiedContent []byte
}

// TextReplacer defines the interface for text replacement operations
type TextReplacer interface {
	// ReplaceText applies a set of replacement rules to the content
	// Returns a ReplacementResult containing the modified content and metadata
	ReplaceText(ctx context.Context, content io.Reader, rules []ReplacementRule) (*ReplacementResult, error)

	// ValidateRules checks that all rules are valid
	ValidateRules(rules []ReplacementRule) error
}

// TODO(dr.methodical): 🧪 Add tests for ReplacementRule validation
// TODO(dr.methodical): 🧪 Add tests for ReplacementResult creation
// TODO(dr.methodical): 📝 Add examples of TextReplacer usage
