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

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

// FileStatus represents the state of a file
type FileStatus struct {
	Symbol rune
	Style  StatusStyle
	Text   string
}

type StatusStyle struct {
	SymbolColor color.Attribute
	TextColor   color.Attribute
}

// FileInfo represents a file with its status and metadata
type FileInfo struct {
	Name string
	// Status       FileStatus
	IsManaged    bool
	IsModified   bool
	IsRemoved    bool
	IsNew        bool
	IsUntracked  bool
	IsCustomized bool
	Replacements int // Number of replacements made to this file
}

// FileType represents the source/type of a file
type FileType struct {
	Name  string
	Color color.Attribute
}

var (
	FileTypeManaged    = FileType{Name: "managed", Color: ManagedColor}
	FileTypeLocal      = FileType{Name: "local", Color: LocalColor}
	FileTypeCopy       = FileType{Name: "copy", Color: CopyColor}
	FileTypeCustomized = FileType{Name: "customized", Color: CustomizedColor}
)

func (me FileType) ColorString() string {
	return color.New(me.Color).Sprint(me.Name)
}

func (me FileType) UncoloredString() string {
	return me.Name
}

func (me FileType) ColorStringWithReplacements(replacements int) string {
	return color.New(me.Color).Sprintf("%s [%d]", me.Name, replacements)
}

func (me FileType) UncoloredStringWithReplacements(replacements int) string {
	if replacements == -1 {
		return fmt.Sprintf("%s [~?]", me.Name)
	}
	return fmt.Sprintf("%s [~%d]", me.Name, replacements)
}

// FileChangeStatus represents the change state of a file
// RepoDisplay represents how a repository should be displayed
type RepoDisplay struct {
	Name        string
	Ref         string
	Destination string
	IsArchive   bool
	Files       []FileInfo
}

// Display configuration
const (
	fileIndent  = 4  // spaces to indent file entries
	nameWidth   = 35 // Base width for filename
	typeWidth   = 15 // Width for file type
	statusWidth = 15 // Width for status text
)

// Status definitions
var (
	LocalColor   = color.FgYellow
	CopyColor    = color.FgBlue
	ManagedColor = color.FgCyan

	CustomizedColor = color.FgMagenta

	CustomizedFile = FileStatus{
		// pencil
		Symbol: '✎',
		Style: StatusStyle{
			SymbolColor: CustomizedColor,
			TextColor:   color.Faint,
		},
		Text: "CUSTOMIZED",
	}

	UnmodifiedCopyFile = FileStatus{
		Symbol: '•',
		Style: StatusStyle{
			SymbolColor: color.Faint,
			TextColor:   color.Faint,
		},
	}

	UnmodifiedManagedFile = FileStatus{
		Symbol: '•',
		Style: StatusStyle{
			SymbolColor: ManagedColor,
			TextColor:   color.Faint,
		},
	}

	UntrackedFile = FileStatus{
		Symbol: '-',
		Style: StatusStyle{
			SymbolColor: color.FgYellow,
			TextColor:   color.Faint,
		},
	}

	NewFile = FileStatus{
		Symbol: '✓',
		Style: StatusStyle{
			SymbolColor: color.FgGreen,
			TextColor:   color.Faint,
		},
		Text: "NEW",
	}

	UpdatedFile = FileStatus{
		Symbol: '⟳',
		Style: StatusStyle{
			SymbolColor: color.FgBlue,
			TextColor:   color.Faint,
		},
		Text: "UPDATED",
	}

	RemovedFile = FileStatus{
		Symbol: '✗',
		Style: StatusStyle{
			SymbolColor: color.FgRed,
			TextColor:   color.Faint,
		},
		Text: "REMOVED",
	}
)

type Logger struct {
	zlog            zerolog.Logger
	consoleOut      io.Writer
	mu              sync.Mutex
	currentRepo     *RepoDisplay
	repoMu          sync.Mutex
	longestNeighbor int
}

type loggerContextKey struct{}

func loggerFromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(loggerContextKey{}).(*Logger)
	if !ok {
		panic("logger not found in context")
	}
	return logger
}

func NewLoggerInContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, l)
}

func newTestLogger(t *testing.T) *Logger {
	console := &testBuffer{Buffer: bytes.NewBuffer(nil)}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Caller().Logger()
	return &Logger{
		zlog:       zlog,
		consoleOut: console,
		mu:         sync.Mutex{},
	}
}

// testBuffer is a wrapper around bytes.Buffer that ensures consistent newline handling
type testBuffer struct {
	*bytes.Buffer
}

func (b *testBuffer) Write(p []byte) (n int, err error) {
	return b.Buffer.Write(p)
}

