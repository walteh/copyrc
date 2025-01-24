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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/walteh/copyrc/cmd/copyrc-refactor-1/pkg/config"
	"github.com/walteh/copyrc/cmd/copyrc-refactor-1/pkg/provider"
	"gitlab.com/tozd/go/errors"
)

func init() {
	provider.Register("github", New)
}

// 🔌 Provider implements the provider.Provider interface for GitHub
type Provider struct {
	token string
}

// 🏭 New creates a new GitHub provider
func New(ctx context.Context) (provider.Provider, error) {
	token := os.Getenv("GITHUB_TOKEN")

	return &Provider{
		token: token,
	}, nil
}

// 🔍 parseRepo parses a GitHub repository URL
func (p *Provider) parseRepo(repo string) (owner string, name string, err error) {
	// Remove https:// prefix if present
	repo = strings.TrimPrefix(repo, "https://")

	parts := strings.Split(strings.TrimPrefix(repo, "github.com/"), "/")
	if len(parts) != 2 {
		return "", "", errors.Errorf("invalid GitHub repository URL: %s", repo)
	}
	return parts[0], parts[1], nil
}

// 🌐 request makes an HTTP request to the GitHub API
func (p *Provider) request(ctx context.Context, method string, path string, query url.Values) (*http.Response, error) {
	u := &url.URL{
		Scheme:   "https",
		Host:     "api.github.com",
		Path:     path,
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, errors.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+p.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Errorf("making request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp, nil
}

// 📂 ListFiles returns a list of files in the given path
func (p *Provider) ListFiles(ctx context.Context, args config.ProviderArgs) ([]string, error) {
	owner, name, err := p.parseRepo(args.Repo)
	if err != nil {
		return nil, errors.Errorf("parsing repo: %w", err)
	}

	// Get tree
	query := url.Values{}
	query.Set("recursive", "1")
	resp, err := p.request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/git/trees/%s", owner, name, args.Ref), query)
	if err != nil {
		return nil, errors.Errorf("getting tree: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var tree struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, errors.Errorf("parsing tree: %w", err)
	}

	// Filter files
	var files []string
	prefix := filepath.Clean(args.Path) + "/"
	for _, item := range tree.Tree {
		if item.Type != "blob" {
			continue
		}
		if !strings.HasPrefix(item.Path, prefix) {
			continue
		}
		files = append(files, strings.TrimPrefix(item.Path, prefix))
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
	query := url.Values{}
	query.Set("ref", args.Ref)
	resp, err := p.request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/contents/%s", owner, name, filepath.Join(args.Path, path)), query)
	if err != nil {
		return nil, errors.Errorf("getting file: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var content struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, errors.Errorf("parsing content: %w", err)
	}

	// Decode content
	if content.Encoding != "base64" {
		return nil, errors.Errorf("unexpected content encoding: %s", content.Encoding)
	}
	data, err := base64.StdEncoding.DecodeString(content.Content)
	if err != nil {
		return nil, errors.Errorf("decoding content: %w", err)
	}

	return io.NopCloser(strings.NewReader(string(data))), nil
}

// 🎯 GetCommitHash returns the commit hash for the current ref
func (p *Provider) GetCommitHash(ctx context.Context, args config.ProviderArgs) (string, error) {
	owner, name, err := p.parseRepo(args.Repo)
	if err != nil {
		return "", errors.Errorf("parsing repo: %w", err)
	}

	// Get reference
	resp, err := p.request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", owner, name, args.Ref), nil)
	if err != nil {
		return "", errors.Errorf("getting reference: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ref); err != nil {
		return "", errors.Errorf("parsing reference: %w", err)
	}

	return ref.Object.SHA, nil
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

	return fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, name, args.Ref), nil
}
