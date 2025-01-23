package testutils

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/gen/mockery"
	"github.com/walteh/copyrc/pkg/state"
)

// TestSimpleStateFileOperation tests the basic file operations with state
func TestSimpleStateFileOperation(t *testing.T) {
	// Create a temporary directory
	dir := t.TempDir()

	// Initialize state
	st, err := state.New(dir)
	require.NoError(t, err, "creating state")

	// Create mock objects
	mockFile := mockery.NewMockRawTextFile_remote(t)
	mockRelease := mockery.NewMockRelease_remote(t)
	mockRepo := mockery.NewMockRepository_remote(t)

	// Setup basic expectations
	mockFile.EXPECT().GetContent(mock.Anything).Return(io.NopCloser(strings.NewReader("test content")), nil).Times(1)
	mockFile.EXPECT().Path().Return("test.txt").Times(4)
	mockFile.EXPECT().WebViewPermalink().Return("web_link").Times(4)
	mockFile.EXPECT().Release().Return(mockRelease).Times(4)
	mockRelease.EXPECT().Ref().Return("ref").Times(4)
	mockRelease.EXPECT().Repository().Return(mockRepo).Times(4)
	mockRepo.EXPECT().Name().Return("repo").Times(4)

	// Create a logger context
	logger := zerolog.New(zerolog.TestWriter{T: t}).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())

	// Put file in state
	localPath := filepath.Join(dir, "test.copy.txt")
	_, err = st.PutRemoteTextFile(ctx, mockFile, localPath)
	require.NoError(t, err, "putting file in state")

	// Verify file content
	content, err := os.ReadFile(localPath)
	require.NoError(t, err, "reading file")
	require.Equal(t, "test content", string(content), "file content should match")
}
