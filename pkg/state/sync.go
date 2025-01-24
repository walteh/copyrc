package state

import (
	"context"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/remote"
	"gitlab.com/tozd/go/errors"
)

func Sync(ctx context.Context, cn config.Config) error {
	zerolog.Ctx(ctx).Debug().Msg("syncing files from remote repositories")

	staticConfig := config.NewStaticConfig(ctx, cn)

	// Load state from disk
	stt, err := LoadState(ctx, cn)
	if err != nil {
		return errors.Errorf("loading state: %w", err)
	}

	if stt.file.Config == nil {
		return errors.New("state is empty")
	}

	zerolog.Ctx(ctx).Debug().Interface("state", staticConfig).Msg("state")
	// Check if state is consistent
	consistent, err := stt.IsConsistent(ctx)
	if err != nil {
		return errors.Errorf("checking state consistency: %w", err)
	}
	if !consistent {
		zerolog.Ctx(ctx).Warn().Msg("state is inconsistent, proceeding with sync")
	}

	// Check if config has changed
	stateHash := stt.ConfigHash()
	configHash := staticConfig.GetHash()

	if stateHash != configHash {
		zerolog.Ctx(ctx).Info().
			Str("state_hash", stateHash).
			Str("config_hash", configHash).
			Msg("config has changed")
	}

	// Process each repository
	for _, repo := range staticConfig.GetRepositories() {

		provider, err := remote.GetProviderFromConfig(ctx, repo)
		if err != nil {
			return errors.Errorf("getting provider: %w", err)
		}

		// Get repository from provider
		remoteRepo, err := provider.GetRepository(ctx, repo.Name)
		if err != nil {
			return errors.Errorf("getting repository %s: %w", repo.Name, err)
		}

		// Get repository name for validation
		repoName := remoteRepo.Name()
		if repoName != repo.Name {
			return errors.Errorf("unexpected repository name: got %s, want %s", repoName, repo.Name)
		}

		lrelease, err := remoteRepo.GetLatestRelease(ctx)
		if err != nil {
			return errors.Errorf("getting latest release: %w", err)
		}

		release, err := remoteRepo.GetReleaseFromRef(ctx, repo.Ref)
		if err != nil {
			return errors.Errorf("getting release %s: %w", repo.Ref, err)
		}

		lic, err := release.GetLicense(ctx)
		if err != nil {
			return errors.Errorf("getting license: %w", err)
		}

		stt.file.Repositories = append(stt.file.Repositories, Repository{
			Provider:  repo.Provider,
			Name:      repo.Name,
			LatestRef: lrelease.Ref(),
			Release: &Release{
				Ref:          release.Ref(),
				RefHash:      release.RefHash(),
				WebPermalink: release.WebPermalink(),
				License: &License{
					SPDX:            lic.SPDXID,
					RemotePermalink: lic.WebPermalink,
				},
				Archive: &Archive{},
			},
		})

		copies := config.FilterCopiesByRepository(staticConfig.GetCopies(), repo)

		// Find copies for this repository
		for _, copy := range copies {

			if copy.Options.SaveArchiveToPath != "" {
				return errors.New("save_archive_to_path is not supported yet")
			}

			if copy.Options.SaveArchiveToPath != "" {
				archive, err := release.GetTarball(ctx)
				if err != nil {
					return errors.Errorf("getting tarball: %w", err)
				}
				defer archive.Close()

				// hash, err := stt.hashContent(archive)
				// if err != nil {
				// 	return errors.Errorf("hashing content: %w", err)
				// }

				// _, err = stt.PutArchiveFile(ctx, ArchiveFile{
				// 	LocalPath:   copy.Paths.Local,
				// 	ContentHash: hash,
				// })
				// if err != nil {
				// 	return errors.Errorf("putting archive file: %w", err)
				// }

				// if copy.Options.CreateGoEmbedForArchive {
				// 	stt.PutGeneratedFile(ctx, GeneratedFile{
				// 		LocalPath:   copy.Paths.Local,
				// 		LastUpdated: time.Now(),
				// 		Content:     archive,
				// 	})
				// }
			}

			// List files at path
			files, err := release.ListFilesAtPath(ctx, copy.Paths.Remote)
			if err != nil {
				return errors.Errorf("listing files at %s: %w", copy.Paths.Remote, err)
			}

			isIgnored := func(file remote.RawTextFile) bool {
				for _, ignore := range copy.Options.IgnoreFilesGlobs {
					matched, err := doublestar.Match(ignore, file.Path())
					if err != nil {
						return false
					}
					if matched {
						return true
					}
				}
				return false
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
				if _, err := stt.PutRemoteTextFile(ctx, file, copy.Paths.Local, PutRemoteTextFileOptions{
					IsIgnored: isIgnored(file),
				}); err != nil {
					return errors.Errorf("putting file %s in state: %w", file.Path(), err)
				}
			}
		}
	}

	// Save state
	if err := stt.Save(ctx); err != nil {
		return errors.Errorf("saving state: %w", err)
	}

	return nil
}
