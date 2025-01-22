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

package github

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/pkg/config"
)

func TestParseRepo(t *testing.T) {
	tests := []struct {
		name        string
		repo        string
		wantOwner   string
		wantName    string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid_repo",
			repo:      "github.com/walteh/copyrc",
			wantOwner: "walteh",
			wantName:  "copyrc",
		},
		{
			name:      "valid_repo_with_https",
			repo:      "https://github.com/walteh/copyrc",
			wantOwner: "walteh",
			wantName:  "copyrc",
		},
		{
			name:        "invalid_repo",
			repo:        "invalid",
			wantErr:     true,
			errContains: "invalid GitHub repository URL",
		},
	}

	// Create provider with mock token
	os.Setenv("GITHUB_TOKEN", "mock_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	ctx := zerolog.New(os.Stderr).WithContext(context.Background())
	p, err := New(ctx)
	require.NoError(t, err, "creating provider should succeed")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := p.(*Provider).parseRepo(tt.repo)
			if tt.wantErr {
				require.Error(t, err, "parseRepo should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "parseRepo should succeed")
			assert.Equal(t, tt.wantOwner, owner, "owner should match")
			assert.Equal(t, tt.wantName, name, "name should match")
		})
	}
}

func TestGetSourceInfo(t *testing.T) {
	tests := []struct {
		name        string
		args        config.ProviderArgs
		commitHash  string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_source_info",
			args: config.ProviderArgs{
				Repo: "github.com/walteh/copyrc",
				Ref:  "main",
			},
			commitHash: "1234567890abcdef",
			want:       "github.com/walteh/copyrc@1234567",
		},
	}

	// Create provider with mock token
	os.Setenv("GITHUB_TOKEN", "mock_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	ctx := zerolog.New(os.Stderr).WithContext(context.Background())
	p, err := New(ctx)
	require.NoError(t, err, "creating provider should succeed")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.GetSourceInfo(ctx, tt.args, tt.commitHash)
			if tt.wantErr {
				require.Error(t, err, "GetSourceInfo should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "GetSourceInfo should succeed")
			assert.Equal(t, tt.want, got, "source info should match")
		})
	}
}

func TestGetPermalink(t *testing.T) {
	tests := []struct {
		name        string
		args        config.ProviderArgs
		commitHash  string
		file        string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_permalink",
			args: config.ProviderArgs{
				Repo: "github.com/walteh/copyrc",
				Ref:  "main",
				Path: "pkg/provider",
			},
			commitHash: "1234567890abcdef",
			file:       "github.go",
			want:       "https://github.com/walteh/copyrc/blob/1234567890abcdef/pkg/provider/github.go",
		},
		{
			name: "invalid_repo",
			args: config.ProviderArgs{
				Repo: "invalid",
				Ref:  "main",
				Path: "pkg/provider",
			},
			commitHash:  "1234567890abcdef",
			file:        "github.go",
			wantErr:     true,
			errContains: "invalid GitHub repository URL",
		},
	}

	// Create provider with mock token
	os.Setenv("GITHUB_TOKEN", "mock_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	ctx := zerolog.New(os.Stderr).WithContext(context.Background())
	p, err := New(ctx)
	require.NoError(t, err, "creating provider should succeed")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.GetPermalink(ctx, tt.args, tt.commitHash, tt.file)
			if tt.wantErr {
				require.Error(t, err, "GetPermalink should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "GetPermalink should succeed")
			assert.Equal(t, tt.want, got, "permalink should match")
		})
	}
}

func TestListFiles(t *testing.T) {
	// Create provider with mock token
	os.Setenv("GITHUB_TOKEN", "mock_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	tests := []struct {
		name        string
		args        config.ProviderArgs
		wantErr     bool
		errContains string
		check       func(t *testing.T, files []string)
	}{
		{
			name: "list_files",
			args: config.ProviderArgs{
				Repo: "github.com/walteh/copyrc",
				Ref:  "main",
				Path: "pkg/provider",
			},
			check: func(t *testing.T, files []string) {
				assert.NotEmpty(t, files, "should return files")
				assert.Contains(t, files, "github.go", "should contain github.go")
			},
		},
		{
			name: "invalid_repo",
			args: config.ProviderArgs{
				Repo: "invalid",
				Ref:  "main",
				Path: "pkg/provider",
			},
			wantErr:     true,
			errContains: "invalid GitHub repository URL",
		},
	}

	ctx := zerolog.New(os.Stderr).WithContext(context.Background())
	p, err := New(ctx)
	require.NoError(t, err, "creating provider should succeed")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := p.ListFiles(ctx, tt.args)
			if tt.wantErr {
				require.Error(t, err, "ListFiles should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "ListFiles should succeed")
			if tt.check != nil {
				tt.check(t, files)
			}
		})
	}
}

func TestGetFile(t *testing.T) {
	// Create provider with mock token
	os.Setenv("GITHUB_TOKEN", "mock_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	tests := []struct {
		name        string
		args        config.ProviderArgs
		path        string
		wantErr     bool
		errContains string
		check       func(t *testing.T, content []byte)
	}{
		{
			name: "get_file",
			args: config.ProviderArgs{
				Repo: "github.com/walteh/copyrc",
				Ref:  "main",
				Path: "pkg/provider",
			},
			path: "github.go",
			check: func(t *testing.T, content []byte) {
				assert.NotEmpty(t, content, "should return content")
				assert.Contains(t, string(content), "package github", "should contain package declaration")
			},
		},
		{
			name: "invalid_repo",
			args: config.ProviderArgs{
				Repo: "invalid",
				Ref:  "main",
				Path: "pkg/provider",
			},
			path:        "github.go",
			wantErr:     true,
			errContains: "invalid GitHub repository URL",
		},
	}

	ctx := zerolog.New(os.Stderr).WithContext(context.Background())
	p, err := New(ctx)
	require.NoError(t, err, "creating provider should succeed")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc, err := p.GetFile(ctx, tt.args, tt.path)
			if tt.wantErr {
				require.Error(t, err, "GetFile should return error")
				assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				return
			}

			require.NoError(t, err, "GetFile should succeed")
			defer rc.Close()

			content, err := io.ReadAll(rc)
			require.NoError(t, err, "reading content should succeed")

			if tt.check != nil {
				tt.check(t, content)
			}
		})
	}
}
