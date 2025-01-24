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
	"fmt"
)

// Emoji constants for file status
const (
	// 🎨 Status Emojis
	EmojiCreated   = "✨"
	EmojiModified  = "📝"
	EmojiRemoved   = "🗑️"
	EmojiUnchanged = "👍"
	EmojiError     = "❌"
	EmojiProgress  = "⏳"
	EmojiComplete  = "✅"
)

// Message format constants
const (
	// 📝 Message Templates
	MsgCreated   = "%s Created %s"
	MsgModified  = "%s Modified %s"
	MsgRemoved   = "%s Removed %s"
	MsgUnchanged = "%s Unchanged %s"
	MsgError     = "%s Failed %s"
	MsgProgress  = "%s Progress: %d/%d (%d%%)"
)

// 🎨 FileFormatter defines how file operations are formatted for display
type FileFormatter interface {
	FormatHeader() string
	FormatSectionHeader(path string) string
	FormatRepoInfo(repo, ref string) string
	FormatFileStatus(filename string, status FileStatus, metadata map[string]string) string
	FormatFileOperation(path, fileType, status string, isNew, isModified, isRemoved bool) string
	FormatProgress(current, total int) string
	FormatError(err error) string
}

// DefaultFileFormatter provides a basic implementation of FileFormatter
type DefaultFileFormatter struct{}

// NewDefaultFileFormatter creates a new DefaultFileFormatter
func NewDefaultFileFormatter() *DefaultFileFormatter {
	return &DefaultFileFormatter{}
}

func (f *DefaultFileFormatter) FormatHeader() string {
	return "copyrc • file operations"
}

func (f *DefaultFileFormatter) FormatSectionHeader(path string) string {
	return fmt.Sprintf("[syncing %s]", path)
}

func (f *DefaultFileFormatter) FormatRepoInfo(repo, ref string) string {
	return fmt.Sprintf("♦ %s • %s", repo, ref)
}

func (f *DefaultFileFormatter) FormatFileStatus(filename string, status FileStatus, metadata map[string]string) string {
	switch status {
	case StatusNew:
		return fmt.Sprintf(MsgCreated, EmojiCreated, filename)
	case StatusModified:
		return fmt.Sprintf(MsgModified, EmojiModified, filename)
	case StatusDeleted:
		return fmt.Sprintf(MsgRemoved, EmojiRemoved, filename)
	case StatusUnchanged:
		return fmt.Sprintf(MsgUnchanged, EmojiUnchanged, filename)
	default:
		return fmt.Sprintf(MsgError, EmojiError, filename)
	}
}

func (f *DefaultFileFormatter) FormatFileOperation(path, fileType, status string, isNew, isModified, isRemoved bool) string {
	if status == "error" {
		return fmt.Sprintf("❌ Failed %s", path)
	}
	if isNew {
		return fmt.Sprintf("✨ Created %s", path)
	}
	if isModified {
		return fmt.Sprintf("📝 Modified %s", path)
	}
	if isRemoved {
		return fmt.Sprintf("🗑️  Removed %s", path)
	}
	return fmt.Sprintf("👍 Unchanged %s", path)
}

func (f *DefaultFileFormatter) FormatProgress(current, total int) string {
	// Handle negative values
	if current < 0 || total < 0 {
		return fmt.Sprintf(MsgProgress, EmojiProgress, 0, 0, 0)
	}

	// Handle zero total
	if total == 0 {
		return fmt.Sprintf(MsgProgress, EmojiComplete, current, total, 0)
	}

	// Calculate percentage
	percentage := (current * 100) / total
	if percentage > 100 {
		percentage = 100
	}

	// Choose emoji based on completion
	emoji := EmojiProgress
	if current >= total {
		emoji = EmojiComplete
	}

	return fmt.Sprintf(MsgProgress, emoji, current, total, percentage)
}

func (f *DefaultFileFormatter) FormatError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%s Error: %s", EmojiError, err.Error())
}
