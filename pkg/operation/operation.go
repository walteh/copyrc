// Package operation provides core functionality for copying and validating files
package operation

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/remote"
	"github.com/walteh/copyrc/pkg/state"
	"gitlab.com/tozd/go/errors"
)

// 📋 CopyFiles copies files from remote repositories according to the provided config
func CopyFiles(ctx context.Context, cfg *config.CopyrcConfig, provider remote.Provider, st state.StateManager) error {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("copying files from remote repositories")

	// Load state from disk
	if err := st.Load(ctx); err != nil {
		return errors.Errorf("loading state: %w", err)
	}

	// Process each repository
	for _, repo := range cfg.Repositories {
		// Get repository from provider
		remoteRepo, err := provider.GetRepository(ctx, repo.Name)
		if err != nil {
			return errors.Errorf("getting repository %s: %w", repo.Name, err)
		}

		// Get repository details
		release, err := remoteRepo.GetLatestRelease(ctx)
		if err != nil {
			return errors.Errorf("getting latest release: %w", err)
		}

		// Find copies for this repository
		for _, copy := range cfg.Copies {
			if copy.Repository.Name != repo.Name {
				continue
			}

			// List files at path
			files, err := release.ListFilesAtPath(ctx, copy.Paths.Remote)
			if err != nil {
				return errors.Errorf("listing files: %w", err)
			}

			// Process each file
			for _, file := range files {
				// Get file content
				content, err := file.GetContent(ctx)
				if err != nil {
					return errors.Errorf("getting file content: %w", err)
				}
				defer content.Close()

				// Put file in state
				_, err = st.PutRemoteTextFile(ctx, file, copy.Paths.Local)
				if err != nil {
					return errors.Errorf("putting file in state: %w", err)
				}
			}
		}
	}

	// Validate local state
	if err := st.ValidateLocalState(ctx); err != nil {
		return errors.Errorf("validating local state: %w", err)
	}

	return nil
}

// ✅ ValidateFiles checks that all copied files match their expected state
func ValidateFiles(ctx context.Context, st state.StateManager) error {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("starting validation")

	return st.ValidateLocalState(ctx)
}

// 🔍 CheckStatus checks if files need to be fetched from remote
func CheckStatus(ctx context.Context, cfg *config.CopyrcConfig, st state.StateManager) (bool, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("checking status")

	// Load state from disk
	if err := st.Load(ctx); err != nil {
		return false, errors.Errorf("loading state: %w", err)
	}

	// Check if state is consistent
	isConsistent, err := st.IsConsistent(ctx)
	if err != nil {
		return false, errors.Errorf("checking consistency: %w", err)
	}

	if !isConsistent {
		logger.Debug().Msg("local state is inconsistent")
		return true, nil
	}

	logger.Debug().Msg("local state is consistent")
	return false, nil
}
