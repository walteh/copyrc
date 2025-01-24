package main

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/cmd/copyrc-next/pkg/state"
	"github.com/walteh/copyrc/cmd/copyrc/commands"
)

func main() {
	// Setup logging
	setupLogging()
	ctx := log.Logger.WithContext(context.Background())

	// Create user logger
	userLogger := state.NewUserLogger(ctx)

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "copyrc",
		Short: "A tool for copying and managing files from remote repositories",
		Long: `copyrc allows you to copy files from remote repositories (like GitHub)
and manage them locally with features like text replacement and state tracking.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// userLogger.LogStateChange("Starting copyrc...")
		},
	}

	// Add shared flags
	addRootFlags(rootCmd)

	// Create root options
	opts, err := newRootOpts(ctx)
	if err != nil {
		userLogger.LogValidation(false, "Failed to initialize", err)
		os.Exit(1)
	}

	// Add commands
	rootCmd.AddCommand(
		commands.NewSyncCmd(opts),
		commands.NewCleanCmd(opts),
		commands.NewStatusCmd(opts),
	)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		userLogger.LogValidation(false, "Command failed", err)
		os.Exit(1)
	}
}

// TODO(dr.methodical): 🧪 Add tests for command line parsing
// TODO(dr.methodical): 🧪 Add tests for context cancellation
// TODO(dr.methodical): 📝 Add examples in help text
