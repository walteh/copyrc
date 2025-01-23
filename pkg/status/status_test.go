package status_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/gen/mockery"
	"github.com/walteh/copyrc/pkg/status"
)

// 🛠️ setupTestManager creates a new Manager with a mock formatter
func setupTestManager(t *testing.T, tmpDir string) (*status.Manager, *mockery.MockFileFormatter_status) {
	formatter := mockery.NewMockFileFormatter_status(t)
	mgr := status.NewManager(tmpDir, formatter)
	return mgr, formatter
}

// 🧪 TestDefaultFileFormatter tests the default formatter implementation
func TestDefaultFileFormatter(t *testing.T) {
	formatter := status.NewDefaultFileFormatter()

	t.Run("new_file", func(t *testing.T) {
		result := formatter.FormatFileStatus("test.txt", status.StatusNew, nil)
		assert.Equal(t, "✨ Created test.txt", result)
	})

	t.Run("modified_file", func(t *testing.T) {
		result := formatter.FormatFileStatus("test.txt", status.StatusModified, nil)
		assert.Equal(t, "📝 Modified test.txt", result)
	})

	t.Run("removed_file", func(t *testing.T) {
		result := formatter.FormatFileStatus("test.txt", status.StatusDeleted, nil)
		assert.Equal(t, "🗑️ Removed test.txt", result)
	})

	t.Run("unchanged_file", func(t *testing.T) {
		result := formatter.FormatFileStatus("test.txt", status.StatusUnchanged, nil)
		assert.Equal(t, "👍 Unchanged test.txt", result)
	})

	t.Run("error_status", func(t *testing.T) {
		result := formatter.FormatFileStatus("error.txt", status.StatusUnknown, nil)
		assert.Equal(t, "❌ Failed error.txt", result)
	})
}

// 🧪 TestProgressFormatting tests progress formatting
func TestProgressFormatting(t *testing.T) {
	formatter := status.NewDefaultFileFormatter()

	t.Run("zero_progress", func(t *testing.T) {
		result := formatter.FormatProgress(0, 10)
		assert.Equal(t, "⏳ Progress: 0/10 (0%)", result)
	})

	t.Run("half_progress", func(t *testing.T) {
		result := formatter.FormatProgress(5, 10)
		assert.Equal(t, "⏳ Progress: 5/10 (50%)", result)
	})

	t.Run("complete", func(t *testing.T) {
		result := formatter.FormatProgress(10, 10)
		assert.Equal(t, "✅ Progress: 10/10 (100%)", result)
	})

	t.Run("zero_total", func(t *testing.T) {
		result := formatter.FormatProgress(0, 0)
		assert.Equal(t, "✅ Progress: 0/0 (0%)", result)
	})

	t.Run("zero_total_with_current", func(t *testing.T) {
		result := formatter.FormatProgress(5, 0)
		assert.Equal(t, "✅ Progress: 5/0 (0%)", result)
	})
}

// 🧪 TestErrorFormatting tests error formatting
func TestErrorFormatting(t *testing.T) {
	formatter := status.NewDefaultFileFormatter()

	t.Run("with_error", func(t *testing.T) {
		result := formatter.FormatError(assert.AnError)
		assert.Equal(t, "❌ Error: assert.AnError general error for testing", result)
	})

	t.Run("nil_error", func(t *testing.T) {
		result := formatter.FormatError(nil)
		assert.Equal(t, "", result)
	})
}

