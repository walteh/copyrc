package state

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/pkg/remote"
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

func TestPutRemoteTextFile(t *testing.T) {
	ctx := setupTestLogger(t)

	t.Run("adds_new_file", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Create mock file
		mockFile := &mockRawTextFile{
			content:   "test content",
			path:      "test.txt",
			permalink: "https://github.com/test/repo/blob/main/test.txt",
			repository: &mockRepository{
				name: "test/repo",
				release: &mockRelease{
					ref: "v1.0.0",
				},
			},
		}

		// Add file to state
		localPath := filepath.Join(dir, "test.copy.txt")
		file, err := state.PutRemoteTextFile(ctx, mockFile, localPath)
		require.NoError(t, err, "putting remote text file")
		assert.NotNil(t, file, "file should not be nil")

		// Verify file was written
		content, err := os.ReadFile(localPath)
		require.NoError(t, err, "reading file")
		assert.Equal(t, "test content", string(content), "file content should match")

		// Verify state was updated
		assert.Len(t, state.file.RemoteTextFiles, 1, "should have one file")
		assert.Equal(t, localPath, state.file.RemoteTextFiles[0].LocalPath, "file path should match")
		assert.Equal(t, "test/repo", state.file.RemoteTextFiles[0].RepoName, "repo name should match")
		assert.Equal(t, "v1.0.0", state.file.RemoteTextFiles[0].ReleaseRef, "release ref should match")
	})

	t.Run("updates_existing_file", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Create mock file
		mockFile := &mockRawTextFile{
			content:   "test content",
			path:      "test.txt",
			permalink: "https://github.com/test/repo/blob/main/test.txt",
			repository: &mockRepository{
				name: "test/repo",
				release: &mockRelease{
					ref: "v1.0.0",
				},
			},
		}

		// Add file to state twice
		localPath := filepath.Join(dir, "test.copy.txt")
		_, err = state.PutRemoteTextFile(ctx, mockFile, localPath)
		require.NoError(t, err, "putting remote text file first time")

		// Update content
		mockFile.content = "updated content"
		file, err := state.PutRemoteTextFile(ctx, mockFile, localPath)
		require.NoError(t, err, "putting remote text file second time")
		assert.NotNil(t, file, "file should not be nil")

		// Verify file was updated
		content, err := os.ReadFile(localPath)
		require.NoError(t, err, "reading file")
		assert.Equal(t, "updated content", string(content), "file content should be updated")

		// Verify state was updated
		assert.Len(t, state.file.RemoteTextFiles, 1, "should still have one file")
	})

	t.Run("validates_file_suffix", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Create mock file
		mockFile := &mockRawTextFile{
			content:   "test content",
			path:      "test.txt",
			permalink: "https://github.com/test/repo/blob/main/test.txt",
			repository: &mockRepository{
				name: "test/repo",
				release: &mockRelease{
					ref: "v1.0.0",
				},
			},
		}

		// Try to add file with invalid suffix
		localPath := filepath.Join(dir, "test.txt")
		_, err = state.PutRemoteTextFile(ctx, mockFile, localPath)
		require.Error(t, err, "putting file with invalid suffix should error")
		assert.Contains(t, err.Error(), "invalid file suffix", "error should mention invalid suffix")
	})
}

func TestPutGeneratedFile(t *testing.T) {
	ctx := setupTestLogger(t)

	t.Run("adds_new_file", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Create test files
		mainFile := filepath.Join(dir, "main.go")
		genFile := filepath.Join(dir, "main_gen.go")
		err = os.WriteFile(mainFile, []byte("package main"), 0644)
		require.NoError(t, err, "writing main file")
		err = os.WriteFile(genFile, []byte("package main // generated"), 0644)
		require.NoError(t, err, "writing generated file")

		// Add file to state
		file := GeneratedFile{
			LocalPath:     genFile,
			ReferenceFile: mainFile,
		}
		result, err := state.PutGeneratedFile(ctx, file)
		require.NoError(t, err, "putting generated file")
		assert.NotNil(t, result, "result should not be nil")

		// Verify state was updated
		assert.Len(t, state.file.GeneratedFiles, 1, "should have one file")
		assert.Equal(t, genFile, state.file.GeneratedFiles[0].LocalPath, "file path should match")
		assert.Equal(t, mainFile, state.file.GeneratedFiles[0].ReferenceFile, "reference file should match")
	})

	t.Run("updates_existing_file", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Create test files
		mainFile := filepath.Join(dir, "main.go")
		genFile := filepath.Join(dir, "main_gen.go")
		err = os.WriteFile(mainFile, []byte("package main"), 0644)
		require.NoError(t, err, "writing main file")
		err = os.WriteFile(genFile, []byte("package main // generated"), 0644)
		require.NoError(t, err, "writing generated file")

		// Add file to state twice
		file := GeneratedFile{
			LocalPath:     genFile,
			ReferenceFile: mainFile,
		}
		_, err = state.PutGeneratedFile(ctx, file)
		require.NoError(t, err, "putting generated file first time")

		// Update reference file
		newMainFile := filepath.Join(dir, "new_main.go")
		err = os.WriteFile(newMainFile, []byte("package main"), 0644)
		require.NoError(t, err, "writing new main file")

		file.ReferenceFile = newMainFile
		result, err := state.PutGeneratedFile(ctx, file)
		require.NoError(t, err, "putting generated file second time")
		assert.NotNil(t, result, "result should not be nil")

		// Verify state was updated
		assert.Len(t, state.file.GeneratedFiles, 1, "should still have one file")
		assert.Equal(t, newMainFile, state.file.GeneratedFiles[0].ReferenceFile, "reference file should be updated")
	})

	t.Run("validates_file_existence", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Try to add non-existent file
		file := GeneratedFile{
			LocalPath:     filepath.Join(dir, "missing_gen.go"),
			ReferenceFile: filepath.Join(dir, "missing.go"),
		}
		_, err = state.PutGeneratedFile(ctx, file)
		require.Error(t, err, "putting non-existent file should error")
		assert.Contains(t, err.Error(), "file does not exist", "error should mention missing file")

		// Create generated file but not reference file
		err = os.WriteFile(file.LocalPath, []byte("package main"), 0644)
		require.NoError(t, err, "writing generated file")

		_, err = state.PutGeneratedFile(ctx, file)
		require.Error(t, err, "putting file with missing reference should error")
		assert.Contains(t, err.Error(), "reference file does not exist", "error should mention missing reference")
	})
}

