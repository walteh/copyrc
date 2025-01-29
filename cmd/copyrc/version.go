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

// VersionInfo represents the version information of the binary
type VersionInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
	VCS       string `json:"vcs"`
	Revision  string `json:"revision"`
	Time      string `json:"time"`
	Modified  bool   `json:"modified"`
}

// GetVersionInfo returns the version information from build info
func GetVersionInfo() *VersionInfo {
	info := &VersionInfo{
		Version:   "dev",
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		info.Version = buildInfo.Main.Version
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs":
				info.VCS = setting.Value
			case "vcs.revision":
				info.Revision = setting.Value
			case "vcs.time":
				info.Time = setting.Value
			case "vcs.modified":
				info.Modified = setting.Value == "true"
			}
		}
	}

	return info
}

// FormatVersion returns a formatted string of version information
func FormatVersion() string {
	info := GetVersionInfo()
	modified := ""
	if info.Modified {
		modified = " (modified)"
	}
	return fmt.Sprintf(`🚀 copyrc version info:
Version:   %s
Revision:  %s%s
Built:     %s
Go:        %s
Platform:  %s
`, info.Version, info.Revision, modified, info.Time, info.GoVersion, info.Platform)
}