// 🧪 TestFileOperations tests file operations
func TestFileOperations(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, formatter := setupTestManager(t, tmpDir)
	ctx := context.Background()

	// Test tracking a new file
	entry := &status.FileEntry{
		Status: status.StatusNew,
		Metadata: map[string]string{
			"size": "100",
			"mode": "0644",
		},
	}

	formatter.EXPECT().FormatFileStatus("test.txt", status.StatusNew, entry.Metadata).Return("✨ Created test.txt")
	mgr.UpdateStatus(ctx, "test.txt", status.StatusNew, entry)

	// Test file existence
	exists, err := mgr.FileExists(ctx, filepath.Join(tmpDir, "test.txt"))
	require.NoError(t, err, "checking file existence should not error")
	assert.False(t, exists, "file should not exist yet")

	// Test writing file
	content := []byte("test content")
	err = mgr.WriteFile(ctx, "test.txt", content)
	require.NoError(t, err, "writing file should succeed")

	// Test reading file
	readContent, err := mgr.ReadFile(ctx, "test.txt")
	require.NoError(t, err, "reading file should succeed")
	assert.Equal(t, content, readContent, "file content should match")

	// Test deleting file
	err = mgr.DeleteFile(ctx, "test.txt")
	require.NoError(t, err, "deleting file should succeed")
}

// 🧪 TestCopyFileComplexCases tests complex scenarios for file copying
func TestCopyFileComplexCases(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, _ := setupTestManager(t, tmpDir)

	// Test copying to existing file with no write permission
	srcPath := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(srcPath, []byte("test"), 0644)
	require.NoError(t, err, "creating source file should succeed")

	destPath := filepath.Join(tmpDir, "dest.txt")
	err = os.WriteFile(destPath, []byte("original"), 0444)
	require.NoError(t, err, "creating read-only destination should succeed")

	err = mgr.CopyFile(srcPath, destPath)
	require.Error(t, err, "copying to read-only file should fail")

	// Test copying from non-existent file
	err = mgr.CopyFile("nonexistent.txt", destPath)
	require.Error(t, err, "copying non-existent file should fail")

	// Test copying to invalid path
	err = mgr.CopyFile(srcPath, string([]byte{0x00}))
	require.Error(t, err, "copying to invalid path should fail")
}

// 🧪 TestRestoreFileComplexCases tests complex scenarios for file restoration
func TestRestoreFileComplexCases(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	mgr, _ := setupTestManager(t, tmpDir)

	// Test restoring with backup file that can't be opened
	err := mgr.WriteFile(ctx, "test.txt", []byte("original"))
	require.NoError(t, err, "creating test file should succeed")

	err = mgr.BackupFile(ctx, "test.txt")
	require.NoError(t, err, "creating backup should succeed")

	backupPath := filepath.Join(tmpDir, "test.txt.bak")
	err = os.Chmod(backupPath, 0000)
	require.NoError(t, err, "making backup unreadable should succeed")

	err = mgr.RestoreFile(ctx, "test.txt")
	require.Error(t, err, "restoring from unreadable backup should fail")

	// Cleanup
	err = os.Chmod(backupPath, 0644)
	require.NoError(t, err, "restoring backup permissions should succeed")
}

// 🧪 TestCopyFileErrorCases tests error cases for file copying
func TestCopyFileErrorCases(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, _ := setupTestManager(t, tmpDir)

	// Test copying from a directory
	dirPath := filepath.Join(tmpDir, "dir")
	err := os.Mkdir(dirPath, 0755)
	require.NoError(t, err, "creating directory should succeed")

	destPath := filepath.Join(tmpDir, "dest.txt")
	err = mgr.CopyFile(dirPath, destPath)
	require.Error(t, err, "copying from directory should fail")

	// Test copying to a directory
	srcPath := filepath.Join(tmpDir, "source.txt")
	err = os.WriteFile(srcPath, []byte("test"), 0644)
	require.NoError(t, err, "creating source file should succeed")

	err = mgr.CopyFile(srcPath, dirPath)
	require.Error(t, err, "copying to directory should fail")
}

