package state

// func TestValidation(t *testing.T) {
// 	ctx := setupTestLogger(t)

// 	t.Run("validates_clean_state", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		// Empty state should validate
// 		err = state.ValidateLocalState(ctx)
// 		assert.NoError(t, err)

// 		consistent, err := state.IsConsistent(ctx)
// 		assert.NoError(t, err)
// 		assert.True(t, consistent)
// 	})

// 	t.Run("validates_existing_files", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		// Create a test file
// 		content := []byte("test content")
// 		testFile := filepath.Join(dir, "test.copy.txt")
// 		err = os.WriteFile(testFile, content, 0600)
// 		require.NoError(t, err)

// 		// Add file to state
// 		hash, err := state.hashFile(testFile)
// 		require.NoError(t, err)

// 		state.file.RemoteTextFiles = []RemoteTextFile{{
// 			LocalPath:   testFile,
// 			ContentHash: hash,
// 		}}

// 		// Should validate
// 		err = state.ValidateLocalState(ctx)
// 		assert.NoError(t, err)

// 		consistent, err := state.IsConsistent(ctx)
// 		assert.NoError(t, err)
// 		assert.True(t, consistent)
// 	})

// 	t.Run("detects_missing_files", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		// Add non-existent file to state
// 		state.file.RemoteTextFiles = []RemoteTextFile{{
// 			LocalPath:   filepath.Join(dir, "missing.copy.txt"),
// 			ContentHash: "abc123",
// 		}}

// 		// Should fail validation
// 		err = state.ValidateLocalState(ctx)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "does not exist")

// 		consistent, err := state.IsConsistent(ctx)
// 		assert.NoError(t, err)
// 		assert.False(t, consistent)
// 	})

// 	t.Run("detects_modified_files", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		// Create a test file
// 		content := []byte("test content")
// 		testFile := filepath.Join(dir, "test.copy.txt")
// 		err = os.WriteFile(testFile, content, 0600)
// 		require.NoError(t, err)

// 		// Add file to state with wrong hash
// 		state.file.RemoteTextFiles = []RemoteTextFile{{
// 			LocalPath:   testFile,
// 			ContentHash: "wrong hash",
// 		}}

// 		// Should fail validation
// 		err = state.ValidateLocalState(ctx)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "hash mismatch")

// 		consistent, err := state.IsConsistent(ctx)
// 		assert.NoError(t, err)
// 		assert.False(t, consistent)
// 	})

// 	t.Run("cleanup_removes_orphaned_files", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		// Create some files
// 		files := []string{"keep.copy.txt", "orphan1.copy.txt", "orphan2.copy.txt"}
// 		for _, name := range files {
// 			err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0600)
// 			require.NoError(t, err)
// 		}

// 		// Add only one file to state
// 		state.file.RemoteTextFiles = []RemoteTextFile{{
// 			LocalPath: filepath.Join(dir, "keep.copy.txt"),
// 		}}

// 		// Clean up should remove orphaned files
// 		err = state.CleanupOrphanedFiles(ctx)
// 		assert.NoError(t, err)

// 		// Verify only tracked file remains
// 		for _, name := range files {
// 			exists := true
// 			if _, err := os.Stat(filepath.Join(dir, name)); os.IsNotExist(err) {
// 				exists = false
// 			}
// 			if name == "keep.copy.txt" {
// 				assert.True(t, exists, "tracked file should exist")
// 			} else {
// 				assert.False(t, exists, "orphaned file should be removed")
// 			}
// 		}
// 	})

// 	t.Run("validates_schema_version", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		// Set invalid schema version
// 		state.file.SchemaVersion = "0.9"

// 		// Should fail validation
// 		err = state.ValidateLocalState(ctx)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "invalid schema version")
// 	})

// 	t.Run("validates_file_suffixes", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		validCopy := filepath.Join(dir, "test.copy.txt")
// 		validPatch := filepath.Join(dir, "test.patch.txt")
// 		invalid := filepath.Join(dir, "test.txt")

// 		// Create test files with known content
// 		content := []byte("test")
// 		for _, path := range []string{validCopy, validPatch, invalid} {
// 			err := os.WriteFile(path, content, 0600)
// 			require.NoError(t, err)
// 		}

// 		// Expected hash for content "test"
// 		expectedHash := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"

