package github

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/gen/mockery"
)

func TestRelease(t *testing.T) {
	t.Run("test_repository_returns_parent", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
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

		assert.Equal(t, repo, release.Repository(), "release should return parent repository")
		assert.Equal(t, "walteh/copyrc", release.Repository().Name(), "repository should return correct name")
	})

	t.Run("test_ref_returns_tag", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
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

		assert.Equal(t, "v1.0.0", release.Ref(), "release should return correct ref")
	})

	t.Run("test_get_tarball", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		mockClient.EXPECT().DownloadContents(mock.Anything, "walteh", "copyrc", "repos/walteh/copyrc/tarball/v1.0.0", mock.Anything).Return(
			io.NopCloser(strings.NewReader("mock tarball content that looks like a tar")),
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

		rc, err := release.GetTarball(context.Background())
		require.NoError(t, err, "getting tarball should not error")
		defer rc.Close()

		data, err := io.ReadAll(rc)
		require.NoError(t, err, "reading tarball should not error")
		assert.Equal(t, "mock tarball content that looks like a tar", string(data), "tarball content should match")
		mockClient.AssertExpectations(t)
	})

	t.Run("test_list_files_from_tarball", func(t *testing.T) {
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

		// Verify each file has required fields
		for _, file := range files {
			assert.NotEmpty(t, file.Path(), "file should have path")
			assert.True(t, strings.HasPrefix(file.Path(), "pkg/remote/"), "file path should be under requested directory")
			assert.NotEmpty(t, file.RawTextPermalink(), "file should have raw permalink")
			assert.NotEmpty(t, file.WebViewPermalink(), "file should have web permalink")
			assert.Equal(t, release, file.Release(), "file should reference parent release")
		}
		mockClient.AssertExpectations(t)
	})

	t.Run("test_get_file_from_tarball", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		mockClient.EXPECT().GetContents(mock.Anything, "walteh", "copyrc", "README.md", mock.Anything).Return(
			&github.RepositoryContent{
				Path:    github.String("README.md"),
				SHA:     github.String("abc123"),
				Content: github.String("mock file content"),
			},
			nil,
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

		file, err := release.GetFileAtPath(context.Background(), "README.md")
		require.NoError(t, err, "getting file should not error")

		content, err := file.GetContent(context.Background())
		require.NoError(t, err, "getting content should not error")
		defer content.Close()

		data, err := io.ReadAll(content)
		require.NoError(t, err, "reading content should not error")
		assert.Equal(t, "mock file content", string(data), "file content should match")
		mockClient.AssertExpectations(t)
	})

	t.Run("test_get_license_from_tarball", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		mockClient.EXPECT().License(mock.Anything, "walteh", "copyrc").Return(
			&github.RepositoryLicense{
				Content: github.String("mock license content"),
				License: &github.License{
					SPDXID: github.String("MIT"),
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

		license, err := release.GetLicense(context.Background())
		require.NoError(t, err, "getting license should not error")

		assert.Equal(t, "MIT", license.SPDXID, "SPDX ID should match")
		assert.Equal(t, "https://github.com/walteh/copyrc/blob/v1.0.0/LICENSE", license.WebPermalink, "Web permalink should match")
		mockClient.AssertExpectations(t)
	})

	t.Run("test_tarball_caching", func(t *testing.T) {
		mockClient := mockery.NewMockGitHubClient_github(t)
		mockClient.EXPECT().DownloadContents(mock.Anything, "walteh", "copyrc", "repos/walteh/copyrc/tarball/v1.0.0", mock.Anything).Return(
			io.NopCloser(strings.NewReader("mock tarball content")),
			&github.Response{},
			nil,
		).Once() // Should only be called once due to caching

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

		// First call should download tarball
		rc1, err := release.GetTarball(context.Background())
		require.NoError(t, err, "getting tarball first time should not error")
		defer rc1.Close()

		// Second call should use cached tarball
		rc2, err := release.GetTarball(context.Background())
		require.NoError(t, err, "getting tarball second time should not error")
		defer rc2.Close()

		// Both should return same content
		data1, err := io.ReadAll(rc1)
		require.NoError(t, err, "reading first tarball should not error")
		data2, err := io.ReadAll(rc2)
		require.NoError(t, err, "reading second tarball should not error")
		assert.Equal(t, data1, data2, "cached tarball should match original")
		mockClient.AssertExpectations(t)
	})
}

// TODO(dr.methodical): 🧪 Add more edge case tests for:
//   - Network timeouts during tarball download
//   - Invalid/corrupted tarballs
//   - Missing files in tarball
//   - Cache invalidation scenarios
// TODO(dr.methodical): 🔬 Add benchmarks for:
//   - Tarball download and extraction
//   - File lookup performance
//   - Cache hit/miss scenarios
// TODO(dr.methodical): 📝 Add example tests for common use cases
