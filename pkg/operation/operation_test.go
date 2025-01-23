package operation

import (
	"context"
	"io"
	"os"
	"path/filepath"
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

func setupTestDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "copyrc-operation-test-*")
	require.NoError(t, err, "creating temp dir")
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// TestCopyFiles tests the CopyFiles function
func TestCopyFiles(t *testing.T) {
	type testCase struct {
		name    string
		setup   func(t *testing.T, dir string) (*config.CopyrcConfig, remote.Provider, *state.State)
		wantErr bool
	}

	tests := []testCase{
		{
			name: "copies_single_file",
			setup: func(t *testing.T, dir string) (*config.CopyrcConfig, remote.Provider, *state.State) {
				mockFile := mockery.NewMockRawTextFile_remote(t)
				mockRelease := mockery.NewMockRelease_remote(t)
				mockRepo := mockery.NewMockRepository_remote(t)
				mockProvider := mockery.NewMockProvider_remote(t)

				// Setup mock expectations
				mockFile.EXPECT().GetContent(mock.Anything).Return(io.NopCloser(strings.NewReader("old content")), nil).Times(2)
				mockFile.EXPECT().Path().Return("test.txt").Times(4)
				mockFile.EXPECT().RawTextPermalink().Return("raw_link").Times(2)
				mockFile.EXPECT().WebViewPermalink().Return("web_link").Times(2)
				mockFile.EXPECT().Release().Return(mockRelease).Times(4)

				mockRelease.EXPECT().Ref().Return("ref").Times(4)
				mockRelease.EXPECT().Repository().Return(mockRepo).Times(4)
				mockRepo.EXPECT().Name().Return("repo").Times(4)

				mockRepo.EXPECT().GetLatestRelease(mock.Anything).Return(mockRelease, nil).Times(1)
				mockRelease.EXPECT().ListFilesAtPath(mock.Anything, "remote/path").Return([]remote.RawTextFile{mockFile}, nil).Times(1)

				mockProvider.EXPECT().GetRepository(mock.Anything, "repo").Return(mockRepo, nil).Times(1)
				mockProvider.EXPECT().Name().Return("provider").Times(2)

				// Create config
				cfg := &config.CopyrcConfig{
					Repositories: []config.RepositoryDefinition{
						{
							Provider: "provider",
							Name:     "repo",
						},
					},
					Copies: []config.Copy{
						{
							Repository: config.RepositoryDefinition{
								Provider: "provider",
								Name:     "repo",
							},
							Paths: config.CopyPaths{
								Remote: "remote/path",
								Local:  "test.txt",
							},
							Options: config.CopyOptions{},
						},
					},
				}

				// Create state
				st, err := state.New(dir)
				require.NoError(t, err, "creating state")

				return cfg, mockProvider, st
			},
			wantErr: false,
		},
		{
			name: "applies_text_replacements",
			setup: func(t *testing.T, dir string) (*config.CopyrcConfig, remote.Provider, *state.State) {
				mockFile := mockery.NewMockRawTextFile_remote(t)
				mockRelease := mockery.NewMockRelease_remote(t)
				mockRepo := mockery.NewMockRepository_remote(t)
				mockProvider := mockery.NewMockProvider_remote(t)

				// Create a reusable content reader
				content := "old content"
				mockFile.EXPECT().GetContent(mock.Anything).RunAndReturn(func(ctx context.Context) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(content)), nil
				}).Times(2)

				mockFile.EXPECT().Path().Return("test.txt").Times(4)
				mockFile.EXPECT().RawTextPermalink().Return("https://raw.githubusercontent.com/test/repo/main/test.txt").Times(4)
				mockFile.EXPECT().WebViewPermalink().Return("https://github.com/test/repo/blob/main/test.txt").Times(4)
				mockFile.EXPECT().Release().Return(mockRelease).Times(4)

				mockRelease.EXPECT().Ref().Return("v1.0.0").Times(4)
				mockRelease.EXPECT().ListFilesAtPath(mock.Anything, "test.txt").Return([]remote.RawTextFile{mockFile}, nil).Once()
				mockRelease.EXPECT().Repository().Return(mockRepo).Times(4)

				mockRepo.EXPECT().Name().Return("test/repo").Times(4)
				mockRepo.EXPECT().GetLatestRelease(mock.Anything).Return(mockRelease, nil).Once()

				mockProvider.EXPECT().GetRepository(mock.Anything, "test/repo").Return(mockRepo, nil).Once()
				mockProvider.EXPECT().Name().Return("github").Times(2)

				// Create config
				cfg := &config.CopyrcConfig{
					Repositories: []config.RepositoryDefinition{
						{
							Provider: "github",
							Name:     "test/repo",
						},
					},
					Copies: []config.Copy{
						{
							Repository: config.RepositoryDefinition{
								Provider: "github",
								Name:     "test/repo",
							},
							Paths: config.CopyPaths{
								Remote: "test.txt",
								Local:  ".",
							},
							Options: config.CopyOptions{
								TextReplacements: []config.TextReplacement{
									{
										FromText: "old",
										ToText:   "new",
									},
								},
							},
						},
					},
				}

				// Create state
				st, err := state.New(dir)
				require.NoError(t, err, "creating state")

				return cfg, mockProvider, st
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			dir := t.TempDir()

			cfg, mockProvider, st := tt.setup(t, dir)

			err := CopyFiles(ctx, cfg, mockProvider, st)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify file content
			content, err := os.ReadFile(filepath.Join(st.Dir(), "test.copy.txt"))
			require.NoError(t, err)
			require.Equal(t, "old content", string(content))
		})
	}
}

