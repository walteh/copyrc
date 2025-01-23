package state

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestLogger(t *testing.T) context.Context {
	// Create a logger that writes to the test log
	logger := zerolog.New(zerolog.TestWriter{T: t}).With().Timestamp().Logger()
	return logger.WithContext(context.Background())
}

func setupTestDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "copyrc-state-test-*")
	require.NoError(t, err, "creating temp dir")
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestNew(t *testing.T) {
	t.Run("creates_new_state", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating new state")
		assert.NotNil(t, state, "state should not be nil")
		assert.Equal(t, filepath.Join(dir, ".copyrc.lock"), state.path)
		assert.Equal(t, "1.0.0", state.file.SchemaVersion)
	})
}

func TestLoadAndSave(t *testing.T) {
	ctx := setupTestLogger(t)

	t.Run("load_nonexistent_creates_clean", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating new state")

		err = state.Load(ctx)
		require.NoError(t, err, "loading nonexistent state")

		assert.Empty(t, state.file.Repositories)
		assert.Empty(t, state.file.RemoteTextFiles)
		assert.Empty(t, state.file.GeneratedFiles)
	})

	t.Run("save_and_load", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating new state")

		// Add some test data
		state.file.Repositories = []Repository{
			{
				Provider:  "github",
				Name:      "test/repo",
				LatestRef: "v1.0.0",
				Release: &Release{
					LastUpdated:  time.Now(),
					Ref:          "v1.0.0",
					RefHash:      "abc123",
					WebPermalink: "https://github.com/test/repo/releases/v1.0.0",
				},
			},
		}

		// Save the state
		err = state.Save(ctx)
		require.NoError(t, err, "saving state")

		// Load into a new state object
		state2, err := New(dir)
		require.NoError(t, err, "creating second state")
		err = state2.Load(ctx)
		require.NoError(t, err, "loading saved state")

		// Verify the data
		require.Len(t, state2.file.Repositories, 1)
		assert.Equal(t, "github", state2.file.Repositories[0].Provider)
		assert.Equal(t, "test/repo", state2.file.Repositories[0].Name)
		assert.Equal(t, "v1.0.0", state2.file.Repositories[0].LatestRef)
	})

	t.Run("concurrent_save_uses_lock", func(t *testing.T) {
		dir := setupTestDir(t)
		_, err := New(dir)
		require.NoError(t, err, "creating first state")
		state2, err := New(dir)
		require.NoError(t, err, "creating second state")

		// Create the lock file manually to simulate state1 having the lock
		lockPath := filepath.Join(dir, ".copyrc.lock.lock")
		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		require.NoError(t, err, "creating lock file")
		defer os.Remove(lockPath)
		defer lockFile.Close()

		// Try to save state2 while the lock file exists
		err = state2.Save(ctx)
		require.Error(t, err, "saving state2 should fail due to lock")
		assert.Contains(t, err.Error(), "creating lock file")
	})

	t.Run("invalid_json_returns_error", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Write invalid JSON
		err = os.WriteFile(filepath.Join(dir, ".copyrc.lock"), []byte("{invalid json}"), 0600)
		require.NoError(t, err, "writing invalid json")

		err = state.Load(ctx)
		require.Error(t, err, "loading invalid json should fail")
		assert.Contains(t, err.Error(), "parsing state file")
	})
}

func TestUserLogging(t *testing.T) {
	ctx := setupTestLogger(t)
	logger := NewUserLogger(ctx)

	t.Run("logs_file_changes", func(t *testing.T) {
		changes := []FileChange{
			{Type: FileAdded, Path: "foo.txt", Description: "new file"},
			{Type: FileUpdated, Path: "bar.txt", Description: "content changed"},
			{Type: FileDeleted, Path: "old.txt"},
			{Type: FileSkipped, Path: "skip.txt", Description: "already exists"},
			{Type: FileError, Path: "error.txt", Error: assert.AnError},
		}

		for _, change := range changes {
			logger.LogFileChange(change)
		}
	})

	t.Run("logs_state_changes", func(t *testing.T) {
		logger.LogStateChange("Initialized new state")
		logger.LogStateChange("Updated repository state")
	})

	t.Run("logs_validation", func(t *testing.T) {
		logger.LogValidation(true, "All files valid", nil)
		logger.LogValidation(false, "Missing files", assert.AnError)
		logger.LogValidation(false, "Hash mismatch", nil)
	})

	t.Run("logs_lock_operations", func(t *testing.T) {
		logger.LogLockOperation(true, ".copyrc.lock", nil)
		logger.LogLockOperation(false, ".copyrc.lock", assert.AnError)
		logger.LogLockOperation(false, ".copyrc.lock", nil)
	})
}
