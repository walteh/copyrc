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

package log

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

// 🎨 Display configuration
const (
	fileIndent  = 4  // spaces to indent file entries
	nameWidth   = 35 // Base width for filename
	typeWidth   = 15 // Width for file type
	statusWidth = 15 // Width for status text
)

// 🎯 FileOperation represents a file operation for logging
type FileOperation struct {
	Path         string // File path
	Type         string // File type (managed/local/copy)
	Status       string // Operation status
	IsNew        bool   // Whether this is a new file
	IsModified   bool   // Whether the file was modified
	IsRemoved    bool   // Whether the file was removed
	IsManaged    bool   // Whether this is a managed file
	Replacements int    // Number of replacements made
}

// 📦 RepoOperation represents a repository operation for logging
type RepoOperation struct {
	Name        string // Repository name
	Ref         string // Repository ref
	Destination string // Destination path
	IsArchive   bool   // Whether this is an archive
}

// 🎯 Logger handles structured logging with console output
type Logger struct {
	zlog       zerolog.Logger
	console    io.Writer
	mu         sync.Mutex
	currentOp  *RepoOperation
	operations []FileOperation
}

// 🏭 New creates a new logger
func New(console io.Writer, level zerolog.Level) *Logger {
	zlog := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().Level(level)
	return &Logger{
		zlog:    zlog,
		console: console,
		mu:      sync.Mutex{},
	}
}

// 🔑 contextKey is the type for context values
type contextKey struct{}

// 🎯 FromContext gets the logger from context
func FromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(contextKey{}).(*Logger)
	if !ok {
		panic("logger not found in context")
	}
	return logger
}

// 🎯 NewContext adds the logger to context
func NewContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// 📝 formatFileOperation formats a file operation for display
func (l *Logger) formatFileOperation(op FileOperation) string {
	// Determine symbol and color
	var symbol rune
	var symbolColor color.Attribute
	switch {
	case op.IsRemoved:
		symbol = '✗'
		symbolColor = color.FgRed
	case op.IsNew:
		symbol = '✓'
		symbolColor = color.FgGreen
	case op.IsModified:
		symbol = '⟳'
		symbolColor = color.FgBlue
	default:
		if op.IsManaged {
			symbol = '•'
			symbolColor = color.FgCyan
		} else {
			symbol = '-'
			symbolColor = color.FgYellow
		}
	}

	// Format type with color
	var typeColor color.Attribute
	switch op.Type {
	case "managed":
		typeColor = color.FgCyan
	case "local":
		typeColor = color.FgYellow
	default:
		typeColor = color.FgBlue
	}

	// Build the line
	return fmt.Sprintf("%s%s %s %s %s",
		fmt.Sprintf("%*s", fileIndent, ""),
		color.New(symbolColor).Sprint(string(symbol)),
		fmt.Sprintf("%-*s", nameWidth, op.Path),
		color.New(typeColor).Sprint(fmt.Sprintf("%-*s", typeWidth, op.Type)),
		fmt.Sprintf("%-*s", statusWidth, op.Status))
}

// 📝 LogFileOperation logs a file operation
func (l *Logger) LogFileOperation(ctx context.Context, op FileOperation) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Add to operations list
	l.operations = append(l.operations, op)

	// Format and print
	fmt.Fprintln(l.console, l.formatFileOperation(op))

	// Log to zerolog
	l.zlog.Info().
		Str("file", op.Path).
		Str("type", op.Type).
		Str("status", op.Status).
		Bool("is_new", op.IsNew).
		Bool("is_modified", op.IsModified).
		Bool("is_removed", op.IsRemoved).
		Bool("is_managed", op.IsManaged).
		Int("replacements", op.Replacements).
		Msg("file operation")
}

// 📝 StartRepoOperation starts a new repository operation
func (l *Logger) StartRepoOperation(ctx context.Context, op RepoOperation) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.currentOp = &op
	l.operations = nil

	// Print repo header
	fmt.Fprintf(l.console, "[syncing %s]\n",
		color.New(color.FgCyan).Sprint(op.Destination))

	fmt.Fprintf(l.console, "%s %s %s %s\n",
		color.New(color.FgMagenta).Sprint("◆"),
		color.New(color.Bold).Sprint(op.Name),
		color.New(color.Faint).Sprint("•"),
		color.New(color.FgYellow).Sprint(op.Ref))

	// Log to zerolog
	l.zlog.Info().
		Str("repo", op.Name).
		Str("ref", op.Ref).
		Str("destination", op.Destination).
		Bool("is_archive", op.IsArchive).
		Msg("starting repository operation")
}

// 📝 EndRepoOperation ends the current repository operation
func (l *Logger) EndRepoOperation(ctx context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentOp == nil {
		return
	}

	// Log summary
	l.zlog.Info().
		Str("repo", l.currentOp.Name).
		Int("files", len(l.operations)).
		Msg("repository operation complete")

	l.currentOp = nil
	l.operations = nil
}

// 📝 LogNewline logs a newline
func (l *Logger) LogNewline() {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.console)
}

// 📝 Header logs a header
func (l *Logger) Header(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	copyrcText := color.New(color.Bold, color.FgCyan).Sprint("copyrc")
	fmt.Fprintf(l.console, "\n%s %s\n\n", copyrcText, color.New(color.Faint).Sprint("• "+msg))
	l.zlog.Info().Msg(msg)
}

// 📝 Success logs a success message
func (l *Logger) Success(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.console, "✅ %s\n", color.New(color.FgGreen).Sprint(msg))
	l.zlog.Info().Msg(msg)
}

// 📝 Warning logs a warning message
func (l *Logger) Warning(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.console, "⚠️  %s\n", color.New(color.FgYellow).Sprint(msg))
	l.zlog.Warn().Msg(msg)
}

// 📝 Error logs an error message
func (l *Logger) Error(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.console, "❌ %s\n", color.New(color.FgRed).Sprint(msg))
	l.zlog.Error().Msg(msg)
}

// 📝 Info logs an info message
func (l *Logger) Info(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.console, "ℹ️  %s\n", color.New(color.FgCyan).Sprint(msg))
	l.zlog.Info().Msg(msg)
}

// 📝 Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// 📝 Warningf logs a formatted warning message
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.Warning(fmt.Sprintf(format, args...))
}

// 📝 Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// 📝 Successf logs a formatted success message
func (l *Logger) Successf(format string, args ...interface{}) {
	l.Success(fmt.Sprintf(format, args...))
}
