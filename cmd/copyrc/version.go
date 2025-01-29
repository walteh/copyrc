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
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	// Version is the version of the binary, set during build
	Version = "dev"
	// CommitSHA is the git commit SHA, set during build
	CommitSHA = "unknown"
	// BuildTime is the time the binary was built, set during build
	BuildTime = "unknown"
)

// VersionInfo represents the version information of the binary
type VersionInfo struct {
	Version      string `json:"version"`
	CommitSHA    string `json:"commit_sha"`
	BuildTime    string `json:"build_time"`
	GoVersion    string `json:"go_version"`
	Platform     string `json:"platform"`
	BuildInfo    string `json:"build_info"`
	Dependencies []struct {
		Path    string `json:"path"`
		Version string `json:"version"`
	} `json:"dependencies,omitempty"`
}

// GetVersionInfo returns the version information of the binary
func GetVersionInfo() *VersionInfo {
	info := &VersionInfo{
		Version:   Version,
		CommitSHA: CommitSHA,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	// Get build info from debug package
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		info.BuildInfo = buildInfo.Main.Path
		for _, dep := range buildInfo.Deps {
			info.Dependencies = append(info.Dependencies, struct {
				Path    string `json:"path"`
				Version string `json:"version"`
			}{
				Path:    dep.Path,
				Version: dep.Version,
			})
		}
	}

	return info
}

// FormatVersion returns a formatted string of version information
func FormatVersion() string {
	info := GetVersionInfo()
	return fmt.Sprintf(`🚀 copyrc version info:
Version:   %s
Commit:    %s
Built:     %s
Go:        %s
Platform:  %s
`, info.Version, info.CommitSHA, info.BuildTime, info.GoVersion, info.Platform)
}
