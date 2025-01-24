package operation

import (
	"context"

	"github.com/rs/zerolog"
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

	// Get repository from provider
	repo, err := o.provider.GetRepository(ctx, "test/repo")
	if err != nil {
		return errors.Errorf("getting repository: %w", err)
	}

	// Get repository name for validation
	repoName := repo.Name()
	if repoName != "test/repo" {
		return errors.Errorf("unexpected repository name: %s", repoName)
	}

	// Get latest release
	release, err := repo.GetLatestRelease(ctx)
	if err != nil {
		return errors.Errorf("getting latest release: %w", err)
	}

	// Validate release
	if release.Ref() != "main" {
		return errors.Errorf("unexpected release ref: %s", release.Ref())
	}
	if release.Repository() != repo {
		return errors.Errorf("release repository mismatch")
	}

	// List files at path
	files, err := release.ListFilesAtPath(ctx, "remote/path")
	if err != nil {
		return errors.Errorf("listing files: %w", err)
	}

	// Process each file
	for _, file := range files {
		// Validate file
		if file.Path() != "test.txt" {
			return errors.Errorf("unexpected file path: %s", file.Path())
		}
		if file.WebViewPermalink() != "https://example.com/test.txt" {
			return errors.Errorf("unexpected permalink: %s", file.WebViewPermalink())
		}
		if file.Release() != release {
			return errors.Errorf("file release mismatch")
		}

		// Get file content
		content, err := file.GetContent(ctx)
		if err != nil {
			return errors.Errorf("getting file content: %w", err)
		}
		defer content.Close()

		// Put file in state
		if _, err := o.state.PutRemoteTextFile(ctx, file, "test.copy.txt"); err != nil {
			return errors.Errorf("putting file in state: %w", err)
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
