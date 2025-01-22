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

package base

import (
	"context"
	"io"

	"github.com/walteh/copyrc/pkg/config"
	"gitlab.com/tozd/go/errors"
)

// 🔌 Provider is the base interface for repository providers
type Provider interface {
	// 📂 ListFiles returns a list of files in the given path
	ListFiles(ctx context.Context, args config.ProviderArgs) ([]string, error)

	// 🔍 GetFile retrieves a single file's contents
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

// 🏭 Factory is a function that creates a new provider
type Factory func(ctx context.Context) (Provider, error)

var (
	// 🗺️ providers is a map of provider names to their factories
	providers = make(map[string]Factory)
)

// 📝 Register registers a provider factory
func Register(name string, factory Factory) {
	providers[name] = factory
}

// 🎯 Get returns a provider by name
func Get(ctx context.Context, name string) (Provider, error) {
	factory, ok := providers[name]
	if !ok {
		return nil, errors.Errorf("unknown provider: %s", name)
	}

	return factory(ctx)
}

// 🔍 GetFromRepo determines the provider from a repo URL
func GetFromRepo(ctx context.Context, repo string) (Provider, error) {
	// TODO: Implement provider detection based on repo URL
	// For now, just return GitHub provider
	factory, ok := providers["github"]
	if !ok {
		return nil, errors.Errorf("github provider not registered")
	}

	return factory(ctx)
}
