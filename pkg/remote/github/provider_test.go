package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	t.Run("test_name", func(t *testing.T) {
		provider := NewProvider()
		assert.Equal(t, "github", provider.Name(), "provider should return correct name")
	})

	t.Run("test_get_repository", func(t *testing.T) {
		provider := NewProvider()
		repo, err := provider.GetRepository(context.Background(), "walteh/copyrc")
		require.NoError(t, err, "getting repository should not error")
		assert.Equal(t, "walteh/copyrc", repo.Name(), "repository should return correct name")

		// Check that the repository has the correct owner and repo
		ghRepo, ok := repo.(*Repository)
		require.True(t, ok, "repository should be a *Repository")
		assert.Equal(t, "walteh", ghRepo.owner, "repository should have correct owner")
		assert.Equal(t, "copyrc", ghRepo.repo, "repository should have correct repo")
	})

	t.Run("test_get_repository_invalid_name", func(t *testing.T) {
		provider := NewProvider()
		_, err := provider.GetRepository(context.Background(), "invalid")
		require.Error(t, err, "getting repository with invalid name should error")
		assert.Contains(t, err.Error(), "invalid repository name", "error should mention invalid name")
	})

	t.Run("test_get_repository_empty_name", func(t *testing.T) {
		provider := NewProvider()
		_, err := provider.GetRepository(context.Background(), "")
		require.Error(t, err, "getting repository with empty name should error")
		assert.Contains(t, err.Error(), "empty repository name", "error should mention empty name")
	})

	t.Run("test_get_repository_too_many_slashes", func(t *testing.T) {
		provider := NewProvider()
		_, err := provider.GetRepository(context.Background(), "a/b/c")
		require.Error(t, err, "getting repository with too many slashes should error")
		assert.Contains(t, err.Error(), "invalid repository name", "error should mention invalid name")
	})
}

// TODO(dr.methodical): 🧪 Add more edge case tests
// TODO(dr.methodical): 🔬 Add mock-based tests for error cases
// TODO(dr.methodical): �� Add example tests
