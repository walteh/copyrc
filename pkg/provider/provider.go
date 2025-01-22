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
	"io"
	"net/http"

	"github.com/walteh/copyrc/pkg/config"
	"gitlab.com/tozd/go/errors"
)

// 🔌 Provider is the interface for repository providers
type Provider interface {
	// 📂 ListFiles returns a list of files in the given path
	ListFiles(ctx context.Context, args config.ProviderArgs) ([]string, error)

	// 📄 GetFile retrieves a single file's contents
	GetFile(ctx context.Context, args config.ProviderArgs, path string) (io.ReadCloser, error)

	// 🎯 GetCommitHash returns the commit hash for the current ref
	GetCommitHash(ctx context.Context, args config.ProviderArgs) (string, error)

	// 🔗 GetPermalink returns a permanent link to the file
	GetPermalink(ctx context.Context, args config.ProviderArgs, commitHash string, file string) (string, error)

	// 📝 GetSourceInfo returns a string describing the source
	GetSourceInfo(ctx context.Context, args config.ProviderArgs, commitHash string) (string, error)

	// 📦 GetArchiveURL returns the URL to download the repository archive
	GetArchiveURL(ctx context.Context, args config.ProviderArgs) (string, error)
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
