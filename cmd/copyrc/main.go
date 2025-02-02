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
	"flag"
	"fmt"
	"os"
	"strings"

	"gitlab.com/tozd/go/errors"
)

// 🔄 Replacement represents a string replacement
type Replacement struct {
	Old  string  `json:"old" hcl:"old" yaml:"old" cty:"old"`
	New  string  `json:"new" hcl:"new" yaml:"new" cty:"new"`
	File *string `json:"file,omitempty" hcl:"file,optional" yaml:"file,omitempty"`
}

// 📦 Input represents raw command line input
type Input struct {
	SrcRefType   string     // commit, branch, empty (normal ref)
	SrcRepo      string     // Full repo URL (e.g. github.com/org/repo)
	SrcRef       string     // Branch or tag
	SrcPath      string     // Path within repo
	DestPath     string     // Local destination path
	Replacements arrayFlags // String replacements
	IgnoreFiles  arrayFlags // Files to ignore
	Clean        bool       // Whether to clean destination directory
	Status       bool       // Whether to check local status
	RemoteStatus bool       // Whether to check remote status
	Force        bool       // Whether to force update even if status is ok
	UseTarball   bool       // Whether to use tarball-based file access
	Async        bool       // Whether to process files asynchronously
}

// 🏭 Create config from input (backward compatibility)
func NewConfigFromInput(input Input, provider RepoProvider) (*Config, error) {
	replacements := make([]Replacement, 0, len(input.Replacements))
	for _, r := range input.Replacements {
		parts := strings.SplitN(r, ":", 2)
		if len(parts) == 2 {
			replacements = append(replacements, Replacement{Old: parts[0], New: parts[1]})
		}
	}

	return &Config{
		Source: Source{
			Repo:    input.SrcRepo,
			Ref:     input.SrcRef,
			Path:    input.SrcPath,
			RefType: input.SrcRef,
		},
		DestPath: input.DestPath,
		CopyArgs: &CopyEntry_Options{
			Replacements: replacements,
			IgnoreFiles:  []string(input.IgnoreFiles),
		},
		Clean:        input.Clean,
		Status:       input.Status,
		RemoteStatus: input.RemoteStatus,
		Force:        input.Force,
		Async:        input.Async,
	}, nil
}

func main() {
	ctx := context.Background()
	logger := NewDiscardDebugLogger(os.Stdout)
	ctx = NewLoggerInContext(ctx, logger)

	// 🎯 Parse command line flags
	var input Input
	var configFile string
	var showVersion bool

	flag.StringVar(&configFile, "config", ".copyrc.hcl", "path to config file")
	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.StringVar(&input.SrcRepo, "src-repo", "", "Source repository (e.g. github.com/org/repo)")
	flag.StringVar(&input.SrcRef, "ref", "main", "Source branch/ref")
	flag.StringVar(&input.SrcPath, "src-path", "", "Source path within repository")
	flag.StringVar(&input.DestPath, "dest-path", "", "Destination path")
	flag.StringVar(&input.SrcRefType, "src-ref-type", "", "source ref type (commit, branch, empty)")
	flag.Var(&input.Replacements, "replacements", "JSON array or comma-separated list of replacements in old:new format")
	flag.Var(&input.IgnoreFiles, "ignore", "JSON array or comma-separated list of files to ignore")
	flag.BoolVar(&input.Clean, "clean", false, "Clean destination directory before copying")
	flag.BoolVar(&input.Status, "status", false, "Check if files are up to date (local check only)")
	flag.BoolVar(&input.RemoteStatus, "remote-status", false, "Check if files are up to date (includes remote check)")
	flag.BoolVar(&input.Force, "force", false, "Force update even if status is ok")
	flag.BoolVar(&input.Async, "async", false, "Process files asynchronously")
	flag.BoolVar(&input.UseTarball, "use-tarball", false, "Whether to use tarball-based file access")
	flag.Parse()

	if showVersion {
		fmt.Print(FormatVersion())
		os.Exit(0)
	}

	gh, err := NewGithubProvider()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// 🔍 Check if using config file
	if configFile != "" {
		cfg, err := LoadConfig(configFile)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		if err := cfg.RunAll(ctx, input.Clean, input.Status, input.RemoteStatus, input.Force, gh); err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		return
	}

	// 🔍 Validate required flags
	var missingFlags []string
	if input.SrcRepo == "" {
		missingFlags = append(missingFlags, "src-repo")
	}
	if input.SrcPath == "" {
		missingFlags = append(missingFlags, "src-path")
	}
	if input.DestPath == "" {
		missingFlags = append(missingFlags, "dest-path")
	}

	if len(missingFlags) > 0 {
		logger.Error(fmt.Sprintf("Required flags missing: %s", strings.Join(missingFlags, ", ")))
		flag.Usage()
		os.Exit(1)
	}

	// 🚀 Run the copy operation
	cfg, err := NewConfigFromInput(input, gh)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	if err := process(ctx, cfg, gh); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// processDirectory is defined in process.go

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	// Try to parse as JSON array first
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		var arr []string
		d := json.NewDecoder(strings.NewReader(value))
		d.UseNumber() // To prevent numbers from being converted to float64
		if err := d.Decode(&arr); err != nil {
			return errors.Errorf("unmarshalling json: %w", err)
		}
		*i = arr
		return nil
	}

	// If not JSON, treat as comma-separated list
	if strings.Contains(value, ",") {
		*i = strings.Split(value, ",")
		return nil
	}

	// Single value
	*i = append(*i, value)
	return nil
}
