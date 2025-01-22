package status

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/gen/mockery"
)

// 🛠️ setupTestManager creates a new Manager with a test logger and mock formatter
func setupTestManager(t *testing.T, tmpDir string) *Manager {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	formatter := mockery.NewMockFileFormatter_status(t)
	mgr := New(tmpDir, &logger)
	mgr.formatter = formatter
	return mgr
}

// 🎭 MockFileFormatter mocks the FileFormatter interface
type MockFileFormatter struct {
	mock.Mock
}

func (m *MockFileFormatter) FormatFileOperation(path, fileType, status string, isNew, isModified, isRemoved bool) string {
	args := m.Called(path, fileType, status, isNew, isModified, isRemoved)
	return args.String(0)
}

func (m *MockFileFormatter) FormatProgress(current, total int) string {
	args := m.Called(current, total)
	return args.String(0)
}

func (m *MockFileFormatter) FormatError(err error) string {
	args := m.Called(err)
	return args.String(0)
}

// 🧪 TestManager tests the status manager functionality
func TestManager(t *testing.T) {
	// Create test logger and mock formatter
	logger := zerolog.New(zerolog.NewTestWriter(t))
	formatter := mockery.NewMockFileFormatter_status(t)

	// Setup expectations
	formatter.EXPECT().FormatFileOperation("test.txt", "file", "ok", true, false, false).
		Return("✨ Created test.txt")
	formatter.EXPECT().FormatProgress(0, 1).
		Return("⏳ Progress: 0/1 (0%)")
	formatter.EXPECT().FormatProgress(1, 1).
		Return("✅ Progress: 1/1 (100%)")

	// Create manager with mock formatter
	mgr := New(t.TempDir(), &logger)
	mgr.formatter = formatter

	// Test tracking a new file
	ctx := context.Background()
	mgr.TrackFile(ctx, "test.txt", FileInfo{
		Path:     "test.txt",
		Status:   StatusNew,
		Size:     100,
		Checksum: "abc123",
	})

	// Verify file was tracked
	info, err := mgr.GetFileInfo(ctx, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, "test.txt", info.Path)
	assert.Equal(t, StatusNew, info.Status)
	assert.Equal(t, int64(100), info.Size)
	assert.Equal(t, "abc123", info.Checksum)

	// Test progress reporting
	mgr.StartOperation(ctx, 1)
	mgr.UpdateProgress(ctx, 1)
	mgr.FinishOperation(ctx)
}

// 🧪 TestFileOperations tests the file management functionality
func TestFileOperations(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	mgr := New(tmpDir, &logger)

	ctx := context.Background()

	// Test writing a file
	content := []byte("test content")
	err := mgr.WriteFile(ctx, "test.txt", content)
	require.NoError(t, err)

	// Test reading the file
	readContent, err := mgr.ReadFile(ctx, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, content, readContent)

	// Test file exists
	exists, err := mgr.FileExists(ctx, "test.txt")
	require.NoError(t, err)
	assert.True(t, exists)

	// Test backup and restore
	err = mgr.BackupFile(ctx, "test.txt")
	require.NoError(t, err)

	err = mgr.WriteFile(ctx, "test.txt", []byte("modified content"))
	require.NoError(t, err)

	err = mgr.RestoreFile(ctx, "test.txt")
	require.NoError(t, err)

	// Verify restored content
	restoredContent, err := mgr.ReadFile(ctx, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, content, restoredContent)

	// Test deleting the file
	err = mgr.DeleteFile(ctx, "test.txt")
	require.NoError(t, err)

	exists, err = mgr.FileExists(ctx, "test.txt")
	require.NoError(t, err)
	assert.False(t, exists)
}

