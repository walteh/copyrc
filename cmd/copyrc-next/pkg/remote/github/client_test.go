package github_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gh "github.com/walteh/copyrc/cmd/copyrc-next/pkg/remote/github"
)

func TestGitHubProvider(t *testing.T) {
	t.Run("test_get_repository_name_parsing", func(t *testing.T) {
		tests := []struct {
			name        string
			input       string
			wantOwner   string
			wantRepo    string
			wantErr     bool
			errContains string
		}{
			{
				name:      "valid_repository",
				input:     "walteh/copyrc",
				wantOwner: "walteh",
				wantRepo:  "copyrc",
				wantErr:   false,
			},
			{
				name:        "empty_name",
				input:       "",
				wantErr:     true,
				errContains: "empty repository name",
			},
			{
				name:        "missing_slash",
				input:       "waltehcopyrc",
				wantErr:     true,
				errContains: "invalid repository name",
			},
			{
				name:        "too_many_slashes",
				input:       "walteh/copyrc/extra",
				wantErr:     true,
				errContains: "invalid repository name",
			},
			{
				name:        "empty_owner",
				input:       "/copyrc",
				wantErr:     true,
				errContains: "invalid repository name",
			},
			{
				name:        "empty_repo",
				input:       "walteh/",
				wantErr:     true,
				errContains: "invalid repository name",
			},
			{
				name:      "with_whitespace",
				input:     " walteh / copyrc ",
				wantOwner: "walteh",
				wantRepo:  "copyrc",
				wantErr:   false,
			},
		}

		provider := gh.NewProvider()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, err := provider.GetRepository(context.Background(), tt.input)

				if tt.wantErr {
					require.Error(t, err, "GetRepository should error")
					assert.Contains(t, err.Error(), tt.errContains, "error message should match")
					return
				}

				require.NoError(t, err, "GetRepository should not error")
				ghRepo, ok := repo.(*gh.Repository)
				require.True(t, ok, "repository should be a *Repository")
				assert.Equal(t, tt.wantOwner, ghRepo.Owner(), "owner should match")
				assert.Equal(t, tt.wantRepo, ghRepo.Repo(), "repo should match")
			})
		}
	})

	t.Run("test_provider_name", func(t *testing.T) {
		provider := gh.NewProvider()
		assert.Equal(t, "github", provider.Name(), "provider should return correct name")
	})
}

// TODO(dr.methodical): 🧪 Add more edge case tests for:
//   - Network timeouts
//   - Invalid responses
//   - Rate limit handling
//   - Cache invalidation
// TODO(dr.methodical): 🔬 Add benchmarks for:
//   - Cache hit/miss scenarios
//   - Large repository operations
// TODO(dr.methodical): 📝 Add example tests for common use cases
