// Package operation provides core functionality for copying and validating files
package operation

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/remote"
	"github.com/walteh/copyrc/pkg/state"
	"gitlab.com/tozd/go/errors"
)

// 📋 CopyFiles copies files from remote repositories according to the provided config
func CopyFiles(ctx context.Context, cfg *config.CopyrcConfig, provider remote.Provider, st *state.State) error {
	logger := zerolog.Ctx(ctx)

	// Process each copy operation
	for _, copy := range cfg.Copies {
		// Get repository
		repo, err := provider.GetRepository(ctx, copy.Repository.Name)
		if err != nil {
			return errors.Errorf("getting repository %s: %w", copy.Repository.Name, err)
		}

		// Get release
		var release remote.Release
		if copy.Repository.Ref == "" {
			release, err = repo.GetLatestRelease(ctx)
		} else {
			release, err = repo.GetReleaseFromRef(ctx, copy.Repository.Ref)
		}
		if err != nil {
			return errors.Errorf("getting release for %s@%s: %w", copy.Repository.Name, copy.Repository.Ref, err)
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

			// Read content
			data, err := io.ReadAll(content)
			if err != nil {
				return errors.Errorf("reading content for %s: %w", file.Path(), err)
			}

			// Apply text replacements if any
			if len(copy.Options.TextReplacements) > 0 {
				contentStr := string(data)
				logger.Debug().
					Str("original", contentStr).
					Interface("replacements", copy.Options.TextReplacements).
					Msg("applying text replacements")
				for _, replacement := range copy.Options.TextReplacements {
					contentStr = strings.ReplaceAll(contentStr, replacement.FromText, replacement.ToText)
					logger.Debug().
						Str("from", replacement.FromText).
						Str("to", replacement.ToText).
						Str("result", contentStr).
						Msg("applied replacement")
				}
				data = []byte(contentStr)
			}

			// Determine local path
			localPath := filepath.Join(copy.Paths.Local, filepath.Base(file.Path()))
			if !strings.HasSuffix(localPath, ".copy.") && !strings.HasSuffix(localPath, ".patch.") {
				ext := filepath.Ext(localPath)
				localPath = localPath[:len(localPath)-len(ext)] + ".copy" + ext
			}
			localPath = filepath.Join(st.Dir(), localPath)
			logger.Debug().
				Str("state_dir", st.Dir()).
				Str("local_path", localPath).
				Str("file_path", file.Path()).
				Msg("determined local path")

			// Create directory if needed
			if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
				return errors.Errorf("creating directory for %s: %w", localPath, err)
			}

			// Write file
			logger.Debug().
				Str("path", localPath).
				Str("content", string(data)).
				Msg("writing file")
			if err := os.WriteFile(localPath, data, 0644); err != nil {
				return errors.Errorf("writing file %s: %w", localPath, err)
			}

			// Add to state
			_, err = st.PutRemoteTextFile(ctx, file, localPath)
			if err != nil {
				return errors.Errorf("adding file to state: %w", err)
			}

			logger.Info().
				Str("file", file.Path()).
				Str("local", localPath).
				Msg("copied file")
		}

		// Handle archive if requested
		if copy.Options.SaveArchiveToPath != "" {
			// Get tarball
			tarball, err := release.GetTarball(ctx)
			if err != nil {
				return errors.Errorf("getting tarball: %w", err)
			}
			defer tarball.Close()

			// Create directory if needed
			if err := os.MkdirAll(filepath.Dir(copy.Options.SaveArchiveToPath), 0755); err != nil {
				return errors.Errorf("creating directory for archive: %w", err)
			}

			// Create file
			f, err := os.Create(copy.Options.SaveArchiveToPath)
			if err != nil {
				return errors.Errorf("creating archive file: %w", err)
			}
			defer f.Close()

			// Copy content
			if _, err := io.Copy(f, tarball); err != nil {
				return errors.Errorf("writing archive: %w", err)
			}

			logger.Info().
				Str("path", copy.Options.SaveArchiveToPath).
				Msg("saved archive")
		}
	}

	return nil
}

// ✅ ValidateFiles checks that all copied files match their expected state
func ValidateFiles(ctx context.Context, st *state.State) error {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("starting validation")

	return st.ValidateLocalState(ctx)
}

// 🔍 CheckStatus checks if files need to be fetched from remote
func CheckStatus(ctx context.Context, cfg *config.CopyrcConfig, st *state.State) (bool, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("checking status")

	// First check if state is consistent
	consistent, err := st.IsConsistent(ctx)
	if err != nil {
		return false, errors.Errorf("checking state consistency: %w", err)
	}
	if !consistent {
		logger.Info().Msg("local state is inconsistent")
		return true, nil
	}

	// TODO(dr.methodical): 🔨 Compare config hash with last saved config
	// For now just return false since we haven't implemented config hashing
	return false, nil
}
