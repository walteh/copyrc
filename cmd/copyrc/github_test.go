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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGithubRepo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOrg  string
		wantRepo string
		wantErr  bool
	}{
		{
			name:     "simple repo",
			input:    "github.com/org/repo",
			wantOrg:  "org",
			wantRepo: "repo",
			wantErr:  false,
		},
		{
			name:     "repo with ref",
			input:    "github.com/golang/tools@master",
			wantOrg:  "golang",
			wantRepo: "tools",
			wantErr:  false,
		},
		{
			name:     "repo with From prefix",
			input:    "From github.com/golang/tools@master",
			wantOrg:  "golang",
			wantRepo: "tools",
			wantErr:  false,
		},
		{
			name:    "invalid format",
			input:   "not/enough/parts",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, repo, err := parseGithubRepo(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOrg, org)
			assert.Equal(t, tt.wantRepo, repo)
		})
	}
}
