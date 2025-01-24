package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/cmd/copyrc-next/pkg/config"
	"github.com/walteh/copyrc/cmd/copyrc-next/pkg/state"
	"github.com/walteh/copyrc/cmd/copyrc/opts"
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
	cfg, err := config.LoadConfig(ctx, configFile)
	if err != nil {
		return nil, errors.Errorf("loading config: %w", err)
	}

	// Create state manager

	return &opts.RootOpts{
		Config:     cfg,
		UserLogger: userLogger,
	}, nil
}

// addRootFlags adds shared flags to the root command
func addRootFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", ".copyrc.hcl", "config file path")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug logging")
}

// setupLogging configures zerolog based on flags
func setupLogging() {
	// if debug {
	// 	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// } else {
	// 	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// }
	// log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	// zerolog.DefaultContextLogger = &log
}

// TODO(dr.methodical): 🧪 Add tests for config loading
// TODO(dr.methodical): 🧪 Add tests for state manager creation
// TODO(dr.methodical): 📝 Add examples of flag usage
