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
	"github.com/walteh/copyrc/pkg/remote"
	"github.com/walteh/copyrc/pkg/state"
)

func TestSync(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*testing.T) (*mockConfig, remote.Provider, *mockery.MockStateManager_state)
		expectedError string
	}{
		{
			name: "successful_sync_single_file",
			setupMocks: func(t *testing.T) (*mockConfig, remote.Provider, *mockery.MockStateManager_state) {
				mockFile := mockery.NewMockRawTextFile_remote(t)
				mockRelease := mockery.NewMockRelease_remote(t)
				mockRepo := mockery.NewMockRepository_remote(t)
				mockProvider := mockery.NewMockProvider_remote(t)
				mockState := mockery.NewMockStateManager_state(t)
				cfg := &mockConfig{}

				// Config expectations
				cfg.On("Hash").Return("abc123")

				// File expectations
				mockFile.EXPECT().GetContent(mock.Anything).Return(io.NopCloser(strings.NewReader("test content")), nil)
				mockFile.EXPECT().Path().Return("test.txt")
				mockFile.EXPECT().WebViewPermalink().Return("https://example.com/test.txt")
				mockFile.EXPECT().Release().Return(mockRelease)

				// Release expectations
				mockRelease.EXPECT().Ref().Return("main")
				mockRelease.EXPECT().Repository().Return(mockRepo)
				mockRelease.EXPECT().ListFilesAtPath(mock.Anything, "remote/path").Return([]remote.RawTextFile{mockFile}, nil)

				// Repository expectations
				mockRepo.EXPECT().Name().Return("test/repo")
				mockRepo.EXPECT().GetLatestRelease(mock.Anything).Return(mockRelease, nil)

				// Provider expectations
				mockProvider.EXPECT().GetRepository(mock.Anything, "test/repo").Return(mockRepo, nil)

				// State expectations
				mockState.EXPECT().Load(mock.Anything).Return(nil)
				mockState.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				mockState.EXPECT().ConfigHash().Return("def456")
				mockState.EXPECT().PutRemoteTextFile(mock.Anything, mockFile, "test.copy.txt").Return(&state.RemoteTextFile{
					LocalPath: "test.copy.txt",
					RepoName:  "test/repo",
				}, nil)
				mockState.EXPECT().Save(mock.Anything).Return(nil)

				return cfg, mockProvider, mockState
			},
		},
		{
			name: "load_state_error",
			setupMocks: func(t *testing.T) (*mockConfig, remote.Provider, *mockery.MockStateManager_state) {
				mockProvider := mockery.NewMockProvider_remote(t)
				mockState := mockery.NewMockStateManager_state(t)
				cfg := &mockConfig{}

				mockState.EXPECT().Load(mock.Anything).Return(assert.AnError)

				return cfg, mockProvider, mockState
			},
			expectedError: "loading state",
		},
		{
			name: "get_repository_error",
			setupMocks: func(t *testing.T) (*mockConfig, remote.Provider, *mockery.MockStateManager_state) {
				mockProvider := mockery.NewMockProvider_remote(t)
				mockState := mockery.NewMockStateManager_state(t)
				cfg := &mockConfig{}

				// Config expectations
				cfg.On("Hash").Return("abc123")

				// State expectations
				mockState.EXPECT().Load(mock.Anything).Return(nil)
				mockState.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				mockState.EXPECT().ConfigHash().Return("def456")

				// Provider expectations
				mockProvider.EXPECT().GetRepository(mock.Anything, mock.Anything).Return(nil, assert.AnError)

				return cfg, mockProvider, mockState
			},
			expectedError: "getting repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := zerolog.New(zerolog.NewTestWriter(t)).WithContext(context.Background())
			cfg, provider, mockState := tt.setupMocks(t)

			op, err := New(Options{
				Config:       cfg,
				StateManager: mockState,
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
		})
	}
}

// TODO(dr.methodical): 🧪 Add tests for text replacement scenarios
// TODO(dr.methodical): 🧪 Add tests for multiple files
// TODO(dr.methodical): 🧪 Add tests for file cleanup
// TODO(dr.methodical): 🧪 Add tests for context cancellation
// TODO(dr.methodical): 🧪 Add benchmarks for large repositories