func TestValidateFiles(t *testing.T) {
	ctx := setupTestLogger(t)

	t.Run("validates_consistent_state", func(t *testing.T) {
		dir := setupTestDir(t)
		st, err := state.New(dir)
		require.NoError(t, err, "creating state")

		// Create a file
		err = os.WriteFile(filepath.Join(dir, "test.copy.txt"), []byte("test content"), 0644)
		require.NoError(t, err, "writing test file")

		// Create mock repository for release
		mockRepoForRelease := mockery.NewMockRepository_remote(t)
		mockRepoForRelease.EXPECT().Name().Return("test/repo")

		// Create mock release
		mockRelease := mockery.NewMockRelease_remote(t)
		mockRelease.EXPECT().Ref().Return("v1.0.0").Maybe()
		mockRelease.EXPECT().Repository().Return(mockRepoForRelease).Maybe()

		// Create mock file
		mockFile := mockery.NewMockRawTextFile_remote(t)
		mockFile.EXPECT().GetContent(mock.Anything).Return(io.NopCloser(strings.NewReader("test content")), nil).Maybe()
		mockFile.EXPECT().Path().Return("test.txt").Maybe()
		mockFile.EXPECT().RawTextPermalink().Return("https://github.com/test/repo/blob/main/test.txt").Maybe()
		mockFile.EXPECT().WebViewPermalink().Return("https://github.com/test/repo/blob/main/test.txt").Maybe()
		mockFile.EXPECT().Release().Return(mockRelease).Maybe()

		// Add file to state
		_, err = st.PutRemoteTextFile(ctx, mockFile, filepath.Join(dir, "test.copy.txt"))
		require.NoError(t, err, "putting file in state")

		// Validate files
		err = ValidateFiles(ctx, st)
		require.NoError(t, err, "validating files")
	})

	t.Run("detects_inconsistent_state", func(t *testing.T) {
		dir := setupTestDir(t)
		st, err := state.New(dir)
		require.NoError(t, err, "creating state")

		// Create a file with one content
		err = os.WriteFile(filepath.Join(dir, "test.copy.txt"), []byte("test content"), 0644)
		require.NoError(t, err, "writing test file")

		// Create mock repository for release
		mockRepoForRelease := mockery.NewMockRepository_remote(t)
		mockRepoForRelease.EXPECT().Name().Return("test/repo")

		// Create mock release
		mockRelease := mockery.NewMockRelease_remote(t)
		mockRelease.EXPECT().Ref().Return("v1.0.0").Maybe()
		mockRelease.EXPECT().Repository().Return(mockRepoForRelease).Maybe()

		// Create mock file with different content
		mockFile := mockery.NewMockRawTextFile_remote(t)
		mockFile.EXPECT().GetContent(mock.Anything).Return(io.NopCloser(strings.NewReader("different content")), nil).Maybe()
		mockFile.EXPECT().Path().Return("test.txt").Maybe()
		mockFile.EXPECT().RawTextPermalink().Return("https://github.com/test/repo/blob/main/test.txt").Maybe()
		mockFile.EXPECT().WebViewPermalink().Return("https://github.com/test/repo/blob/main/test.txt").Maybe()
		mockFile.EXPECT().Release().Return(mockRelease).Maybe()

		// Add file to state with different content
		_, err = st.PutRemoteTextFile(ctx, mockFile, filepath.Join(dir, "test.copy.txt"))
		require.NoError(t, err, "putting file in state")

		// Modify the file after adding it to state
		err = os.WriteFile(filepath.Join(dir, "test.copy.txt"), []byte("modified content"), 0644)
		require.NoError(t, err, "modifying test file")

		// Validate files
		err = ValidateFiles(ctx, st)
		require.Error(t, err, "validation should fail")
		assert.Contains(t, err.Error(), "hash mismatch", "error should mention hash mismatch")
	})
}

