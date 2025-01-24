package github

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/cmd/copyrc-next/pkg/remote"
	"gitlab.com/tozd/go/errors"
)

// GitHubClient defines the interface for GitHub API operations we need
type GitHubClient interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
	DownloadContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
	License(ctx context.Context, owner, repo string) (*github.RepositoryLicense, *github.Response, error)
}

// Provider implements the remote.Provider interface for GitHub
type Provider struct {
	client GitHubClient
}

func init() {
	remote.RegisterProvider("github", NewProvider())
}

// NewProvider creates a new GitHub provider
func NewProvider() *Provider {
	client := github.NewClient(nil)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		client = client.WithAuthToken(token)
	}
	return &Provider{
		client: &githubClientWrapper{client: client},
	}
}

// githubClientWrapper wraps the GitHub client to implement our interface
type githubClientWrapper struct {
	client *github.Client
}

func (w *githubClientWrapper) GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
	return w.client.Repositories.GetLatestRelease(ctx, owner, repo)
}

func (w *githubClientWrapper) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error) {
	return w.client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
}

func (w *githubClientWrapper) DownloadContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error) {
	return w.client.Repositories.DownloadContents(ctx, owner, repo, path, opts)
}

func (w *githubClientWrapper) GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	return w.client.Repositories.GetContents(ctx, owner, repo, path, opts)
}

func (w *githubClientWrapper) License(ctx context.Context, owner, repo string) (*github.RepositoryLicense, *github.Response, error) {
	return w.client.Repositories.License(ctx, owner, repo)
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return "github"
}

// GetRepository returns a Repository interface for the given repository name
func (p *Provider) GetRepository(ctx context.Context, name string) (remote.Repository, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("name", name).Msg("getting repository")

	if name == "" {
		return nil, errors.Errorf("empty repository name")
	}

	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		return nil, errors.Errorf("invalid repository name: %s", name)
	}

	owner := strings.TrimSpace(parts[0])
	repo := strings.TrimSpace(parts[1])

	if owner == "" || repo == "" {
		return nil, errors.Errorf("invalid repository name: %s", name)
	}

	return &Repository{
		provider: p,
		owner:    owner,
		repo:     repo,
	}, nil
}

// Repository implements the remote.Repository interface for GitHub
type Repository struct {
	provider *Provider
	owner    string
	repo     string
}

// Name returns the name of the repository
func (r *Repository) Name() string {
	return fmt.Sprintf("%s/%s", r.owner, r.repo)
}

// Owner returns the owner part of the repository name
func (r *Repository) Owner() string {
	return r.owner
}

// Repo returns the repo part of the repository name
func (r *Repository) Repo() string {
	return r.repo
}

// GetLatestRelease returns the latest release for this repository
func (r *Repository) GetLatestRelease(ctx context.Context) (remote.Release, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("repo", r.Name()).Msg("getting latest release")

	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, errors.Errorf("context error: %w", err)
	}

	release, resp, err := r.provider.client.GetLatestRelease(ctx, r.owner, r.repo)
	if err != nil {
		if ctx.Err() != nil {
			return nil, errors.Errorf("context error: %w", ctx.Err())
		}
		if resp != nil && resp.StatusCode == 403 {
			if _, ok := err.(*github.RateLimitError); ok {
				return nil, errors.Errorf("rate limit exceeded: %w", err)
			}
		}
		return nil, errors.Errorf("getting latest release from GitHub: %w", err)
	}

	return &Release{
		repo:    r,
		refName: release.GetTagName(),
		release: release,
	}, nil
}

// GetReleaseFromRef returns a specific release for this repository
func (r *Repository) GetReleaseFromRef(ctx context.Context, ref string) (remote.Release, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("repo", r.Name()).Str("ref", ref).Msg("getting release from ref")

	if ref == "" {
		return nil, errors.Errorf("empty ref")
	}

	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, errors.Errorf("context error: %w", err)
	}

	// Try to get release by tag name
	release, _, err := r.provider.client.GetReleaseByTag(ctx, r.owner, r.repo, ref)
	if err != nil {
		if ctx.Err() != nil {
			return nil, errors.Errorf("context error: %w", ctx.Err())
		}
		// If not found, create a pseudo-release for the ref
		return &Release{
			repo:    r,
			refName: ref,
			release: nil, // We'll handle nil release in the Release methods
		}, nil
	}

	return &Release{
		repo:    r,
		refName: ref,
		release: release,
	}, nil
}

// TODO(dr.methodical): 🔬 Implement Release and RawTextFile types
// TODO(dr.methodical): 🧪 Add tests for Repository methods
// TODO(dr.methodical): 📝 Add examples
