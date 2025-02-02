// Copyright 2025 walteh LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/tozd/go/errors"
)

func TestTarballFunctions(t *testing.T) {
	// Create mock provider
	mock := NewMockProvider(t)
	mock.AddFile("test.txt", []byte("test content"))
	mock.AddFile("dir/nested.txt", []byte("nested content"))

	t.Run("test_get_file", func(t *testing.T) {
		args := Source{
			Repo: "github.com/org/repo",
			Ref:  "main",
			Path: "path/to/files",
		}

		data, err := GetFileFromTarball(context.Background(), mock, args)
		require.NoError(t, err, "getting file should succeed")
		assert.Equal(t, []byte{0x1f, 0x8b}, data[0:2], "should be gzipped data")
	})

	t.Run("test_get_nested_file", func(t *testing.T) {
		args := Source{
			Repo: "github.com/org/repo",
			Ref:  "main",
			Path: "path/to/files",
		}

		data, err := GetFileFromTarball(context.Background(), mock, args)
		require.NoError(t, err, "getting nested file should succeed")
		assert.Equal(t, []byte{0x1f, 0x8b}, data[0:2], "should be gzipped data")
	})

	t.Run("test_file_not_found", func(t *testing.T) {
		args := Source{
			Repo: "github.com/org/repo",
			Ref:  "main",
			Path: "/invalid/path",
		}

		_, err := GetFileFromTarball(context.Background(), mock, args)
		require.Error(t, err, "getting nonexistent file should fail")
		assert.Contains(t, err.Error(), "invalid path")
	})

	t.Run("test_invalid_cache_dir", func(t *testing.T) {
		args := Source{
			Repo: "github.com/org/repo",
			Ref:  "main",
			Path: "/invalid/path",
		}

		_, err := GetFileFromTarball(context.Background(), mock, args)
		require.Error(t, err, "getting file with invalid cache dir should fail")
		assert.Contains(t, err.Error(), "invalid path")
	})

	t.Run("test_invalid_tag_404", func(t *testing.T) {
		mock := NewMockErrorProvider(t)
		mock.shouldReturn404 = true

		args := Source{
			Repo: "github.com/org/repo",
			Ref:  "v999.999.999", // Non-existent tag
			Path: "path/to/files",
		}

		_, err := GetFileFromTarball(context.Background(), mock, args)
		require.Error(t, err, "getting file with invalid tag should fail")
		assert.Contains(t, err.Error(), "invalid tag or reference", "error should indicate invalid tag")
	})
}

type MockErrorProvider struct {
	*MockProvider
	shouldReturn404 bool
}

func NewMockErrorProvider(t *testing.T) *MockErrorProvider {
	return &MockErrorProvider{
		MockProvider: NewMockProvider(t),
	}
}

func (m *MockErrorProvider) GetArchiveUrl(ctx context.Context, args Source) (string, error) {
	if m.shouldReturn404 {
		// Create a temporary file with 404 content
		f, err := os.CreateTemp("", "mock-404-*.tar.gz")
		if err != nil {
			return "", errors.Errorf("creating temp file: %s", err)
		}
		defer f.Close()

		// Write 404 content
		if _, err := f.WriteString("404: Not Found"); err != nil {
			return "", errors.Errorf("writing 404 content: %s", err)
		}

		return "file://" + f.Name(), nil
	}
	return m.MockProvider.GetArchiveUrl(ctx, args)
}