// 		// Test valid copy suffix
// 		state.file.RemoteTextFiles = []RemoteTextFile{{
// 			LocalPath:   validCopy,
// 			ContentHash: expectedHash,
// 		}}
// 		err = state.ValidateLocalState(ctx)
// 		assert.NoError(t, err)

// 		// Test valid patch suffix
// 		state.file.RemoteTextFiles = []RemoteTextFile{{
// 			LocalPath:   validPatch,
// 			ContentHash: expectedHash,
// 		}}
// 		err = state.ValidateLocalState(ctx)
// 		assert.NoError(t, err)

// 		// Test invalid suffix
// 		state.file.RemoteTextFiles = []RemoteTextFile{{
// 			LocalPath: invalid,
// 		}}
// 		err = state.ValidateLocalState(ctx)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "invalid file suffix")
// 	})

// 	t.Run("validates_patch_files", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		// Create main file and patch with known content
// 		mainContent := []byte("test")
// 		patchContent := []byte("patch")
// 		mainFile := filepath.Join(dir, "test.copy.txt")
// 		patchFile := filepath.Join(dir, "test.patch.txt")
// 		err = os.WriteFile(mainFile, mainContent, 0600)
// 		require.NoError(t, err)
// 		err = os.WriteFile(patchFile, patchContent, 0600)
// 		require.NoError(t, err)

// 		// Expected hash for content "test"
// 		expectedHash := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"

// 		// Add file with patch to state
// 		state.file.RemoteTextFiles = []RemoteTextFile{{
// 			LocalPath:   mainFile,
// 			ContentHash: expectedHash,
// 			IsPatched:   true,
// 			Patch: &Patch{
// 				PatchPath: patchFile,
// 			},
// 		}}

// 		// Should validate
// 		err = state.ValidateLocalState(ctx)
// 		assert.NoError(t, err)

// 		// Remove patch file
// 		err = os.Remove(patchFile)
// 		require.NoError(t, err)

// 		// Should fail validation
// 		err = state.ValidateLocalState(ctx)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "does not exist")
// 	})
// }

// func TestHashFile(t *testing.T) {
// 	ctx := setupTestLogger(t)

// 	t.Run("computes_correct_hash", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		// Create test file with known content
// 		content := []byte("test content")
// 		testFile := filepath.Join(dir, "test.txt")
// 		err = os.WriteFile(testFile, content, 0600)
// 		require.NoError(t, err)

// 		// Compute hash
// 		hash, err := state.hashFile(testFile)
// 		require.NoError(t, err)

// 		// Expected SHA-256 hash of "test content"
// 		expected := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"
// 		assert.Equal(t, expected, hash)
// 	})

// 	t.Run("handles_missing_file", func(t *testing.T) {
// 		dir := setupTestDir(t)
// 		state, err := New(dir)
// 		require.NoError(t, err)
// 		err = state.Load(ctx)
// 		require.NoError(t, err)

// 		// Try to hash non-existent file
// 		_, err = state.hashFile(filepath.Join(dir, "missing.txt"))
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "opening file")
// 	})
// }

// func TestValidateLocalState(t *testing.T) {
// 	// Setup test directory
// 	dir := t.TempDir()

// 	// Create state
// 	state, err := New(dir)
// 	require.NoError(t, err, "creating state")

// 	// Create test files
// 	copyFile := filepath.Join(dir, "test.copy.txt")
// 	patchFile := filepath.Join(dir, "test.patch.txt")
// 	archiveFile := filepath.Join(dir, "archive.tar.gz")

// 	err = os.WriteFile(copyFile, []byte("Hello World"), 0644)
// 	require.NoError(t, err, "writing copy file")

// 	err = os.WriteFile(patchFile, []byte("patch content"), 0644)
// 	require.NoError(t, err, "writing patch file")

// 	err = os.WriteFile(archiveFile, []byte("archive content"), 0644)
// 	require.NoError(t, err, "writing archive file")

// 	// Setup state data
// 	state.file.RemoteTextFiles = []RemoteTextFile{
// 		{
// 			LocalPath:   copyFile,
// 			ContentHash: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e", // Hello World
// 			IsPatched:   true,
// 			Patch: &Patch{
// 				PatchPath: patchFile,
// 			},
// 		},
// 	}

