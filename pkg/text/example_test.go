package text_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/walteh/copyrc/pkg/text"
)

func ExampleSimpleTextReplacer_ReplaceText() {
	// Create a replacer
	replacer := text.NewSimpleTextReplacer()

	// Define some replacement rules
	rules := []text.ReplacementRule{
		{
			FromText:       "World",
			ToText:         "Universe",
			FileFilterGlob: "*.txt",
		},
		{
			FromText:       "Hello",
			ToText:         "Hi",
			FileFilterGlob: "*.txt",
		},
	}

	// Create some content
	content := strings.NewReader("Hello World!")

	// Apply replacements
	result, err := replacer.ReplaceText(context.Background(), content, rules)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print results
	fmt.Printf("Original: %s\n", result.OriginalContent)
	fmt.Printf("Modified: %s\n", result.ModifiedContent)
	fmt.Printf("Changes: %d\n", result.ReplacementCount)
	fmt.Printf("Was Modified: %v\n", result.WasModified)

	// Output:
	// Original: Hello World!
	// Modified: Hi Universe!
	// Changes: 2
	// Was Modified: true
}

func ExampleSimpleTextReplacer_ValidateRules() {
	// Create a replacer
	replacer := text.NewSimpleTextReplacer()

	// Define some rules
	rules := []text.ReplacementRule{
		{
			FromText:       "foo",
			ToText:         "bar",
			FileFilterGlob: "*.txt",
		},
		{
			FromText: "baz", // Missing FileFilterGlob
			ToText:   "qux",
		},
	}

	// Validate rules
	err := replacer.ValidateRules(rules)
	fmt.Printf("Validation error: %v\n", err)

	// Output:
	// Validation error: rule 1: file_filter_glob is required
}

// TODO(dr.methodical): 📝 Add example with multiple replacements
// TODO(dr.methodical): 📝 Add example with file glob filtering
// TODO(dr.methodical): 📝 Add example with error handling