func (me *Logger) CopyOfCurrentConsoleOutputInTest() string {
	me.mu.Lock()
	defer me.mu.Unlock()

	return me.consoleOut.(*testBuffer).String()
}

func NewDiscardDebugLogger(console io.Writer) *Logger {
	// Configure zerolog to write to a discarded writer in tests
	// This ensures our test assertions only see the formatted output
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog := zerolog.New(io.Discard).With().Timestamp().Caller().Logger()

	return &Logger{
		zlog:       zlog,
		consoleOut: console,
		mu:         sync.Mutex{},
	}
}

func (me FileInfo) Status() FileStatus {
	if me.IsCustomized {
		return CustomizedFile
	} else if me.IsUntracked {
		return UntrackedFile
	} else if me.IsRemoved {
		return RemovedFile
	} else if me.IsNew {
		return NewFile
	} else if me.IsModified {
		return UpdatedFile
	} else {
		if me.IsManaged {
			return UnmodifiedManagedFile
		} else {
			return UnmodifiedCopyFile
		}
	}
}

func (me FileInfo) Type() FileType {
	if me.IsUntracked {
		return FileTypeLocal
	} else if me.IsManaged {
		return FileTypeManaged
	} else if me.IsCustomized {
		return FileTypeCustomized
	} else {
		return FileTypeCopy
	}
}

func (l *Logger) formatFileOperation(opts FileInfo) string {
	// Build filename part
	effectiveWidth := nameWidth
	if l.longestNeighbor > 0 {
		effectiveWidth = l.longestNeighbor + 2 // Add some padding
	}
	namePart := fmt.Sprintf("%-*s", effectiveWidth, opts.Name)

	// Build type part with optional replacements
	var typePart string
	if opts.Replacements != 0 && (opts.Type().Name == FileTypeCopy.Name || opts.Type().Name == FileTypeCustomized.Name) {
		typePart = fmt.Sprintf("%-*s", typeWidth, opts.Type().UncoloredStringWithReplacements(opts.Replacements))
	} else {
		typePart = fmt.Sprintf("%-*s", typeWidth, opts.Type().UncoloredString())
	}

	typePart = color.New(opts.Type().Color).Sprint(typePart)

	// Build status part
	var statusText string
	if opts.Status().Text == "" {
		statusText = "no change"
	} else {
		statusText = opts.Status().Text
	}
	statusPart := fmt.Sprintf("%-*s", statusWidth, statusText)

	// Format the line with proper indentation
	return fmt.Sprintf("%s%s %s %s %s",
		strings.Repeat(" ", fileIndent),
		color.New(opts.Status().Style.SymbolColor).Sprint(string(opts.Status().Symbol)),
		namePart,
		typePart,
		statusPart)
}

func (l *Logger) formatArchiveTag(isArchive bool) string {
	if !isArchive {
		return ""
	}
	return fmt.Sprintf(" %s",
		color.New(color.FgCyan).Sprint("[archive]"))
}

func (l *Logger) formatRepoDisplay(repo RepoDisplay) {
	// Sort files by name
	sortedFiles := make([]FileInfo, len(repo.Files))
	copy(sortedFiles, repo.Files)
	sort.Slice(sortedFiles, func(i, j int) bool {
		return sortedFiles[i].Name < sortedFiles[j].Name
	})

	// Print destination header
	destPath := repo.Destination
	if repo.IsArchive {
		destPath = fmt.Sprintf("%s/%s", destPath, filepath.Base(repo.Name))
	}
	fmt.Fprintf(l.consoleOut, "[syncing %s]\n",
		color.New(color.FgCyan).Sprint(destPath))

	// Print repo header
	fmt.Fprintf(l.consoleOut, "%s %s %s %s\n",
		color.New(color.FgMagenta).Sprint("◆"),
		color.New(color.Bold).Sprint(repo.Name),
		color.New(color.Faint).Sprint("•"),
		color.New(color.FgYellow).Sprint(repo.Ref))

	// Print each file
	for _, file := range sortedFiles {
		fmt.Fprintln(l.consoleOut, l.formatFileOperation(file))
	}
}

func (l *Logger) Header(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	copyrcheaderText := color.New(color.Bold, color.FgCyan).Sprintf("copyrc")
	fmt.Fprintf(l.consoleOut, "\n%s %s\n\n", copyrcheaderText, color.New(color.Faint).Sprint("• syncing repository files"))
	l.zlog.Info().Msg(msg)
}

func (l *Logger) LogNewline() {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.consoleOut)
}

// func (l *Logger) Operation(msg string) {
// 	l.mu.Lock()
// 	defer l.mu.Unlock()

