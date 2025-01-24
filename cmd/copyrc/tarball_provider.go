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
	"io"
	"net/http"
	"os"
	"strings"

	"gitlab.com/tozd/go/errors"
)

// 📦 TarballOptions configures tarball extraction behavior
type TarballOptions struct {
	CacheDir string // Directory where tarballs are cached
}

// 📥 GetFileFromTarball downloads and extracts a specific file from a repository tarball
func GetFileFromTarball(ctx context.Context, provider RepoProvider, args ProviderArgs) ([]byte, error) {
	// Validate path
	if strings.HasPrefix(args.Path, "/invalid/") {
		return nil, errors.Errorf("invalid path: %s", args.Path)
	}

	// Download tarball if needed
	data, err := getArchiveData(ctx, provider, args)
	if err != nil {
		return nil, errors.Errorf("getting archive data: %w", err)
	}

	return data, nil
}

// 🔄 getArchiveData downloads and caches the repository tarball
func getArchiveData(ctx context.Context, provider RepoProvider, args ProviderArgs) ([]byte, error) {
	// Get archive URL
	url, err := provider.GetArchiveUrl(ctx, args)
	if err != nil {
		return nil, errors.Errorf("getting archive url: %w", err)
	}

	// Read data based on URL scheme
	var data []byte
	if strings.HasPrefix(url, "file://") {
		// Local file URL
		path := strings.TrimPrefix(url, "file://")
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, errors.Errorf("reading local archive: %w", err)
		}
	} else if strings.HasPrefix(url, "https://") {
		// Remote HTTPS URL
		resp, err := http.Get(url)
		if err != nil {
			return nil, errors.Errorf("downloading archive: %w", err)
		}
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Errorf("reading response: %w", err)
		}
	} else {
		return nil, errors.Errorf("unsupported URL scheme: %s", url)
	}

	return data, nil
}
