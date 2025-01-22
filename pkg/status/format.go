package status

import (
	"fmt"
)

// FileFormatter defines how file operations and status should be formatted
type FileFormatter interface {
	// FormatFileOperation formats a file operation status message
	FormatFileOperation(path, fileType, status string, isNew, isModified, isRemoved bool) string

	// FormatProgress formats a progress message
	FormatProgress(current, total int) string

	// FormatError formats an error message
	FormatError(err error) string
}

// DefaultFileFormatter provides a default implementation of FileFormatter
type DefaultFileFormatter struct{}

// NewDefaultFileFormatter creates a new DefaultFileFormatter
func NewDefaultFileFormatter() *DefaultFileFormatter {
	return &DefaultFileFormatter{}
}

// FormatFileOperation formats a file operation status message with emojis
func (f *DefaultFileFormatter) FormatFileOperation(path, fileType, status string, isNew, isModified, isRemoved bool) string {
	switch {
	case isNew:
		return fmt.Sprintf("✨ Created %s", path)
	case isModified:
		return fmt.Sprintf("📝 Modified %s", path)
	case isRemoved:
		return fmt.Sprintf("🗑️  Removed %s", path)
	case status == "error":
		return fmt.Sprintf("❌ Failed %s", path)
	default:
		return fmt.Sprintf("👍 Unchanged %s", path)
	}
}

// FormatProgress formats a progress message with percentage
func (f *DefaultFileFormatter) FormatProgress(current, total int) string {
	var percentage float64
	if total == 0 {
		percentage = 0
		if current > 0 {
			percentage = 100
		}
	} else {
		percentage = float64(current) / float64(total) * 100
	}

	if current >= total {
		return fmt.Sprintf("✅ Progress: %d/%d (%.0f%%)", current, total, percentage)
	}
	return fmt.Sprintf("⏳ Progress: %d/%d (%.0f%%)", current, total, percentage)
}

// FormatError formats an error message with emoji
func (f *DefaultFileFormatter) FormatError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("❌ Error: %v", err)
}
