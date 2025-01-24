package text

import (
	"context"
	"io"
	"io/ioutil"
	"strings"

	"gitlab.com/tozd/go/errors"
)

// SimpleTextReplacer implements TextReplacer using basic string replacement
type SimpleTextReplacer struct{}

// NewSimpleTextReplacer creates a new SimpleTextReplacer
func NewSimpleTextReplacer() *SimpleTextReplacer {
	return &SimpleTextReplacer{}
}

// ReplaceText implements TextReplacer.ReplaceText
func (r *SimpleTextReplacer) ReplaceText(ctx context.Context, content io.Reader, rules []ReplacementRule) (*ReplacementResult, error) {
	// Read all content
	originalContent, err := ioutil.ReadAll(content)
	if err != nil {
		return nil, errors.Errorf("reading content: %w", err)
	}

	// Create result with original content
	result := &ReplacementResult{
		OriginalContent: originalContent,
		ModifiedContent: originalContent,
	}

	// Apply each rule
	currentContent := string(originalContent)
	for _, rule := range rules {
		// Skip empty rules
		if rule.FromText == "" {
			continue
		}

		// Apply replacement
		newContent := strings.ReplaceAll(currentContent, rule.FromText, rule.ToText)

		// Update counts if changed
		if newContent != currentContent {
			result.WasModified = true
			result.ReplacementCount += strings.Count(currentContent, rule.FromText)
		}

		currentContent = newContent
	}

	// Update final content
	result.ModifiedContent = []byte(currentContent)
	return result, nil
}

// ValidateRules implements TextReplacer.ValidateRules
func (r *SimpleTextReplacer) ValidateRules(rules []ReplacementRule) error {
	for i, rule := range rules {
		if rule.FromText == "" {
			return errors.Errorf("rule %d: from_text is required", i)
		}
		if rule.FileFilterGlob == "" {
			return errors.Errorf("rule %d: file_filter_glob is required", i)
		}
	}
	return nil
}

// TODO(dr.methodical): 🧪 Add tests for edge cases (empty content, empty rules)
// TODO(dr.methodical): 🧪 Add tests for multiple replacements
// TODO(dr.methodical): 🧪 Add benchmarks for large content
