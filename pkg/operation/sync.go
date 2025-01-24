package operation

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/remote"
	"gitlab.com/tozd/go/errors"
)

// Sync implements Operator.Sync
func (o *operator) Sync(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("syncing files from remote repositories")

	// Load state from disk
	if err := o.state.Load(ctx); err != nil {
		return errors.Errorf("loading state: %w", err)
	}

	// Check if state is consistent
	consistent, err := o.state.IsConsistent(ctx)
	if err != nil {
		return errors.Errorf("checking state consistency: %w", err)
	}
	if !consistent {
		logger.Warn().Msg("state is inconsistent, proceeding with sync")
	}

	// Check if config has changed
	stateHash := o.state.ConfigHash()
	configHash := o.config.Hash()
	if stateHash != configHash {
		logger.Info().
			Str("state_hash", stateHash).
			Str("config_hash", configHash).
			Msg("config has changed")
	}

	// Process each repository
	for _, repo := range o.config.GetRepositories() {
		// Get repository from provider
		remoteRepo, err := o.provider.GetRepository(ctx, repo.Name)
		if err != nil {
			return errors.Errorf("getting repository %s: %w", repo.Name, err)
		}

		// Get repository name for validation
		repoName := remoteRepo.Name()
		if repoName != repo.Name {
			return errors.Errorf("unexpected repository name: got %s, want %s", repoName, repo.Name)
		}

		// Get release based on ref
		var release remote.Release
		if repo.Ref == "latest" {
			release, err = remoteRepo.GetLatestRelease(ctx)
		} else {
			release, err = remoteRepo.GetReleaseFromRef(ctx, repo.Ref)
		}
		if err != nil {
			return errors.Errorf("getting release %s: %w", repo.Ref, err)
		}

		// Find copies for this repository
		for _, copy := range o.config.GetCopies() {
			if copy.Repository.Name != repo.Name {
				continue
			}

			// List files at path
			files, err := release.ListFilesAtPath(ctx, copy.Paths.Remote)
			if err != nil {
				return errors.Errorf("listing files at %s: %w", copy.Paths.Remote, err)
			}

			// Process each file
			for _, file := range files {
				// Get file content
				content, err := file.GetContent(ctx)
				if err != nil {
					return errors.Errorf("getting content for %s: %w", file.Path(), err)
				}
				defer content.Close()

				// Put file in state - let state package handle file paths and suffixes
				if _, err := o.state.PutRemoteTextFile(ctx, file, copy.Paths.Local); err != nil {
					return errors.Errorf("putting file %s in state: %w", file.Path(), err)
				}
			}
		}
	}

	// Save state
	if err := o.state.Save(ctx); err != nil {
		return errors.Errorf("saving state: %w", err)
	}

	return nil
}

// TODO(dr.methodical): 🧪 Add tests for text replacement scenarios
// TODO(dr.methodical): 🧪 Add tests for multiple files
// TODO(dr.methodical): 🧪 Add tests for file cleanup
// TODO(dr.methodical): 🧪 Add tests for context cancellation
// TODO(dr.methodical): 🧪 Add benchmarks for large repositories