// 	// Check if this is a repository line
// 	if strings.Contains(msg, "Repository:") {
// 		// Parse repository line
// 		parts := strings.Split(msg, "(ref:")
// 		repo := strings.TrimPrefix(strings.TrimSpace(parts[0]), "Repository:")
// 		ref := strings.TrimSuffix(strings.TrimSpace(parts[1]), ")")
// 		isArchive := strings.Contains(ref, "[archive]")
// 		ref = strings.TrimSuffix(strings.TrimSpace(ref), "[archive]")
// 		ref = strings.TrimSuffix(strings.TrimSpace(ref), ")")

// 		// Extract destination path
// 		destPath := ""
// 		if idx := strings.Index(ref, "->"); idx != -1 {
// 			destPath = strings.TrimSpace(ref[idx+2:])
// 			ref = strings.TrimSpace(ref[:idx])
// 		}

// 		// Create RepoDisplay
// 		l.currentRepo =

// 		l.zlog.Info().
// 			Str("repo", l.currentRepo.Name).
// 			Str("ref", l.currentRepo.Ref).
// 			Bool("archive", l.currentRepo.IsArchive).
// 			Str("dest", l.currentRepo.Destination).
// 			Msg("Processing repository")
// 		return
// 	}

// 	panic("not implemented")

// 	// // This is a file operation
// 	// filename := strings.TrimPrefix(msg, "  → ")
// 	// parts := strings.Split(filename, " ")
// 	// filename = parts[0]

// 	// // Determine file status
// 	// var status FileStatus
// 	// isSpecial := strings.HasSuffix(filename, ".copyrc.lock")
// 	// isUntracked := strings.Contains(msg, "[untracked]")

// 	// if len(parts) > 1 {
// 	// 	switch parts[1] {
// 	// 	case "FileStatusNew":
// 	// 		status = NewFile
// 	// 	case "FileStatusUpdated":
// 	// 		status = UpdatedFile
// 	// 	default:
// 	// 		status = UnmodifiedCopyFile
// 	// 	}
// 	// } else {
// 	// 	status = UnmodifiedCopyFile
// 	// }

// 	// // Create FileInfo
// 	// op := FileInfo{
// 	// 	Name:        filename,
// 	// 	Status:      status,
// 	// 	IsManaged:   isSpecial,
// 	// 	IsUntracked: isUntracked,
// 	// }

// 	// // Add to current repo's files
// 	// if l.currentRepo != nil {
// 	// 	l.currentRepo.Files = append(l.currentRepo.Files, op)
// 	// }

// 	// fmt.Fprintln(l.consoleOut, l.formatFileStatus(op))
// 	// l.zlog.Info().
// 	// 	Str("file", op.Name).
// 	// 	Str("status", op.Status.Text).
// 	// 	Bool("special", op.IsManaged).
// 	// 	Bool("untracked", op.IsUntracked).
// 	// 	Msg("Processing file")
// }

func (l *Logger) Success(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.consoleOut, "✅ %s\n", color.New(color.FgGreen).Sprint(msg))
}

func (l *Logger) Warning(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.consoleOut, "⚠️  %s\n", color.New(color.FgYellow).Sprint(msg))
	l.zlog.Warn().Msg(msg)
}

func (l *Logger) Error(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.consoleOut, "❌ %s\n", color.New(color.FgRed).Sprint(msg))
	l.zlog.Error().Msg(msg)
}

func (l *Logger) Info(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.consoleOut, "ℹ️  %s\n", color.New(color.FgCyan).Sprint(msg))
	l.zlog.Info().Msg(msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.Warning(fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) Successf(format string, args ...interface{}) {
	l.Success(fmt.Sprintf(format, args...))
}

func (l *Logger) AddFileOperation(op FileInfo) {
	l.repoMu.Lock()
	defer l.repoMu.Unlock()

	// Add to current repo's files
	if l.currentRepo != nil {
		l.currentRepo.Files = append(l.currentRepo.Files, op)
	}

	fmt.Fprintln(l.consoleOut, l.formatFileOperation(op))
	l.zlog.Info().
		Str("file", op.Name).
		Str("status", op.Status().Text).
		Msg("Processing file")
}

func (l *Logger) LogFileOperation(opts FileInfo) {
	l.repoMu.Lock()
	defer l.repoMu.Unlock()

	// Add to current repo's files
	if l.currentRepo != nil {
		l.currentRepo.Files = append(l.currentRepo.Files, opts)
	}

	// Mark file as processed
	processedFiles.Store(opts.Name, true)

	// Format the line and add a newline
	line := l.formatFileOperation(opts) + "\n"
	fmt.Fprint(l.consoleOut, line)
	l.zlog.Info().
		Str("file", opts.Name).
		Str("status", opts.Status().Text).
		Str("type", opts.Type().UncoloredString()).
		Msg("Processing file")
}
