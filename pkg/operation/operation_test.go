package operation

import (
	"context"
	"errors"
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
	"github.com/walteh/copyrc/pkg/state"
)

func setupTestLogger(t *testing.T) context.Context {
	logger := zerolog.New(zerolog.TestWriter{T: t}).With().Timestamp().Logger()
	return logger.WithContext(context.Background())
}

// TestCopyFiles tests the CopyFiles function
func TestCopyFiles(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*config.CopyrcConfig, remote.Provider, *mockery.MockStateManager_state)
		wantErr bool
	}{
		{
			name: "copies_single_file",
			setup: func(t *testing.T) (*config.CopyrcConfig, remote.Provider, *mockery.MockStateManager_state) {
				mockFile := mockery.NewMockRawTextFile_remote(t)
				mockRelease := mockery.NewMockRelease_remote(t)
				mockRepo := mockery.NewMockRepository_remote(t)
				mockProvider := mockery.NewMockProvider_remote(t)
				mockState := mockery.NewMockStateManager_state(t)

				// Setup basic expectations
				mockFile.EXPECT().GetContent(mock.Anything).Return(io.NopCloser(strings.NewReader("old content")), nil).Maybe()
				mockFile.EXPECT().Path().Return("test.txt").Maybe()
				mockFile.EXPECT().WebViewPermalink().Return("web_link").Maybe()
				mockFile.EXPECT().Release().Return(mockRelease).Maybe()
				mockRelease.EXPECT().Ref().Return("ref").Maybe()
				mockRelease.EXPECT().Repository().Return(mockRepo).Maybe()
				mockRepo.EXPECT().Name().Return("repo").Maybe()
				mockProvider.EXPECT().GetRepository(mock.Anything, "repo").Return(mockRepo, nil).Maybe()
				mockRepo.EXPECT().GetLatestRelease(mock.Anything).Return(mockRelease, nil).Maybe()
				mockRelease.EXPECT().ListFilesAtPath(mock.Anything, "remote/path").Return([]remote.RawTextFile{mockFile}, nil).Maybe()

				// Setup state expectations
				mockState.EXPECT().Load(mock.Anything).Return(nil).Maybe()
				mockState.EXPECT().PutRemoteTextFile(mock.Anything, mockFile, "test.copy.txt").Return(&state.RemoteTextFile{
					LocalPath: "test.copy.txt",
					RepoName:  "repo",
				}, nil).Maybe()
				mockState.EXPECT().ValidateLocalState(mock.Anything).Return(nil).Maybe()
				mockState.EXPECT().IsConsistent(mock.Anything).Return(true, nil).Maybe()

				cfg := &config.CopyrcConfig{
					Repositories: []config.RepositoryDefinition{
						{
							Provider: "github",
							Name:     "repo",
							Ref:      "main",
						},
					},
					Copies: []config.Copy{
						{
							Repository: config.RepositoryDefinition{
								Provider: "github",
								Name:     "repo",
								Ref:      "main",
							},
							Paths: config.CopyPaths{
								Remote: "remote/path",
								Local:  "test.copy.txt",
							},
						},
					},
				}

				return cfg, mockProvider, mockState
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, provider, mockState := tt.setup(t)
			err := CopyFiles(context.Background(), cfg, provider, mockState)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateFiles(t *testing.T) {
	ctx := setupTestLogger(t)

	t.Run("validates_consistent_state", func(t *testing.T) {
		mockState := mockery.NewMockStateManager_state(t)
		mockState.EXPECT().ValidateLocalState(mock.Anything).Return(nil).Maybe()

		err := ValidateFiles(ctx, mockState)
		require.NoError(t, err)
	})

	t.Run("detects_inconsistent_state", func(t *testing.T) {
		mockState := mockery.NewMockStateManager_state(t)
		mockState.EXPECT().ValidateLocalState(mock.Anything).Return(errors.New("hash mismatch")).Maybe()

		err := ValidateFiles(ctx, mockState)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hash mismatch")
	})
}

func TestCheckStatus(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) (*config.CopyrcConfig, *mockery.MockStateManager_state)
		wantBool bool
		wantErr  bool
	}{
		{
			name: "inconsistent_state_returns_true",
			setup: func(t *testing.T) (*config.CopyrcConfig, *mockery.MockStateManager_state) {
				mockState := mockery.NewMockStateManager_state(t)
				mockState.EXPECT().Load(mock.Anything).Return(nil).Maybe()
				mockState.EXPECT().IsConsistent(mock.Anything).Return(false, nil).Maybe()

				cfg := &config.CopyrcConfig{
					Repositories: []config.RepositoryDefinition{
						{
							Provider: "github",
							Name:     "repo",
							Ref:      "main",
						},
					},
					Copies: []config.Copy{
						{
							Repository: config.RepositoryDefinition{
								Provider: "github",
								Name:     "repo",
								Ref:      "main",
							},
							Paths: config.CopyPaths{
								Remote: "remote/path",
								Local:  "test.copy.txt",
							},
						},
					},
				}

				return cfg, mockState
			},
			wantBool: true,
			wantErr:  false,
		},
		{
			name: "consistent_state_returns_false",
			setup: func(t *testing.T) (*config.CopyrcConfig, *mockery.MockStateManager_state) {
				mockState := mockery.NewMockStateManager_state(t)
				mockState.EXPECT().Load(mock.Anything).Return(nil).Maybe()
				mockState.EXPECT().IsConsistent(mock.Anything).Return(true, nil).Maybe()

				cfg := &config.CopyrcConfig{
					Repositories: []config.RepositoryDefinition{
						{
							Provider: "github",
							Name:     "repo",
							Ref:      "main",
						},
					},
					Copies: []config.Copy{
						{
							Repository: config.RepositoryDefinition{
								Provider: "github",
								Name:     "repo",
								Ref:      "main",
							},
							Paths: config.CopyPaths{
								Remote: "remote/path",
								Local:  "test.copy.txt",
							},
						},
					},
				}

				return cfg, mockState
			},
			wantBool: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, mockState := tt.setup(t)
			needsFetch, err := CheckStatus(context.Background(), cfg, mockState)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantBool, needsFetch)
		})
	}
}
