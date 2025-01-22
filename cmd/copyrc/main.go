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
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/operation"
	"github.com/walteh/copyrc/pkg/provider"
	"github.com/walteh/copyrc/pkg/status"
	"gitlab.com/tozd/go/errors"
)

// 🎯 Handler handles the main command
type Handler struct {
	configFile   string
	trace        bool
	debug        bool
	clean        bool
	status       bool
	remoteStatus bool
	force        bool
	async        bool
}

// 🏃 Run runs the command
func (h *Handler) Run(ctx context.Context) error {
	// Set up logging
	level := zerolog.InfoLevel
	if h.trace {
		level = zerolog.TraceLevel
	} else if h.debug {
		level = zerolog.DebugLevel
	}
	logger := zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
	ctx = logger.WithContext(ctx)

	// Load config
	cfg, err := config.Load(ctx, h.configFile)
	if err != nil {
		return errors.Errorf("loading config: %w", err)
	}

	// Create status manager
	statusPath := h.configFile + ".lock"
	statusMgr, err := status.New(statusPath)
	if err != nil {
		return errors.Errorf("creating status manager: %w", err)
	}

	// Clean if requested
	if h.clean {
		if err := statusMgr.Clean(); err != nil {
			return errors.Errorf("cleaning status: %w", err)
		}
	}

	// Get provider
	factory := provider.Get("github")
	if factory == nil {
		return errors.Errorf("github provider not registered")
	}
	p, err := factory(ctx)
	if err != nil {
		return errors.Errorf("creating provider: %w", err)
	}

	// Check status if requested
	if h.status {
		if err := statusMgr.CheckStatus(ctx, cfg, p); err != nil {
			return errors.Errorf("checking status: %w", err)
		}
	}

	// Create operation manager
	mgr := operation.New(cfg, p, &logger)

	// List files
	files, err := p.ListFiles(ctx, cfg.Provider)
	if err != nil {
		return errors.Errorf("listing files: %w", err)
	}

	// Process files
	if err := mgr.ProcessFiles(ctx, files); err != nil {
		return errors.Errorf("processing files: %w", err)
	}

	// Update status
	commitHash, err := p.GetCommitHash(ctx, cfg.Provider)
	if err != nil {
		return errors.Errorf("getting commit hash: %w", err)
	}
	if err := statusMgr.UpdateCommitHash(ctx, commitHash); err != nil {
		return errors.Errorf("updating commit hash: %w", err)
	}
	if err := statusMgr.UpdateConfig(ctx, cfg); err != nil {
		return errors.Errorf("updating config: %w", err)
	}

	return nil
}

// 🎯 NewCommand creates the root command
func NewCommand() *cobra.Command {
	h := &Handler{}

	cmd := &cobra.Command{
		Use:   "copyrc",
		Short: "Copy files from a repository with replacements",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&h.configFile, "config", "c", ".copyrc.yaml", "Config file path")
	cmd.Flags().BoolVarP(&h.trace, "trace", "t", false, "Enable trace logging")
	cmd.Flags().BoolVarP(&h.debug, "debug", "d", false, "Enable debug logging")
	cmd.Flags().BoolVar(&h.clean, "clean", false, "Clean destination directory")
	cmd.Flags().BoolVar(&h.status, "status", false, "Check local status")
	cmd.Flags().BoolVar(&h.remoteStatus, "remote-status", false, "Check remote status")
	cmd.Flags().BoolVar(&h.force, "force", false, "Force update even if status is ok")
	cmd.Flags().BoolVar(&h.async, "async", false, "Process files asynchronously")

	return cmd
}

// 🎯 main is the entry point
func main() {
	if err := NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
