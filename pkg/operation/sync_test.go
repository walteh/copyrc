package operation

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/gen/mockery"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/remote"
	statepkg "github.com/walteh/copyrc/pkg/state"
)

func TestSync(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(t *testing.T, cfg *mockery.MockConfig_config, provider *mockery.MockProvider_remote, state *mockery.MockStateManager_state)
		expectedError string
	}{
		{
			name: "successful_sync_single_file",
			setupMocks: func(t *testing.T, cfg *mockery.MockConfig_config, provider *mockery.MockProvider_remote, state *mockery.MockStateManager_state) {
				mockFile := mockery.NewMockRawTextFile_remote(t)
				mockRelease := mockery.NewMockRelease_remote(t)
				mockRepo := mockery.NewMockRepository_remote(t)

				// Create test config
				cfg.EXPECT().GetRepositories().Return([]config.RepositoryDefinition{
					{
						Provider: "github",
						Name:     "test/repo",
						Ref:      "main",
					},
				},
				)

				cfg.EXPECT().GetCopies().Return([]config.Copy{
					{
						Repository: config.RepositoryDefinition{
							Provider: "github",
							Name:     "test/repo",
							Ref:      "main",
						},
						Paths: config.CopyPaths{
							Remote: "remote/path",
							Local:  "test.copy.txt",
						},
					},
				},
				)

				// File expectations
				mockFile.EXPECT().GetContent(mock.Anything).Return(io.NopCloser(strings.NewReader("test content")), nil)
				mockFile.EXPECT().Path().Return("test.txt")

				// Release expectations
				mockRelease.EXPECT().ListFilesAtPath(mock.Anything, "remote/path").Return([]remote.RawTextFile{mockFile}, nil)

				// Repository expectations
				mockRepo.EXPECT().Name().Return("test/repo")
				mockRepo.EXPECT().GetReleaseFromRef(mock.Anything, "main").Return(mockRelease, nil)

				// Provider expectations
				provider.EXPECT().GetRepository(mock.Anything, "test/repo").Return(mockRepo, nil)

				// State expectations
				state.EXPECT().Load(mock.Anything).Return(nil)
				state.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				state.EXPECT().ConfigHash().Return("def456")
				state.EXPECT().PutRemoteTextFile(mock.Anything, mockFile, "test.copy.txt").Return(&statepkg.RemoteTextFile{
					LocalPath: "test.copy.txt",
					RepoName:  "test/repo",
				}, nil)
				state.EXPECT().Save(mock.Anything).Return(nil)
			},
		},
		{
			name: "load_state_error",
			setupMocks: func(t *testing.T, cfg *mockery.MockConfig_config, provider *mockery.MockProvider_remote, state *mockery.MockStateManager_state) {

				// Create test config
				cfg.EXPECT().GetRepositories().Return([]config.RepositoryDefinition{
					{
						Provider: "github",
						Name:     "test/repo",
						Ref:      "main",
					},
				})

				state.EXPECT().Load(mock.Anything).Return(assert.AnError)

			},
			expectedError: "loading state",
		},
		{
			name: "get_repository_error",
			setupMocks: func(t *testing.T, cfg *mockery.MockConfig_config, provider *mockery.MockProvider_remote, state *mockery.MockStateManager_state) {

				// Create test config

				cfg.EXPECT().GetRepositories().Return([]config.RepositoryDefinition{
					{
						Provider: "github",
						Name:     "test/repo",
						Ref:      "main",
					},
				})

				// State expectations
				state.EXPECT().Load(mock.Anything).Return(nil)
				state.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				state.EXPECT().ConfigHash().Return("def456")

				// Provider expectations
				provider.EXPECT().GetRepository(mock.Anything, "test/repo").Return(nil, assert.AnError)

			},
			expectedError: "getting repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := zerolog.New(zerolog.NewTestWriter(t)).WithContext(context.Background())
			cfg := mockery.NewMockConfig_config(t)
			provider := mockery.NewMockProvider_remote(t)
			state := mockery.NewMockStateManager_state(t)
			tt.setupMocks(t, cfg, provider, state)

			op, err := New(Options{
				Config:       cfg,
				StateManager: state,
				Provider:     provider,
			})
			require.NoError(t, err)

			// Execute
			err = op.Sync(ctx)

			// Verify
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			state.AssertExpectations(t)
			provider.AssertExpectations(t)
			cfg.AssertExpectations(t)
		})
	}
}

// TODO(dr.methodical): 🧪 Add tests for text replacement scenarios
// TODO(dr.methodical): 🧪 Add tests for multiple files
// TODO(dr.methodical): 🧪 Add tests for file cleanup
// TODO(dr.methodical): 🧪 Add tests for context cancellation
// TODO(dr.methodical): 🧪 Add benchmarks for large repositories
