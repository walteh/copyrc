package github

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/gen/mockery"
)

func TestRepository(t *testing.T) {
	t.Run("test_name", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		repo := &Repository{
			provider: &Provider{
				client: mockClient,
			},
			owner: "walteh",
			repo:  "copyrc",
		}
		assert.Equal(t, "walteh/copyrc", repo.Name(), "repository should return correct name")
	})

	t.Run("test_get_latest_release", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		mockClient.EXPECT().GetLatestRelease(mock.Anything, "walteh", "copyrc").Return(
			&github.RepositoryRelease{
				TagName: github.String("v1.0.0"),
			},
			&github.Response{},
			nil,
		)

		repo := &Repository{
			provider: &Provider{
				client: mockClient,
			},
			owner: "walteh",
			repo:  "copyrc",
		}

		release, err := repo.GetLatestRelease(context.Background())
		require.NoError(t, err, "getting latest release should not error")
		assert.Equal(t, repo, release.Repository(), "release should reference parent repository")
		assert.Equal(t, "v1.0.0", release.Ref(), "release should have correct ref")
	})

	t.Run("test_get_latest_release_rate_limit", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		mockClient.EXPECT().GetLatestRelease(mock.Anything, "walteh", "copyrc").Return(
			nil,
			&github.Response{
				Response: &http.Response{
					StatusCode: http.StatusForbidden,
				},
			},
			&github.RateLimitError{
				Message: "API rate limit exceeded",
			},
		).Once()

		mockClient.EXPECT().GetLatestRelease(mock.Anything, "walteh", "copyrc").Return(
			&github.RepositoryRelease{
				TagName: github.String("v1.0.0"),
			},
			&github.Response{},
			nil,
		).Once()

		repo := &Repository{
			provider: &Provider{
				client: mockClient,
			},
			owner: "walteh",
			repo:  "copyrc",
		}

		// First call should hit rate limit
		_, err := repo.GetLatestRelease(context.Background())
		require.Error(t, err, "getting latest release should error due to rate limit")
		assert.Contains(t, err.Error(), "rate limit exceeded", "error should mention rate limit")

		// Second call should succeed
		release, err := repo.GetLatestRelease(context.Background())
		require.NoError(t, err, "getting latest release should not error")
		assert.Equal(t, repo, release.Repository(), "release should reference parent repository")
		assert.Equal(t, "v1.0.0", release.Ref(), "release should have correct ref")
	})

	t.Run("test_get_release_from_ref", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		mockClient.EXPECT().GetReleaseByTag(mock.Anything, "walteh", "copyrc", "v1.0.0").Return(
			&github.RepositoryRelease{
				TagName: github.String("v1.0.0"),
			},
			&github.Response{},
			nil,
		)

		repo := &Repository{
			provider: &Provider{
				client: mockClient,
			},
			owner: "walteh",
			repo:  "copyrc",
		}

		release, err := repo.GetReleaseFromRef(context.Background(), "v1.0.0")
		require.NoError(t, err, "getting release from ref should not error")
		assert.Equal(t, repo, release.Repository(), "release should reference parent repository")
		assert.Equal(t, "v1.0.0", release.Ref(), "release should have correct ref")
	})

	t.Run("test_get_release_from_ref_empty", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		repo := &Repository{
			provider: &Provider{
				client: mockClient,
			},
			owner: "walteh",
			repo:  "copyrc",
		}

		_, err := repo.GetReleaseFromRef(context.Background(), "")
		require.Error(t, err, "getting release with empty ref should error")
		assert.Contains(t, err.Error(), "empty ref", "error should mention empty ref")
	})

	t.Run("test_list_files_pagination", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		mockClient.EXPECT().GetContents(mock.Anything, "walteh", "copyrc", "pkg/remote", mock.Anything).Return(
			nil,
			[]*github.RepositoryContent{
				{
					Type: github.String("file"),
					Path: github.String("pkg/remote/interfaces.go"),
					SHA:  github.String("abc123"),
				},
				{
					Type: github.String("file"),
					Path: github.String("pkg/remote/github/client.go"),
					SHA:  github.String("def456"),
				},
			},
			&github.Response{},
			nil,
		)

		repo := &Repository{
			provider: &Provider{
				client: mockClient,
			},
			owner: "walteh",
			repo:  "copyrc",
		}

		release := &Release{
			repo:    repo,
			refName: "v1.0.0",
		}

		files, err := release.ListFilesAtPath(context.Background(), "pkg/remote")
		require.NoError(t, err, "listing files should not error")
		assert.True(t, len(files) > 0, "should find files across multiple pages")

		// Verify each file has required fields
		for _, file := range files {
			assert.NotEmpty(t, file.Path(), "file should have path")
			assert.NotEmpty(t, file.RawTextPermalink(), "file should have raw permalink")
			assert.NotEmpty(t, file.WebViewPermalink(), "file should have web permalink")
			assert.Equal(t, release, file.Release(), "file should reference parent release")
		}
	})

	t.Run("test_context_cancellation", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		mockClient.EXPECT().GetLatestRelease(mock.Anything, "walteh", "copyrc").Return(
			nil,
			nil,
			context.Canceled,
		)

		mockClient.EXPECT().GetReleaseByTag(mock.Anything, "walteh", "copyrc", "v1.0.0").Return(
			nil,
			nil,
			context.Canceled,
		)

		repo := &Repository{
			provider: &Provider{
				client: mockClient,
			},
			owner: "walteh",
			repo:  "copyrc",
		}

		// Create context that's already cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Test that all operations respect context cancellation
		_, err := repo.GetLatestRelease(ctx)
		require.Error(t, err, "operation should fail when context is cancelled")
		assert.Contains(t, err.Error(), "context canceled", "error should mention context cancellation")

		_, err = repo.GetReleaseFromRef(ctx, "v1.0.0")
		require.Error(t, err, "operation should fail when context is cancelled")
		assert.Contains(t, err.Error(), "context canceled", "error should mention context cancellation")
	})
}

// TODO(dr.methodical): 🧪 Add more edge case tests for:
//   - Network timeouts
//   - Invalid responses
//   - Malformed data
//   - Rate limit handling
//   - Cache invalidation
//   - Pagination edge cases
// TODO(dr.methodical): 🔬 Add benchmarks for:
//   - Cache hit/miss scenarios
//   - Large file downloads
//   - Pagination performance
// TODO(dr.methodical): 📝 Add example tests for common use cases