// 🧪 TestErrorHandling tests error cases
func TestErrorHandling(t *testing.T) {
	// Create test logger and mock formatter
	logger := zerolog.New(zerolog.NewTestWriter(t))
	formatter := mockery.NewMockFileFormatter_status(t)

	// Setup error formatting expectation
	testErr := assert.AnError
	formatter.EXPECT().FormatError(testErr).
		Return("❌ Error: test error")
	formatter.EXPECT().FormatFileOperation("error.txt", "file", "ok", true, false, false).
		Return("✨ Created error.txt")

	// Create manager with mock formatter
	mgr := New(t.TempDir(), &logger)
	mgr.formatter = formatter

	// Test tracking a file with error
	ctx := context.Background()
	mgr.TrackFile(ctx, "error.txt", FileInfo{
		Path:   "error.txt",
		Status: StatusNew,
		Size:   0,
		Error:  testErr,
	})

	// Verify error was tracked
	info, err := mgr.GetFileInfo(ctx, "error.txt")
	require.NoError(t, err)
	assert.Equal(t, testErr, info.Error)
}

// 🧪 TestListFiles tests file listing functionality
func TestListFiles(t *testing.T) {
	// Create test logger and mock formatter
	logger := zerolog.New(zerolog.NewTestWriter(t))
	formatter := mockery.NewMockFileFormatter_status(t)

	// Setup formatting expectations
	formatter.EXPECT().FormatFileOperation("file1.txt", "file", "ok", true, false, false).
		Return("✨ Created file1.txt")
	formatter.EXPECT().FormatFileOperation("file2.txt", "file", "ok", false, true, false).
		Return("📝 Modified file2.txt")

	// Create manager with mock formatter
	mgr := New(t.TempDir(), &logger)
	mgr.formatter = formatter

	// Track some files
	ctx := context.Background()
	mgr.TrackFile(ctx, "file1.txt", FileInfo{
		Path:   "file1.txt",
		Status: StatusNew,
		Size:   100,
	})
	mgr.TrackFile(ctx, "file2.txt", FileInfo{
		Path:   "file2.txt",
		Status: StatusModified,
		Size:   200,
	})

	// Test listing files
	files, err := mgr.ListFiles(ctx)
	require.NoError(t, err)
	assert.Len(t, files, 2)

	// Verify file info
	assert.Contains(t, files, FileInfo{
		Path:   "file1.txt",
		Status: StatusNew,
		Size:   100,
	})
	assert.Contains(t, files, FileInfo{
		Path:   "file2.txt",
		Status: StatusModified,
		Size:   200,
	})
}

