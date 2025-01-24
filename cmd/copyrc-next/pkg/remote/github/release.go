package github

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/google/go-github/v60/github"
	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/cmd/copyrc-next/pkg/remote"
	"gitlab.com/tozd/go/errors"
)

// Release implements the remote.Release interface for GitHub
type Release struct {
	repo    *Repository
	refName string
	release *github.RepositoryRelease

	tarballCache     []byte
	tarballCacheLock sync.RWMutex
}

// Repository returns the parent repository
func (r *Release) Repository() remote.Repository {
	return r.repo
}

// Ref returns the reference (tag/branch/commit) for this release
func (r *Release) Ref() string {
	return *r.release.TagName
}

// WebPermalink returns a permanent link to view the repository on the web
func (r *Release) WebPermalink() string {
	return r.release.GetHTMLURL()
}

// RefHash returns the hash of the reference for this repository
func (r *Release) RefHash() string {
	return *r.release.TargetCommitish
}

// GetTarball returns a reader for the tarball of this release
func (r *Release) GetTarball(ctx context.Context) (io.ReadCloser, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("repo", r.repo.Name()).Str("ref", r.refName).Msg("getting tarball")

	// Check cache first
	r.tarballCacheLock.RLock()
	if r.tarballCache != nil {
		defer r.tarballCacheLock.RUnlock()
		return io.NopCloser(bytes.NewReader(r.tarballCache)), nil
	}
	r.tarballCacheLock.RUnlock()

	// Cache miss, download tarball
	url := fmt.Sprintf("repos/%s/%s/tarball/%s", r.repo.owner, r.repo.repo, r.refName)
	rc, _, err := r.repo.provider.client.DownloadContents(ctx, r.repo.owner, r.repo.repo, url, nil)
	if err != nil {
		return nil, errors.Errorf("downloading tarball from GitHub: %w", err)
	}
	defer rc.Close()

	// Read and cache the tarball
	r.tarballCacheLock.Lock()
	defer r.tarballCacheLock.Unlock()

	// Double-check cache in case another goroutine filled it
	if r.tarballCache != nil {
		return io.NopCloser(bytes.NewReader(r.tarballCache)), nil
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return nil, errors.Errorf("reading tarball: %w", err)
	}

	r.tarballCache = buf.Bytes()
	return io.NopCloser(bytes.NewReader(r.tarballCache)), nil
}

// ListFilesAtPath lists all files at a given path in this release
func (r *Release) ListFilesAtPath(ctx context.Context, path string) ([]remote.RawTextFile, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("repo", r.repo.Name()).Str("ref", r.refName).Str("path", path).Msg("listing files")

	_, contents, _, err := r.repo.provider.client.GetContents(ctx, r.repo.owner, r.repo.repo, path, &github.RepositoryContentGetOptions{
		Ref: r.refName,
	})
	if err != nil {
		return nil, errors.Errorf("listing files from GitHub: %w", err)
	}

	var files []remote.RawTextFile
	for _, content := range contents {
		if content.GetType() == "file" {
			files = append(files, &RawTextFile{
				parentRelease: r,
				filePath:      content.GetPath(),
				fileSHA:       content.GetSHA(),
			})
		}
	}

	return files, nil
}

// GetFileAtPath returns a specific file from this release
func (r *Release) GetFileAtPath(ctx context.Context, path string) (remote.RawTextFile, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("repo", r.repo.Name()).Str("ref", r.refName).Str("path", path).Msg("getting file")

	content, _, _, err := r.repo.provider.client.GetContents(ctx, r.repo.owner, r.repo.repo, path, &github.RepositoryContentGetOptions{
		Ref: r.refName,
	})
	if err != nil {
		return nil, errors.Errorf("getting file from GitHub: %w", err)
	}

	return &RawTextFile{
		parentRelease: r,
		filePath:      content.GetPath(),
		fileSHA:       content.GetSHA(),
	}, nil
}

// GetLicense returns the license file for this release
func (r *Release) GetLicense(ctx context.Context) (remote.License, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("repo", r.repo.Name()).Str("ref", r.refName).Msg("getting license")

	license, _, err := r.repo.provider.client.License(ctx, r.repo.owner, r.repo.repo)
	if err != nil {
		return remote.License{}, errors.Errorf("getting license from GitHub: %w", err)
	}

	return remote.License{
		SPDXID:       license.GetLicense().GetSPDXID(),
		WebPermalink: license.GetHTMLURL(),
	}, nil
}

// GetLicenseAtPath returns a license file at a specific path
func (r *Release) GetLicenseAtPath(ctx context.Context, path string) (remote.License, error) {
	return r.GetLicense(ctx)
}

// RawTextFile implements the remote.RawTextFile interface for GitHub
type RawTextFile struct {
	parentRelease *Release
	filePath      string
	fileSHA       string
}

// Release returns the parent release
func (f *RawTextFile) Release() remote.Release {
	return f.parentRelease
}

// RawTextPermalink returns a permanent link to the raw text content
func (f *RawTextFile) RawTextPermalink() string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		f.parentRelease.repo.owner,
		f.parentRelease.repo.repo,
		f.parentRelease.refName,
		f.filePath,
	)
}

// GetContent returns a reader for the file content
func (f *RawTextFile) GetContent(ctx context.Context) (io.ReadCloser, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("repo", f.parentRelease.repo.Name()).Str("ref", f.parentRelease.refName).Str("path", f.filePath).Msg("getting file content")

	content, _, _, err := f.parentRelease.repo.provider.client.GetContents(ctx, f.parentRelease.repo.owner, f.parentRelease.repo.repo, f.filePath, &github.RepositoryContentGetOptions{
		Ref: f.parentRelease.refName,
	})
	if err != nil {
		return nil, errors.Errorf("getting file content from GitHub: %w", err)
	}

	decodedContent, err := content.GetContent()
	if err != nil {
		return nil, errors.Errorf("decoding file content: %w", err)
	}

	return io.NopCloser(strings.NewReader(decodedContent)), nil
}

// Path returns the path of the file in the repository
func (f *RawTextFile) Path() string {
	return f.filePath
}

// WebViewPermalink returns a permanent link to view the file on the web
func (f *RawTextFile) WebViewPermalink() string {
	return fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s",
		f.parentRelease.repo.owner,
		f.parentRelease.repo.repo,
		f.parentRelease.refName,
		f.filePath,
	)
}

// TODO(dr.methodical): 🧪 Add tests for Release methods
// TODO(dr.methodical): 🧪 Add tests for RawTextFile methods
// TODO(dr.methodical): 📝 Add examples
