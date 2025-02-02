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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gitlab.com/tozd/go/errors"
)

// 🏗️ Github implementation
type GithubProvider struct {
}

func NewGithubProvider() (*GithubProvider, error) {
	return &GithubProvider{}, nil
}

func parseGithubRepo(repo string) (org string, name string, err error) {
	// Remove "From " prefix if present
	repo = strings.TrimPrefix(repo, "From ")

	// Remove @ref suffix if present
	if idx := strings.LastIndex(repo, "@"); idx != -1 {
		repo = repo[:idx]
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 3 || parts[0] != "github.com" {
		return "", "", errors.Errorf("invalid github repository format: %s (expected github.com/org/repo)", repo)
	}
	return parts[1], parts[2], nil
}

type ProviderFile struct {
	Path     string `json:"path"`
	Dir      bool   `json:"dir"`
	File     bool   `json:"file"`
	Children []ProviderFile
}

func (g *GithubProvider) ListFiles(ctx context.Context, args ProviderArgs) ([]ProviderFile, error) {
	org, repo, err := parseGithubRepo(args.Repo)
	if err != nil {
		return nil, errors.Errorf("parsing github repository: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		org, repo, args.Path, args.Ref)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Errorf("creating request: %w", err)
	}

	// Add GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Errorf("fetching file list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []ProviderFile{}, nil
	}

	if resp.StatusCode == http.StatusForbidden {
		// Check if we're rate limited
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			resetTime := resp.Header.Get("X-RateLimit-Reset")
			resetTimestamp, err := strconv.ParseInt(resetTime, 10, 64)
			if err == nil {
				waitDuration := time.Until(time.Unix(resetTimestamp, 0))
				if waitDuration > 0 {
					time.Sleep(waitDuration)
					return g.ListFiles(ctx, args) // Retry after waiting
				}
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the response body once
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf("reading response body: %w", err)
	}

	// Try to decode as array first
	var files []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &files); err == nil {
		result := make([]ProviderFile, 0, len(files))
		for _, f := range files {
			fle := ProviderFile{
				Path:     f.Path,
				Dir:      f.Type == "dir",
				File:     f.Type == "file",
				Children: make([]ProviderFile, 0),
			}
			if f.Type == "dir" {
				childs, err := g.ListFiles(ctx, ProviderArgs{
					Repo: args.Repo,
					Ref:  args.Ref,
					Path: f.Path,
				})
				if err != nil {
					return nil, errors.Errorf("listing files: %w", err)
				}
				fle.Children = childs
			}
			result = append(result, fle)
		}
		return result, nil
	}

	// If array decode fails, try single file object
	var file struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &file); err != nil {
		return nil, errors.Errorf("decoding response: %w", err)
	}

	return []ProviderFile{
		{
			Path: file.Path,
			Dir:  file.Type == "dir",
		},
	}, nil
}

func (g *GithubProvider) GetCommitHash(ctx context.Context, args ProviderArgs) (string, error) {
	// Try the specified ref first
	hash, err := g.tryGetCommitHash(ctx, args)
	if err == nil {
		return hash, nil
	}

	return "", errors.Errorf("getting commit hash: %w", err)
}

func (g *GithubProvider) tryGetCommitHash(ctx context.Context, args ProviderArgs) (string, error) {

	if args.RefType == "commit" {
		return args.Ref, nil
	}
	org, repo, err := parseGithubRepo(args.Repo)
	if err != nil {
		return "", errors.Errorf("parsing github repository: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "ls-remote",
		fmt.Sprintf("https://github.com/%s/%s.git", org, repo),
		args.Ref)

	out, err := cmd.Output()
	if err != nil {
		return "", errors.Errorf("running git ls-remote: %w", err)
	}

	parts := strings.Fields(string(out))
	if len(parts) == 0 {
		return "", errors.New("no commit hash found")
	}

	return parts[0], nil
}

func (g *GithubProvider) GetPermalink(ctx context.Context, args ProviderArgs, commitHash string, file string) (string, error) {
	org, repo, err := parseGithubRepo(args.Repo)
	if err != nil {
		return "", errors.Errorf("parsing github repository: %w", err)
	}
	if file == "" && args.Path == "" {
		// archive permalink
		url, err := g.GetArchiveUrl(ctx, args)
		if err != nil {
			return "", errors.Errorf("getting archive url: %w", err)
		}
		return url, nil
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		org, repo, commitHash, file), nil
}

func (g *GithubProvider) GetSourceInfo(ctx context.Context, args ProviderArgs, commitHash string) (string, error) {
	org, repo, err := parseGithubRepo(args.Repo)
	if err != nil {
		return "", errors.Errorf("parsing github repository: %w", err)
	}
	return fmt.Sprintf("github.com/%s/%s@%s", org, repo, commitHash), nil
}

// GetArchiveUrl returns the URL to download the repository archive
func (g *GithubProvider) GetArchiveUrl(ctx context.Context, args ProviderArgs) (string, error) {
	org, repo, err := parseGithubRepo(args.Repo)
	if err != nil {
		return "", errors.Errorf("parsing github repository: %w", err)
	}
	var refPath string
	switch args.RefType {
	case "commit":
		refPath = args.Ref
	case "branch":
		refPath = "heads/" + args.Ref
	default:
		if strings.HasPrefix(args.Ref, "tags/") {
			refPath = "refs/" + args.Ref
		} else {
			refPath = "refs/tags/" + args.Ref
		}
	}

	return fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", org, repo, refPath), nil
}
