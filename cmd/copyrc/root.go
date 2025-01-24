package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/cmd/copyrc/opts"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/state"
	"gitlab.com/tozd/go/errors"
)

var (
	// Flags
	configFile string
	debug      bool
)

// newRootOpts creates a new rootOpts with initialized dependencies
func newRootOpts(ctx context.Context) (*opts.RootOpts, error) {
	// Create user logger
	userLogger := state.NewUserLogger(ctx)

	// Load config
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, errors.Errorf("loading config: %w", err)
	}

	// Create state manager
	stateManager, err := state.New(filepath.Dir(configFile))
	if err != nil {
		return nil, errors.Errorf("creating state manager: %w", err)
	}

	return &opts.RootOpts{
		Config:       cfg,
		StateManager: stateManager,
		UserLogger:   userLogger,
	}, nil
}

// addRootFlags adds shared flags to the root command
func addRootFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", ".copyrc.hcl", "config file path")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug logging")
}

// setupLogging configures zerolog based on flags
func setupLogging() {
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &log
}

// TODO(dr.methodical): 🧪 Add tests for config loading
// TODO(dr.methodical): 🧪 Add tests for state manager creation
// TODO(dr.methodical): 📝 Add examples of flag usage
