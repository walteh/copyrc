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
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/log"
	"github.com/walteh/copyrc/pkg/operation"
	"github.com/walteh/copyrc/pkg/provider"
	"github.com/walteh/copyrc/pkg/status"
	"gitlab.com/tozd/go/errors"
)

// 🎯 Handler holds the root command configuration
type Handler struct {
	// 📝 Configuration file path (required)
	configPath string

	// 🔧 Optional flags
	clean        bool // Whether to clean destination directory
	status       bool // Whether to check local status
	remoteStatus bool // Whether to check remote status
	force        bool // Whether to force update even if status is ok
	debug        bool // Enable debug logging
	trace        bool // Enable trace logging
}

// 🚀 NewRootCommand creates the root command for copyrc
func NewRootCommand() *cobra.Command {
	me := &Handler{}

	cmd := &cobra.Command{
		Use:   "copyrc",
		Short: "A tool for syncing repository files with local copies",
		Long: `copyrc is a tool for maintaining local copies of files from remote repositories.
It supports file replacements, status tracking, and various providers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return me.Run(cmd.Context())
		},
	}

	// 🎛️ Required flags
	cmd.Flags().StringVar(&me.configPath, "config", "", "path to config file (required)")
	cmd.MarkFlagRequired("config")

	// 🔧 Optional flags
	cmd.Flags().BoolVar(&me.clean, "clean", false, "clean destination directory")
	cmd.Flags().BoolVar(&me.status, "status", false, "check local status")
	cmd.Flags().BoolVar(&me.remoteStatus, "remote-status", false, "check remote status")
	cmd.Flags().BoolVar(&me.force, "force", false, "force update even if status is ok")
	cmd.Flags().BoolVar(&me.debug, "debug", false, "enable debug logging")
	cmd.Flags().BoolVar(&me.trace, "trace", false, "enable trace logging")

	return cmd
}

// 🏃 Run executes the root command
func (me *Handler) Run(ctx context.Context) error {
	// 📝 Setup logging
	var level zerolog.Level
	switch {
	case me.trace:
		level = zerolog.TraceLevel
	case me.debug:
		level = zerolog.DebugLevel
	default:
		level = zerolog.InfoLevel
	}

	logger := log.New(os.Stdout, level)
	ctx = log.NewContext(ctx, logger)

	// 📝 Load configuration
	cfg, err := config.Load(ctx, me.configPath)
	if err != nil {
		return errors.Errorf("loading config: %w", err)
	}

	// Override config with command line flags
	cfg.Clean = me.clean
	cfg.Status = me.status
	cfg.RemoteStatus = me.remoteStatus
	cfg.Force = me.force

	// 🔌 Get provider
	p, err := provider.GetFromRepo(ctx, cfg.Provider.Repo)
	if err != nil {
		return errors.Errorf("getting provider: %w", err)
	}

	// 📝 Initialize status manager
	statusPath := filepath.Join(cfg.Destination, ".copyrc.lock")
	statusMgr, err := status.New(statusPath)
	if err != nil {
		return errors.Errorf("initializing status manager: %w", err)
	}

	// 🧹 Clean if requested
	if cfg.Clean {
		if err := statusMgr.Clean(); err != nil {
			return errors.Errorf("cleaning status: %w", err)
		}
	}

	// 🔍 Check status if requested
	if cfg.Status || cfg.RemoteStatus {
		if err := statusMgr.CheckStatus(ctx, cfg, p); err != nil {
			if !cfg.Force {
				return errors.Errorf("status check failed: %w", err)
			}
			logger.Warning("status check failed, but continuing due to --force")
		}
	}

	// 🎯 Create operation manager
	opMgr := operation.New(cfg, p, logger)

	// 📂 List files
	files, err := p.ListFiles(ctx, cfg.Provider)
	if err != nil {
		return errors.Errorf("listing files: %w", err)
	}

	// 🚀 Start repository operation
	commitHash, err := p.GetCommitHash(ctx, cfg.Provider)
	if err != nil {
		return errors.Errorf("getting commit hash: %w", err)
	}

	logger.StartRepoOperation(ctx, log.RepoOperation{
		Name:        cfg.Provider.Repo,
		Ref:         cfg.Provider.Ref,
		Destination: cfg.Destination,
	})

	// 📝 Process files
	if err := opMgr.ProcessFiles(ctx, files); err != nil {
		return errors.Errorf("processing files: %w", err)
	}

	// 💾 Update status
	if err := statusMgr.UpdateCommitHash(ctx, commitHash); err != nil {
		return errors.Errorf("updating commit hash: %w", err)
	}

	if err := statusMgr.UpdateConfig(ctx, cfg); err != nil {
		return errors.Errorf("updating config: %w", err)
	}

	logger.EndRepoOperation(ctx)
	logger.Success("All files processed successfully")

	return nil
}

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
