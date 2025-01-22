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
	"strings"

	"github.com/fatih/color"
)

// 🎨 Display configuration
const (
	fileIndent  = 4  // spaces to indent file entries
	nameWidth   = 35 // Base width for filename
	typeWidth   = 15 // Width for file type
	statusWidth = 15 // Width for status text
)

// 🎯 FormatFileOperation formats a file operation for display
func FormatFileOperation(path, fileType, status string, isNew, isModified, isRemoved bool) string {
	// Determine prefix symbol
	var prefix string
	switch {
	case isNew:
		prefix = color.GreenString("✓")
	case isModified:
		prefix = color.YellowString("⟳")
	case isRemoved:
		prefix = color.RedString("✗")
	default:
		prefix = color.HiBlackString("-")
	}

	// Format parts with padding
	namePart := fmt.Sprintf("%-*s", nameWidth, path)
	typePart := fmt.Sprintf("%-*s", typeWidth, fileType)
	statusPart := fmt.Sprintf("%-*s", statusWidth, status)

	// Build final string with indentation
	return fmt.Sprintf("%s%s %s %s %s",
		strings.Repeat(" ", fileIndent),
		prefix,
		namePart,
		typePart,
		statusPart,
	)
}