// 🧪 TestFileStatusString tests the String method of FileStatus
func TestFileStatusString(t *testing.T) {
	tests := []struct {
		name        string
		status      FileStatus
		want        string
		description string
	}{
		{
			name:        "status_new",
			status:      StatusNew,
			want:        "new",
			description: "should return 'new' for StatusNew",
		},
		{
			name:        "status_modified",
			status:      StatusModified,
			want:        "modified",
			description: "should return 'modified' for StatusModified",
		},
		{
			name:        "status_unchanged",
			status:      StatusUnchanged,
			want:        "unchanged",
			description: "should return 'unchanged' for StatusUnchanged",
		},
		{
			name:        "status_deleted",
			status:      StatusDeleted,
			want:        "deleted",
			description: "should return 'deleted' for StatusDeleted",
		},
		{
			name:        "status_unknown",
			status:      StatusUnknown,
			want:        "unknown",
			description: "should return 'unknown' for StatusUnknown",
		},
		{
			name:        "status_invalid",
			status:      FileStatus(999),
			want:        "unknown",
			description: "should return 'unknown' for invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

// 🧪 TestCalculateChecksum tests the checksum calculation
func TestCalculateChecksum(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		want        string
		description string
	}{
		{
			name:        "empty_content",
			content:     []byte{},
			want:        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			description: "should calculate correct hash for empty content",
		},
		{
			name:        "simple_content",
			content:     []byte("hello world"),
			want:        "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			description: "should calculate correct hash for simple content",
		},
		{
			name:        "binary_content",
			content:     []byte{0x00, 0xFF, 0x42},
			want:        "f803bec586282caafe409609aae90eb09f6d4cddb6e04431ddf76d22e7dcacd6",
			description: "should calculate correct hash for binary content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateChecksum(tt.content)
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

// 🧪 TestDirectoryOperations tests directory management functionality
func TestDirectoryOperations(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	mgr := New(tmpDir, &logger)
	ctx := context.Background()

	// Test creating nested directories
	err := mgr.CreateDir(ctx, "test/nested/dir")
	require.NoError(t, err, "should create nested directories")

	// Verify directory exists
	exists, err := mgr.FileExists(ctx, "test/nested/dir")
	require.NoError(t, err)
	assert.True(t, exists, "directory should exist")

	// Test creating directory that already exists
	err = mgr.CreateDir(ctx, "test/nested/dir")
	require.NoError(t, err, "should handle existing directory")

	// Create a file in the directory
	err = mgr.WriteFile(ctx, "test/nested/dir/file.txt", []byte("test"))
	require.NoError(t, err, "should create file in directory")

	// Test removing directory with content
	err = mgr.RemoveDir(ctx, "test")
	require.NoError(t, err, "should remove directory and contents")

	// Verify directory is gone
	exists, err = mgr.FileExists(ctx, "test")
	require.NoError(t, err)
	assert.False(t, exists, "directory should not exist")

	// Test removing non-existent directory
	err = mgr.RemoveDir(ctx, "nonexistent")
	require.NoError(t, err, "should handle non-existent directory")
}

// 🧪 TestFileOperationErrors tests error cases for file operations
func TestFileOperationErrors(t *testing.T) {
	// Create a test logger
	logger := zerolog.New(zerolog.NewTestWriter(t))

	// Create a mock formatter
	formatter := &MockFileFormatter{}

	// Setup error formatting expectation
	testErr := assert.AnError
	formatter.On("FormatError", testErr).
		Return("❌ Error: test error")
	formatter.On("FormatFileOperation", "error.txt", "file", "ok", true, false, false).
		Return("✨ Created error.txt")

	// Create manager with mock formatter
	mgr := New(t.TempDir(), &logger)
	mgr.formatter = formatter

	// Test tracking a file with error
	ctx := context.Background()
	mgr.TrackFile(ctx, "error.txt", FileInfo{
		Path:   "error.txt",
		Status: StatusNew,
		Size:   0,
		Error:  testErr,
	})

	// Verify error was tracked
	info, err := mgr.GetFileInfo(ctx, "error.txt")
	require.NoError(t, err)
	assert.Equal(t, testErr, info.Error)

	// Verify mock expectations
	formatter.AssertExpectations(t)
}

// 🧪 TestFileOperationEdgeCases tests edge cases for file operations
func TestFileOperationEdgeCases(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	mgr := New(tmpDir, &logger)
	ctx := context.Background()

	// Test writing to directory that doesn't exist
	err := mgr.WriteFile(ctx, "nonexistent/test.txt", []byte("test"))
	require.NoError(t, err, "should create parent directories")

	// Test writing to file that already exists
	err = mgr.WriteFile(ctx, "test.txt", []byte("original"))
	require.NoError(t, err, "should create file")
	err = mgr.WriteFile(ctx, "test.txt", []byte("modified"))
	require.NoError(t, err, "should overwrite file")

	// Test deleting non-existent file
	err = mgr.DeleteFile(ctx, "nonexistent.txt")
	require.Error(t, err, "should error on non-existent file")

	// Test file exists with invalid path
	exists, err := mgr.FileExists(ctx, string([]byte{0x00}))
	require.Error(t, err, "should error on invalid path")
	assert.False(t, exists)

	// Test creating directory that already exists
	err = mgr.CreateDir(ctx, "test")
	require.NoError(t, err, "should create directory")
	err = mgr.CreateDir(ctx, "test")
	require.NoError(t, err, "should handle existing directory")

	// Test creating directory with invalid path
	err = mgr.CreateDir(ctx, string([]byte{0x00}))
	require.Error(t, err, "should error on invalid path")

	// Test removing directory with invalid path
	err = mgr.RemoveDir(ctx, string([]byte{0x00}))
	require.Error(t, err, "should error on invalid path")

	// Test backup file that doesn't exist
	err = mgr.BackupFile(ctx, "nonexistent.txt")
	require.NoError(t, err, "should handle non-existent file")

	// Test backup file with invalid path
	err = mgr.BackupFile(ctx, string([]byte{0x00}))
	require.Error(t, err, "should error on invalid path")

	// Test restore file with invalid path
	err = mgr.RestoreFile(ctx, string([]byte{0x00}))
	require.Error(t, err, "should error on invalid path")

	// Test get file info for non-existent file
	_, err = mgr.GetFileInfo(ctx, "nonexistent.txt")
	require.Error(t, err, "should error on non-existent file")
}

// 🧪 TestCopyFileEdgeCases tests edge cases for the CopyFile function
func TestCopyFileEdgeCases(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	mgr := New(tmpDir, &logger)
	ctx := context.Background()

	// Test copying non-existent file
	err := mgr.CopyFile("nonexistent.txt", "dest.txt")
	require.Error(t, err, "should error on non-existent source")

	// Test copying to invalid destination
	err = mgr.WriteFile(ctx, "source.txt", []byte("test"))
	require.NoError(t, err, "should create source file")
	err = mgr.CopyFile(filepath.Join(tmpDir, "source.txt"), string([]byte{0x00}))
	require.Error(t, err, "should error on invalid destination")

	// Test copying to read-only destination directory
	destDir := filepath.Join(tmpDir, "readonly")
	err = os.MkdirAll(destDir, 0500)
	require.NoError(t, err, "should create read-only directory")
	err = mgr.CopyFile(filepath.Join(tmpDir, "source.txt"), filepath.Join(destDir, "dest.txt"))
	require.Error(t, err, "should error on read-only destination")

	// Restore directory permissions for cleanup
	err = os.Chmod(destDir, 0700)
	require.NoError(t, err, "should restore directory permissions")
}

// 🧪 TestBackupRestoreEdgeCases tests edge cases for backup and restore operations
func TestBackupRestoreEdgeCases(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	mgr := New(tmpDir, &logger)
	ctx := context.Background()

	// Test backup with read-only parent directory
	err := mgr.WriteFile(ctx, "test.txt", []byte("test"))
	require.NoError(t, err, "should create test file")

	err = os.Chmod(tmpDir, 0500)
	require.NoError(t, err, "should make directory read-only")

	err = mgr.BackupFile(ctx, "test.txt")
	require.Error(t, err, "should error on read-only directory")

	err = os.Chmod(tmpDir, 0700)
	require.NoError(t, err, "should restore directory permissions")

	// Test restore with missing backup but existing target
	err = mgr.WriteFile(ctx, "test2.txt", []byte("original"))
	require.NoError(t, err, "should create test file")

	err = mgr.RestoreFile(ctx, "test2.txt")
	require.Error(t, err, "should error on missing backup")

	// Test restore with corrupted backup
	err = mgr.WriteFile(ctx, "test3.txt", []byte("original"))
	require.NoError(t, err, "should create test file")

	err = mgr.BackupFile(ctx, "test3.txt")
	require.NoError(t, err, "should create backup")

	// Corrupt the backup file by making it unreadable
	backupPath := filepath.Join(tmpDir, "test3.txt.bak")
	err = os.Chmod(backupPath, 0000)
	require.NoError(t, err, "should make backup unreadable")

	err = mgr.RestoreFile(ctx, "test3.txt")
	require.Error(t, err, "should error on unreadable backup")

	// Restore permissions for cleanup
	err = os.Chmod(backupPath, 0600)
	require.NoError(t, err, "should restore backup permissions")
}

// 🧪 TestWriteFileEdgeCases tests edge cases for file writing operations
func TestWriteFileEdgeCases(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	mgr := New(tmpDir, &logger)
	ctx := context.Background()

	// Test writing to a path with invalid parent directory name
	err := mgr.WriteFile(ctx, string([]byte{0x00})+"/test.txt", []byte("test"))
	require.Error(t, err, "should error on invalid parent directory")

	// Test writing to a path that exists as a directory
	err = mgr.CreateDir(ctx, "dir")
	require.NoError(t, err, "should create directory")

	err = mgr.WriteFile(ctx, "dir", []byte("test"))
	require.Error(t, err, "should error when writing to directory path")

	// Test writing with read-only parent directory
	err = mgr.CreateDir(ctx, "readonly")
	require.NoError(t, err, "should create directory")

	err = os.Chmod(filepath.Join(tmpDir, "readonly"), 0500)
	require.NoError(t, err, "should make directory read-only")

	err = mgr.WriteFile(ctx, "readonly/test.txt", []byte("test"))
	require.Error(t, err, "should error on read-only directory")

	err = os.Chmod(filepath.Join(tmpDir, "readonly"), 0700)
	require.NoError(t, err, "should restore directory permissions")
}

// 🧪 TestCopyFilePermissions tests file permission handling during copy operations
func TestCopyFilePermissions(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	mgr := New(tmpDir, &logger)

	// Create source file with specific permissions
	srcPath := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(srcPath, []byte("test"), 0600)
	require.NoError(t, err, "should create source file")

	// Copy to destination
	dstPath := filepath.Join(tmpDir, "dest.txt")
	err = mgr.CopyFile(srcPath, dstPath)
	require.NoError(t, err, "should copy file")

	// Verify destination exists
	_, err = os.Stat(dstPath)
	require.NoError(t, err, "destination file should exist")

	// Test copying a read-only file
	err = os.Chmod(srcPath, 0400)
	require.NoError(t, err, "should make source read-only")

	dstPath2 := filepath.Join(tmpDir, "dest2.txt")
	err = mgr.CopyFile(srcPath, dstPath2)
	require.NoError(t, err, "should copy read-only file")

	// Test copying to an existing file
	err = mgr.CopyFile(srcPath, dstPath2)
	require.NoError(t, err, "should overwrite existing file")
}

func TestReadFileEdgeCases(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)

	// Test reading non-existent file
	_, err := mgr.ReadFile(ctx, "nonexistent.txt")
	require.Error(t, err, "should error on non-existent file")

	// Test reading directory
	err = mgr.CreateDir(ctx, "testdir")
	require.NoError(t, err, "should create directory")
	_, err = mgr.ReadFile(ctx, "testdir")
	require.Error(t, err, "should error on reading directory")

	// Test reading unreadable file
	err = mgr.WriteFile(ctx, "unreadable.txt", []byte("test"))
	require.NoError(t, err, "should create file")
	err = os.Chmod(filepath.Join(tmpDir, "unreadable.txt"), 0000)
	require.NoError(t, err, "should make file unreadable")
	_, err = mgr.ReadFile(ctx, "unreadable.txt")
	require.Error(t, err, "should error on unreadable file")
}

func TestRestoreFileEdgeCases(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)

	// Test restoring non-existent backup
	err := mgr.RestoreFile(ctx, "nonexistent.txt")
	require.Error(t, err, "should error on non-existent backup")

	// Test restoring with unreadable original file
	err = mgr.WriteFile(ctx, "original.txt", []byte("test"))
	require.NoError(t, err, "should create file")
	err = mgr.BackupFile(ctx, "original.txt")
	require.NoError(t, err, "should create backup")
	err = os.Chmod(filepath.Join(tmpDir, "original.txt"), 0000)
	require.NoError(t, err, "should make file unreadable")
	err = mgr.RestoreFile(ctx, "original.txt")
	require.Error(t, err, "should error on unreadable original file")
}

func TestCopyFileComplexCases(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)

	// Test copying to existing file with no write permission
	err := mgr.WriteFile(ctx, "source.txt", []byte("test"))
	require.NoError(t, err, "should create source file")
	err = mgr.WriteFile(ctx, "dest.txt", []byte("original"))
	require.NoError(t, err, "should create destination file")
	err = os.Chmod(filepath.Join(tmpDir, "dest.txt"), 0444)
	require.NoError(t, err, "should make destination read-only")
	err = mgr.CopyFile(filepath.Join(tmpDir, "source.txt"), filepath.Join(tmpDir, "dest.txt"))
	require.Error(t, err, "should error on read-only destination")

	// Test copying with source file that can't be opened for reading
	srcPath := filepath.Join(tmpDir, "locked.txt")
	err = os.WriteFile(srcPath, []byte("locked"), 0000)
	require.NoError(t, err, "should create unreadable file")
	err = mgr.CopyFile(srcPath, filepath.Join(tmpDir, "dest2.txt"))
	require.Error(t, err, "should error on unreadable source file")

	// Restore permissions for cleanup
	err = os.Chmod(srcPath, 0600)
	require.NoError(t, err, "should restore file permissions")
}

func TestRestoreFileComplexCases(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)

	// Test restoring with backup file that can't be opened
	err := mgr.WriteFile(ctx, "test.txt", []byte("original"))
	require.NoError(t, err, "should create test file")
	err = mgr.BackupFile(ctx, "test.txt")
	require.NoError(t, err, "should create backup")

	// Make backup file unreadable
	backupPath := filepath.Join(tmpDir, "test.txt.bak")
	err = os.Chmod(backupPath, 0000)
	require.NoError(t, err, "should make backup unreadable")

	err = mgr.RestoreFile(ctx, "test.txt")
	require.Error(t, err, "should error on unreadable backup")

	// Restore permissions for cleanup
	err = os.Chmod(backupPath, 0600)
	require.NoError(t, err, "should restore backup permissions")

	// Test restoring with backup file that exists but is empty
	err = os.WriteFile(backupPath, []byte{}, 0644)
	require.NoError(t, err, "should create empty backup")

	err = mgr.RestoreFile(ctx, "test.txt")
	require.NoError(t, err, "should handle empty backup file")

	// Test restoring with backup file that is a directory
	backupDir := filepath.Join(tmpDir, "dir.txt.bak")
	err = os.Mkdir(backupDir, 0755)
	require.NoError(t, err, "should create backup directory")

	err = mgr.WriteFile(ctx, "dir.txt", []byte("test"))
	require.NoError(t, err, "should create test file")

	err = mgr.RestoreFile(ctx, "dir.txt")
	require.Error(t, err, "should error when backup is a directory")
}

func TestCopyFileErrorCases(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)

	// Test copying from a directory
	err := mgr.CreateDir(ctx, "srcdir")
	require.NoError(t, err, "should create source directory")
	err = mgr.CopyFile(filepath.Join(tmpDir, "srcdir"), filepath.Join(tmpDir, "dest.txt"))
	require.Error(t, err, "should error when source is a directory")

	// Test copying to a directory that exists
	err = mgr.CreateDir(ctx, "destdir")
	require.NoError(t, err, "should create destination directory")
	err = mgr.WriteFile(ctx, "source.txt", []byte("test"))
	require.NoError(t, err, "should create source file")
	err = mgr.CopyFile(filepath.Join(tmpDir, "source.txt"), filepath.Join(tmpDir, "destdir"))
	require.Error(t, err, "should error when destination is a directory")

	// Test copying with invalid source path
	err = mgr.CopyFile(string([]byte{0x00}), filepath.Join(tmpDir, "dest.txt"))
	require.Error(t, err, "should error on invalid source path")

	// Test copying with invalid destination path
	err = mgr.CopyFile(filepath.Join(tmpDir, "source.txt"), string([]byte{0x00}))
	require.Error(t, err, "should error on invalid destination path")

	// Test copying to a destination in a read-only directory
	destDir := filepath.Join(tmpDir, "readonly")
	err = os.MkdirAll(destDir, 0500)
	require.NoError(t, err, "should create read-only directory")
	err = mgr.CopyFile(filepath.Join(tmpDir, "source.txt"), filepath.Join(destDir, "dest.txt"))
	require.Error(t, err, "should error when destination directory is read-only")

	// Test copying to a destination with parent directory that doesn't exist
	err = mgr.CopyFile(filepath.Join(tmpDir, "source.txt"), filepath.Join(tmpDir, "nonexistent/dest.txt"))
	require.NoError(t, err, "should create parent directories")
	content, err := os.ReadFile(filepath.Join(tmpDir, "nonexistent/dest.txt"))
	require.NoError(t, err, "should read copied file")
	assert.Equal(t, []byte("test"), content, "should copy file content correctly")

	// Restore directory permissions for cleanup
	err = os.Chmod(destDir, 0700)
	require.NoError(t, err, "should restore directory permissions")
}
