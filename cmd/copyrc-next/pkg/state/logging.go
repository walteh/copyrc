package state

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/rs/zerolog"
)

func init() {
	// Enable debug output for development
	pterm.EnableDebugMessages()
}

// 📢 UserLogger provides user-friendly feedback about state changes
type UserLogger struct {
	log zerolog.Logger // for debug/error logging
}

// 🎨 FileChangeType represents the type of change made to a file
type FileChangeType int

const (
	FileAdded FileChangeType = iota
	FileUpdated
	FileDeleted
	FileSkipped
	FileError
)

// 🖼️ FileChange represents a change to a file in the state
type FileChange struct {
	Type        FileChangeType
	Path        string
	Description string
	Error       error
}

// 🎯 NewUserLogger creates a new user logger
func NewUserLogger(ctx context.Context) *UserLogger {
	return &UserLogger{
		log: *zerolog.Ctx(ctx),
	}
}

// 📝 LogFileChange logs a file change with appropriate emoji and formatting
func (u *UserLogger) LogFileChange(change FileChange) {
	// Get relative path for cleaner output
	relPath := filepath.Base(change.Path)

	var prefix, action string
	var printer *pterm.PrefixPrinter
	switch change.Type {
	case FileAdded:
		prefix = "✨"
		action = "Added"
		printer = pterm.Success.WithPrefix(pterm.Prefix{Text: prefix})
	case FileUpdated:
		prefix = "🔄"
		action = "Updated"
		printer = pterm.Info.WithPrefix(pterm.Prefix{Text: prefix})
	case FileDeleted:
		prefix = "🗑️"
		action = "Deleted"
		printer = pterm.Warning.WithPrefix(pterm.Prefix{Text: prefix})
	case FileSkipped:
		prefix = "⏭️"
		action = "Skipped"
		printer = pterm.Debug.WithPrefix(pterm.Prefix{Text: prefix})
	case FileError:
		prefix = "❌"
		action = "Error"
		printer = pterm.Error.WithPrefix(pterm.Prefix{Text: prefix})
	}

	msg := fmt.Sprintf("%s %s", action, relPath)
	if change.Description != "" {
		msg += fmt.Sprintf(" (%s)", change.Description)
	}

	if change.Error != nil {
		printer.Println(msg)
		pterm.Error.Println(change.Error)
		u.log.Error().Err(change.Error).Msg(msg) // Also log to zerolog for debugging
	} else {
		printer.Println(msg)
		u.log.Info().Msg(msg) // Also log to zerolog for debugging
	}
}

// 📊 LogStateChange logs a change to the overall state
func (u *UserLogger) LogStateChange(description string) {
	printer := pterm.Info.WithPrefix(pterm.Prefix{Text: "📦"})
	printer.Println(description)
	u.log.Info().Msg(description)
}

// 🔍 LogValidation logs validation results
func (u *UserLogger) LogValidation(valid bool, description string, err error) {
	if valid {
		pterm.Success.WithPrefix(pterm.Prefix{Text: "✅"}).Println(description)
		u.log.Info().Msg(description)
	} else {
		if err != nil {
			pterm.Error.WithPrefix(pterm.Prefix{Text: "❌"}).Println(description)
			pterm.Error.Println(err)
			u.log.Error().Err(err).Msg(description)
		} else {
			pterm.Warning.WithPrefix(pterm.Prefix{Text: "⚠️"}).Println(description)
			u.log.Warn().Msg(description)
		}
	}
}

// 🔒 LogLockOperation logs file locking operations
func (u *UserLogger) LogLockOperation(acquired bool, path string, err error) {
	if acquired {
		pterm.Debug.WithPrefix(pterm.Prefix{Text: "🔒"}).Printf("Acquired lock on %s\n", path)
		u.log.Debug().Msgf("Acquired lock on %s", path)
	} else {
		if err != nil {
			pterm.Error.WithPrefix(pterm.Prefix{Text: "🔓"}).Printf("Failed to acquire lock on %s\n", path)
			pterm.Error.Println(err)
			u.log.Error().Err(err).Msgf("Failed to acquire lock on %s", path)
		} else {
			pterm.Debug.WithPrefix(pterm.Prefix{Text: "🔓"}).Printf("Released lock on %s\n", path)
			u.log.Debug().Msgf("Released lock on %s", path)
		}
	}
}

// LogFileOperation logs a file operation with a custom message
func (l *UserLogger) LogFileOperation(ctx context.Context, operation string, path string) {
	zerolog.Ctx(ctx).Debug().
		Str("path", path).
		Msg(operation)
}
