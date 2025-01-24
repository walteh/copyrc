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

package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/walteh/copyrc/cmd/copyrc-refactor-1/pkg/config"
	"gitlab.com/tozd/go/errors"
)

// 🏭 NewProvider creates a new provider based on the configuration
func NewProvider(ctx context.Context, args config.ProviderArgs) (Provider, error) {
	// Parse repository owner and name
	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return nil, errors.Errorf("invalid repository format %q, expected owner/name", args.Repo)
	}

	// Create GitHub client
	client := github.NewClient(nil)

	return &githubProvider{
		client: client,
		owner:  parts[0],
		repo:   parts[1],
		ref:    args.Ref,
		path:   args.Path,
	}, nil
}

// 📦 Provider interface for source code providers
type Provider interface {
	// ListFiles lists all files in the repository
	ListFiles(ctx context.Context, args config.ProviderArgs) ([]string, error)

	// GetFile gets the content of a file
	GetFile(ctx context.Context, args config.ProviderArgs, path string) (io.ReadCloser, error)

	// GetCommitHash gets the commit hash for the current ref
	GetCommitHash(ctx context.Context, args config.ProviderArgs) (string, error)

	// GetPermalink gets the permanent link to a file
	GetPermalink(ctx context.Context, args config.ProviderArgs, commitHash, path string) (string, error)

	// 🎯 GetSourceInfo returns a string describing the source
	GetSourceInfo(ctx context.Context, args config.ProviderArgs, commitHash string) (string, error)

	// 📦 GetArchiveURL returns the URL to download the repository archive
	GetArchiveURL(ctx context.Context, args config.ProviderArgs) (string, error)
}

// 📦 githubProvider implements Provider for GitHub repositories
type githubProvider struct {
	client *github.Client
	owner  string
	repo   string
	ref    string
	path   string
}

// 📝 ListFiles implements Provider
func (p *githubProvider) ListFiles(ctx context.Context, args config.ProviderArgs) ([]string, error) {
	// Get tree for ref
	tree, _, err := p.client.Git.GetTree(ctx, p.owner, p.repo, p.ref, true)
	if err != nil {
		return nil, errors.Errorf("getting repository tree: %w", err)
	}

	// Filter and collect files
	var files []string
	prefix := p.path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	for _, entry := range tree.Entries {
		if *entry.Type != "blob" {
			continue
		}

		if prefix != "" && !strings.HasPrefix(*entry.Path, prefix) {
			continue
		}

		relPath := strings.TrimPrefix(*entry.Path, prefix)
		if relPath != "" {
			files = append(files, relPath)
		}
	}

	return files, nil
}

// 📝 GetFile implements Provider
func (p *githubProvider) GetFile(ctx context.Context, args config.ProviderArgs, filePath string) (io.ReadCloser, error) {
	fullPath := path.Join(p.path, filePath)

	// Get file content
	content, _, _, err := p.client.Repositories.GetContents(ctx, p.owner, p.repo, fullPath, &github.RepositoryContentGetOptions{
		Ref: p.ref,
	})
	if err != nil {
		return nil, errors.Errorf("getting file content: %w", err)
	}

	// Get download URL
	if content.GetDownloadURL() == "" {
		return nil, errors.New("file download URL not available")
	}

	// Download file
	resp, err := http.Get(content.GetDownloadURL())
	if err != nil {
		return nil, errors.Errorf("downloading file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.Errorf("downloading file: unexpected status code %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// 📝 GetCommitHash implements Provider
func (p *githubProvider) GetCommitHash(ctx context.Context, args config.ProviderArgs) (string, error) {
	// Get reference
	ref, _, err := p.client.Git.GetRef(ctx, p.owner, p.repo, "refs/heads/"+p.ref)
	if err != nil {
		return "", errors.Errorf("getting reference: %w", err)
	}

	return ref.GetObject().GetSHA(), nil
}

// 📝 GetPermalink implements Provider
func (p *githubProvider) GetPermalink(ctx context.Context, args config.ProviderArgs, commitHash, filePath string) (string, error) {
	fullPath := path.Join(p.path, filePath)
	return fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s", p.owner, p.repo, commitHash, fullPath), nil
}

// 📝 GetSourceInfo implements Provider
func (p *githubProvider) GetSourceInfo(ctx context.Context, args config.ProviderArgs, commitHash string) (string, error) {
	return fmt.Sprintf("GitHub repository %s/%s at commit %s", p.owner, p.repo, commitHash), nil
}

// 📝 GetArchiveURL implements Provider
func (p *githubProvider) GetArchiveURL(ctx context.Context, args config.ProviderArgs) (string, error) {
	return fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.zip", p.owner, p.repo, p.ref), nil
}

// 🏭 Factory creates a new provider
type Factory func(ctx context.Context) (Provider, error)

var (
	// 🗺️ providers is a map of provider names to factories
	providers = make(map[string]Factory)
)

// 📝 Register registers a provider factory
func Register(name string, factory Factory) {
	providers[name] = factory
}

// 🎯 Get returns a provider by name
func Get(name string) Factory {
	return providers[name]
}

// 📥 DownloadFile downloads a file from a URL
func DownloadFile(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Errorf("making request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