func TestCheckStatus(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, ctx context.Context, dir string, mockFile *mockery.MockRawTextFile_remote) (*state.State, error)
		wantErr  bool
		wantBool bool
	}{
		{
			name: "inconsistent_state_returns_true",
			setup: func(t *testing.T, ctx context.Context, dir string, mockFile *mockery.MockRawTextFile_remote) (*state.State, error) {
				st, err := state.New(dir)
				if err != nil {
					return nil, err
				}

				mockRelease := mockery.NewMockRelease_remote(t)
				mockRepo := mockery.NewMockRepository_remote(t)

				// Create a reusable content reader
				content := "test content"
				mockFile.EXPECT().GetContent(mock.Anything).RunAndReturn(func(ctx context.Context) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(content)), nil
				}).Times(2)

				mockFile.EXPECT().Path().Return("test.txt").Times(4)
				mockFile.EXPECT().GetContent(mock.Anything).Return(io.NopCloser(strings.NewReader("old content")), nil).Times(1)
				mockFile.EXPECT().RawTextPermalink().Return("raw_link").Times(3)
				mockFile.EXPECT().WebViewPermalink().Return("web_link").Times(3)
				mockFile.EXPECT().Release().Return(mockRelease).Times(3)

				mockRelease.EXPECT().ListFilesAtPath(mock.Anything, mock.Anything).Return([]remote.RawTextFile{mockFile}, nil).Times(1)
				mockRelease.EXPECT().Ref().Return("ref").Times(3)
				mockRelease.EXPECT().Repository().Return(mockRepo).Times(3)

				mockRepo.EXPECT().Name().Return("repo").Times(3)

				// Put the file in state
				_, err = st.PutRemoteTextFile(ctx, mockFile, filepath.Join(dir, "test.copy.txt"))
				if err != nil {
					return nil, err
				}

				// Write different content to make the hash mismatch
				err = os.WriteFile(filepath.Join(dir, "test.copy.txt"), []byte("modified content"), 0644)
				if err != nil {
					return nil, err
				}

				return st, nil
			},
			wantErr:  false,
			wantBool: true,
		},
		{
			name: "consistent_state_returns_false",
			setup: func(t *testing.T, ctx context.Context, dir string, mockFile *mockery.MockRawTextFile_remote) (*state.State, error) {
				st, err := state.New(dir)
				if err != nil {
					return nil, err
				}

				mockRelease := mockery.NewMockRelease_remote(t)
				mockRepo := mockery.NewMockRepository_remote(t)

				// Create a reusable content reader
				content := "test content"
				mockFile.EXPECT().GetContent(mock.Anything).RunAndReturn(func(ctx context.Context) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(content)), nil
				}).Times(2)

				mockFile.EXPECT().Path().Return("test.txt").Times(4)
				mockFile.EXPECT().GetContent(mock.Anything).Return(io.NopCloser(strings.NewReader("old content")), nil).Times(1)
				mockFile.EXPECT().RawTextPermalink().Return("raw_link").Times(3)
				mockFile.EXPECT().WebViewPermalink().Return("web_link").Times(3)
				mockFile.EXPECT().Release().Return(mockRelease).Times(3)

				mockRelease.EXPECT().ListFilesAtPath(mock.Anything, mock.Anything).Return([]remote.RawTextFile{mockFile}, nil).Times(1)
				mockRelease.EXPECT().Ref().Return("ref").Times(3)
				mockRelease.EXPECT().Repository().Return(mockRepo).Times(3)

				mockRepo.EXPECT().Name().Return("repo").Times(3)

				// Put the file in state
				_, err = st.PutRemoteTextFile(ctx, mockFile, filepath.Join(dir, "test.copy.txt"))
				if err != nil {
					return nil, err
				}

				return st, nil
			},
			wantErr:  false,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupTestLogger(t)
			dir := t.TempDir()

			mockFile := mockery.NewMockRawTextFile_remote(t)
			st, err := tt.setup(t, ctx, dir, mockFile)
			require.NoError(t, err)

			needsFetch, err := CheckStatus(ctx, &config.CopyrcConfig{}, st)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantBool, needsFetch)
		})
	}
}
