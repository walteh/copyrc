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
	"flag"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/operation"
	"github.com/walteh/copyrc/pkg/provider"
	"github.com/walteh/copyrc/pkg/status"
)

var (
	configFile  = flag.String("config", ".copyrc.yaml", "Path to config file")
	destination = flag.String("destination", "", "Override destination path")
	async       = flag.Bool("async", false, "Run operations asynchronously")
	debug       = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	// Parse flags
	flag.Parse()

	// Set up logger
	logLevel := zerolog.InfoLevel
	if *debug {
		logLevel = zerolog.DebugLevel
	}
	logger := zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())

	// Load config
	cfg, err := config.Load(ctx, *configFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("loading config file")
	}

	// Override destination if provided
	if *destination != "" {
		cfg.Destination = *destination
	}

	// Create absolute destination path
	absDestination, err := filepath.Abs(cfg.Destination)
	if err != nil {
		logger.Fatal().Err(err).Msg("getting absolute destination path")
	}
	cfg.Destination = absDestination

	// Create provider
	p, err := provider.NewProvider(ctx, cfg.Provider)
	if err != nil {
		logger.Fatal().Err(err).Msg("creating provider")
	}

	// Create status manager
	statusMgr := status.NewManager(cfg.Destination, status.NewDefaultFileFormatter())

	// Create operation runner
	runner := operation.NewRunner(&logger, *async)

	// Create and run copy operation
	op := operation.NewCopyOperation(operation.Options{
		Config:    cfg,
		Provider:  p,
		StatusMgr: statusMgr,
		Logger:    &logger,
	})

	if err := runner.Run(ctx, op); err != nil {
		logger.Fatal().Err(err).Msg("running copy operation")
	}

	// Run clean operation if needed
	if cfg.Clean {
		cleanOp := operation.NewCleanOperation(operation.Options{
			Config:    cfg,
			Provider:  p,
			StatusMgr: statusMgr,
			Logger:    &logger,
		})

		if err := runner.Run(ctx, cleanOp); err != nil {
			logger.Fatal().Err(err).Msg("running clean operation")
		}
	}

	logger.Info().Msg("✅ All operations completed successfully")
}