// 	state.file.Repositories = []Repository{
// 		{
// 			Name: "test/repo",
// 			Release: &Release{
// 				Archive: &Archive{
// 					LocalPath: archiveFile,
// 					Hash:      "fa868b2818c90263b5c2c8e056180232a6f3c34547ca49b7f3ca10599a52db3d", // archive content
// 				},
// 			},
// 		},
// 	}

// 	// Test cases
// 	tests := []struct {
// 		name        string
// 		setup       func()
// 		wantError   string
// 		description string
// 	}{
// 		{
// 			name:        "valid_state",
// 			description: "All files exist and match their recorded state",
// 		},
// 		{
// 			name: "missing_copy_file",
// 			setup: func() {
// 				require.NoError(t, os.Remove(copyFile))
// 			},
// 			wantError:   "file does not exist",
// 			description: "Copy file is missing",
// 		},
// 		{
// 			name: "modified_copy_file",
// 			setup: func() {
// 				require.NoError(t, os.WriteFile(copyFile, []byte("Modified"), 0644))
// 			},
// 			wantError:   "content hash mismatch",
// 			description: "Copy file has been modified",
// 		},
// 		{
// 			name: "missing_patch_file",
// 			setup: func() {
// 				require.NoError(t, os.Remove(patchFile))
// 			},
// 			wantError:   "patch file does not exist",
// 			description: "Patch file is missing",
// 		},
// 		{
// 			name: "missing_archive",
// 			setup: func() {
// 				require.NoError(t, os.Remove(archiveFile))
// 			},
// 			wantError:   "archive file does not exist",
// 			description: "Archive file is missing",
// 		},
// 		{
// 			name: "modified_archive",
// 			setup: func() {
// 				require.NoError(t, os.WriteFile(archiveFile, []byte("modified archive"), 0644))
// 			},
// 			wantError:   "archive hash mismatch",
// 			description: "Archive has been modified",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Reset files
// 			err = os.WriteFile(copyFile, []byte("Hello World"), 0644)
// 			require.NoError(t, err, "resetting copy file")

// 			err = os.WriteFile(patchFile, []byte("patch content"), 0644)
// 			require.NoError(t, err, "resetting patch file")

// 			err = os.WriteFile(archiveFile, []byte("archive content"), 0644)
// 			require.NoError(t, err, "resetting archive file")

// 			// Run setup if provided
// 			if tt.setup != nil {
// 				tt.setup()
// 			}

// 			// Run validation
// 			err := state.ValidateLocalState(context.Background())
// 			if tt.wantError != "" {
// 				assert.ErrorContains(t, err, tt.wantError)
// 				return
// 			}
// 			assert.NoError(t, err)
// 		})
// 	}
// }

// func TestCleanupOrphanedFiles(t *testing.T) {
// 	// Setup test directory
// 	dir := t.TempDir()

// 	// Create state
// 	state, err := New(dir)
// 	require.NoError(t, err, "creating state")

// 	// Create test files
// 	files := []string{
// 		"keep.copy.txt",     // tracked file
// 		"orphan1.copy.txt",  // untracked file
// 		"orphan2.patch.txt", // untracked file
// 		"unrelated.txt",     // unrelated file
// 	}

// 	for _, f := range files {
// 		err := os.WriteFile(filepath.Join(dir, f), []byte("content"), 0644)
// 		require.NoError(t, err, "writing test file")
// 	}

// 	// Setup state data
// 	state.file.RemoteTextFiles = []RemoteTextFile{
// 		{
// 			LocalPath: filepath.Join(dir, "keep.copy.txt"),
// 		},
// 	}

// 	// Run cleanup
// 	err = state.CleanupOrphanedFiles(context.Background())
// 	require.NoError(t, err, "cleaning up orphaned files")

// 	// Verify results
// 	assert.FileExists(t, filepath.Join(dir, "keep.copy.txt"), "tracked file should exist")
// 	assert.FileExists(t, filepath.Join(dir, "unrelated.txt"), "unrelated file should exist")
// 	assert.NoFileExists(t, filepath.Join(dir, "orphan1.copy.txt"), "orphaned copy file should be removed")
// 	assert.NoFileExists(t, filepath.Join(dir, "orphan2.patch.txt"), "orphaned patch file should be removed")
// }