func TestPutArchiveFile(t *testing.T) {
	ctx := setupTestLogger(t)

	t.Run("adds_new_file", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Create test archive file
		archivePath := filepath.Join(dir, "test.tar.gz")
		err = os.WriteFile(archivePath, []byte("mock archive content"), 0644)
		require.NoError(t, err, "writing archive file")

		// Add file to state
		file := ArchiveFile{
			LocalPath: archivePath,
		}
		result, err := state.PutArchiveFile(ctx, file)
		require.NoError(t, err, "putting archive file")
		assert.NotNil(t, result, "result should not be nil")

		// Verify state was updated
		assert.Len(t, state.file.ArchiveFiles, 1, "should have one file")
		assert.Equal(t, archivePath, state.file.ArchiveFiles[0].LocalPath, "file path should match")
		assert.NotEmpty(t, state.file.ArchiveFiles[0].ContentHash, "content hash should be set")
	})

	t.Run("updates_existing_file", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Create test archive file
		archivePath := filepath.Join(dir, "test.tar.gz")
		err = os.WriteFile(archivePath, []byte("mock archive content"), 0644)
		require.NoError(t, err, "writing archive file")

		// Add file to state twice
		file := ArchiveFile{
			LocalPath: archivePath,
		}
		_, err = state.PutArchiveFile(ctx, file)
		require.NoError(t, err, "putting archive file first time")

		// Update file content
		err = os.WriteFile(archivePath, []byte("updated archive content"), 0644)
		require.NoError(t, err, "updating archive file")

		result, err := state.PutArchiveFile(ctx, file)
		require.NoError(t, err, "putting archive file second time")
		assert.NotNil(t, result, "result should not be nil")

		// Verify state was updated
		assert.Len(t, state.file.ArchiveFiles, 1, "should still have one file")
		assert.NotEqual(t, "", state.file.ArchiveFiles[0].ContentHash, "content hash should be updated")
	})

	t.Run("validates_file_existence", func(t *testing.T) {
		dir := setupTestDir(t)
		state, err := New(dir)
		require.NoError(t, err, "creating state")

		// Try to add non-existent file
		file := ArchiveFile{
			LocalPath: filepath.Join(dir, "missing.tar.gz"),
		}
		_, err = state.PutArchiveFile(ctx, file)
		require.Error(t, err, "putting non-existent file should error")
		assert.Contains(t, err.Error(), "file does not exist", "error should mention missing file")
	})
}

// Mock implementations for testing
type mockRawTextFile struct {
	content    string
	path       string
	permalink  string
	repository *mockRepository
}

func (m *mockRawTextFile) GetContent(ctx context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(m.content)), nil
}

func (m *mockRawTextFile) Path() string {
	return m.path
}

func (m *mockRawTextFile) RawTextPermalink() string {
	return m.permalink
}

func (m *mockRawTextFile) WebViewPermalink() string {
	return m.permalink
}

func (m *mockRawTextFile) Release() remote.Release {
	return m.repository.release
}

type mockRepository struct {
	name    string
	release *mockRelease
}

func (m *mockRepository) Name() string {
	return m.name
}

func (m *mockRepository) GetLatestRelease(ctx context.Context) (remote.Release, error) {
	return m.release, nil
}

func (m *mockRepository) GetReleaseFromRef(ctx context.Context, ref string) (remote.Release, error) {
	return m.release, nil
}

type mockRelease struct {
	ref string
}

func (m *mockRelease) Repository() remote.Repository {
	return &mockRepository{name: "test/repo", release: m}
}

func (m *mockRelease) Ref() string {
	return m.ref
}

func (m *mockRelease) GetTarball(ctx context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockRelease) ListFilesAtPath(ctx context.Context, path string) ([]remote.RawTextFile, error) {
	return nil, nil
}

func (m *mockRelease) GetFileAtPath(ctx context.Context, path string) (remote.RawTextFile, error) {
	return nil, nil
}

func (m *mockRelease) GetLicense(ctx context.Context) (io.ReadCloser, string, error) {
	return nil, "", nil
}

func (m *mockRelease) GetLicenseAtPath(ctx context.Context, path string) (io.ReadCloser, string, error) {
	return nil, "", nil
}