// 🧪 TestErrorHandling tests error handling
func TestErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, formatter := setupTestManager(t, tmpDir)
	ctx := context.Background()

	// Test reading non-existent file
	_, err := mgr.ReadFile(ctx, "nonexistent.txt")
	require.Error(t, err, "reading non-existent file should error")

	// Test tracking file with error
	entry := &status.FileEntry{
		Status: status.StatusUnknown,
		Metadata: map[string]string{
			"error": "file not found",
		},
	}

	formatter.EXPECT().FormatFileStatus("error.txt", status.StatusUnknown, entry.Metadata).Return("❌ Failed error.txt")
	mgr.UpdateStatus(ctx, "error.txt", status.StatusUnknown, entry)
}

// 🧪 TestListFiles tests listing files
func TestListFiles(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, formatter := setupTestManager(t, tmpDir)
	ctx := context.Background()

	// Create test files
	files := []struct {
		path   string
		status status.FileStatus
	}{
		{"file1.txt", status.StatusNew},
		{"file2.txt", status.StatusModified},
		{"file3.txt", status.StatusUnchanged},
	}

	for _, f := range files {
		entry := &status.FileEntry{
			Status: f.status,
			Metadata: map[string]string{
				"size": "100",
				"mode": "0644",
			},
		}
		formatter.EXPECT().FormatFileStatus(f.path, f.status, entry.Metadata).Return("✨ Created " + f.path)
		mgr.UpdateStatus(ctx, f.path, f.status, entry)
	}

	// List files
	fileList, err := mgr.ListFiles(ctx)
	require.NoError(t, err, "listing files should succeed")
	assert.Len(t, fileList, len(files), "should list all tracked files")
}

// 🧪 TestMarkFileIgnored tests the MarkFileIgnored functionality
func TestMarkFileIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, formatter := setupTestManager(t, tmpDir)
	ctx := context.Background()

	// Set up expectations
	formatter.EXPECT().FormatFileStatus("test.txt", status.StatusUnchanged, map[string]string{
		"reason":  "test reason",
		"ignored": "true",
	}).Return("👍 Unchanged test.txt")

	// Mark file as ignored
	mgr.MarkFileIgnored(ctx, "test.txt", "test reason")

	// Get tracked files
	files, err := mgr.ListFiles(ctx)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Verify status
	assert.Equal(t, status.StatusUnchanged, files[0].Status)
	assert.Equal(t, "test.txt", files[0].Path)
}

// 🧪 TestUpdateFileContent tests the UpdateFileContent functionality
func TestUpdateFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, formatter := setupTestManager(t, tmpDir)
	ctx := context.Background()

	// Test file in nested directory
	nestedPath := filepath.Join("deep", "nested", "test.txt")
	content := []byte("test content")
	metadata := map[string]string{
		"commit_hash": "abc123",
		"permalink":   "https://test.com/file",
		"source":      "test/repo",
	}

	// Calculate expected hash
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Set up expectations
	formatter.EXPECT().FormatFileStatus(nestedPath, status.StatusModified, map[string]string{
		"hash":        hash,
		"size":        fmt.Sprintf("%d", len(content)),
		"mode":        "0644",
		"commit_hash": "abc123",
		"permalink":   "https://test.com/file",
		"source":      "test/repo",
	}).Return("📝 Modified " + nestedPath)

	// Update file content
	err := mgr.UpdateFileContent(ctx, nestedPath, content, metadata)
	require.NoError(t, err)

	// Verify file content
	readContent, err := mgr.ReadFile(ctx, nestedPath)
	require.NoError(t, err)
	assert.Equal(t, content, readContent)

	// Verify directory was created
	dirInfo, err := os.Stat(filepath.Join(tmpDir, "deep", "nested"))
	require.NoError(t, err)
	assert.True(t, dirInfo.IsDir())
	assert.Equal(t, os.FileMode(0755), dirInfo.Mode()&os.ModePerm)

	// Get tracked files
	files, err := mgr.ListFiles(ctx)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Verify status
	assert.Equal(t, status.StatusModified, files[0].Status)
	assert.Equal(t, nestedPath, files[0].Path)
	assert.Equal(t, int64(len(content)), files[0].Size)
}
