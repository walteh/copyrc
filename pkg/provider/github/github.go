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
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/provider"
	"gitlab.com/tozd/go/errors"
	"golang.org/x/oauth2"
)

func init() {
	provider.Register("github", New)
}

// 🎯 Provider implements the provider interface for GitHub
type Provider struct {
	client *github.Client
	logger zerolog.Logger
}

// 🏭 New creates a new GitHub provider
func New(ctx context.Context) (provider.Provider, error) {
	logger := zerolog.Ctx(ctx)

	// Get token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_TOKEN environment variable not set")
	}

	// Create OAuth2 client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create GitHub client
	client := github.NewClient(tc)

	return &Provider{
		client: client,
		logger: *logger,
	}, nil
}

// 🔍 parseRepo parses a GitHub repository URL
func (p *Provider) parseRepo(repo string) (owner, name string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) < 2 {
		return "", "", errors.Errorf("invalid repository format: %s", repo)
	}

	return parts[len(parts)-2], parts[len(parts)-1], nil
}

// 📂 ListFiles returns a list of files in the given path
func (p *Provider) ListFiles(ctx context.Context, args config.ProviderArgs) ([]string, error) {
	owner, name, err := p.parseRepo(args.Repo)
	if err != nil {
		return nil, errors.Errorf("parsing repo: %w", err)
	}

	// Get repository tree
	tree, _, err := p.client.Git.GetTree(ctx, owner, name, args.Ref, true)
	if err != nil {
		return nil, errors.Errorf("getting repository tree: %w", err)
	}

	// Filter files by path
	var files []string
	for _, entry := range tree.Entries {
		if entry.GetType() != "blob" {
			continue
		}

		path := entry.GetPath()
		if !strings.HasPrefix(path, args.Path) {
			continue
		}

		files = append(files, strings.TrimPrefix(path, args.Path))
	}

	return files, nil
}

// 🔍 GetFile retrieves a single file's contents
func (p *Provider) GetFile(ctx context.Context, args config.ProviderArgs, path string) (io.ReadCloser, error) {
	owner, name, err := p.parseRepo(args.Repo)
	if err != nil {
		return nil, errors.Errorf("parsing repo: %w", err)
	}

	// Get file content
	content, _, _, err := p.client.Repositories.GetContents(ctx, owner, name, filepath.Join(args.Path, path), &github.RepositoryContentGetOptions{
		Ref: args.Ref,
	})
	if err != nil {
		return nil, errors.Errorf("getting file content: %w", err)
	}

	// Decode content
	data, err := content.GetContent()
	if err != nil {
		return nil, errors.Errorf("decoding content: %w", err)
	}

	return io.NopCloser(strings.NewReader(data)), nil
}

// 🎯 GetCommitHash returns the commit hash for the current ref
func (p *Provider) GetCommitHash(ctx context.Context, args config.ProviderArgs) (string, error) {
	owner, name, err := p.parseRepo(args.Repo)
	if err != nil {
		return "", errors.Errorf("parsing repo: %w", err)
	}

	// Get reference
	ref, _, err := p.client.Git.GetRef(ctx, owner, name, "refs/heads/"+args.Ref)
	if err != nil {
		return "", errors.Errorf("getting reference: %w", err)
	}

	return ref.Object.GetSHA(), nil
}

// 🔗 GetPermalink returns a permanent link to the file
func (p *Provider) GetPermalink(ctx context.Context, args config.ProviderArgs, commitHash string, file string) (string, error) {
	owner, name, err := p.parseRepo(args.Repo)
	if err != nil {
		return "", errors.Errorf("parsing repo: %w", err)
	}

	return fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s/%s", owner, name, commitHash, args.Path, file), nil
}

// 📝 GetSourceInfo returns a string describing the source
func (p *Provider) GetSourceInfo(ctx context.Context, args config.ProviderArgs, commitHash string) (string, error) {
	return fmt.Sprintf("%s@%s", args.Repo, commitHash[:7]), nil
}

// 📦 GetArchiveURL returns the URL to download the repository archive
func (p *Provider) GetArchiveURL(ctx context.Context, args config.ProviderArgs) (string, error) {
	owner, name, err := p.parseRepo(args.Repo)
	if err != nil {
		return "", errors.Errorf("parsing repo: %w", err)
	}

	// Get download URL
	url, _, err := p.client.Repositories.GetArchiveLink(ctx, owner, name, github.Tarball, &github.RepositoryContentGetOptions{
		Ref: args.Ref,
	}, true)
	if err != nil {
		return "", errors.Errorf("getting archive link: %w", err)
	}

	return url.String(), nil
}

// 🔍 downloadFile downloads a file from a URL
func (p *Provider) downloadFile(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Errorf("downloading file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
