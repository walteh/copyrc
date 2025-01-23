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

package operation

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/provider"
	"github.com/walteh/copyrc/pkg/status"
)

// 🔧 Operation represents a file operation
type Operation interface {
	Execute(ctx context.Context) error
}

// 🛠️ Options holds common options for operations
type Options struct {
	Config    *config.Config
	Provider  provider.Provider
	StatusMgr *status.Manager
	Logger    *zerolog.Logger
}

// 📦 BaseOperation provides common functionality for operations
type BaseOperation struct {
	Config    *config.Config
	Provider  provider.Provider
	StatusMgr *status.Manager
	Logger    *zerolog.Logger
}

// 🏗️ NewBaseOperation creates a new base operation
func NewBaseOperation(opts Options) BaseOperation {
	return BaseOperation{
		Config:    opts.Config,
		Provider:  opts.Provider,
		StatusMgr: opts.StatusMgr,
		Logger:    opts.Logger,
	}
}

// 🏃 Runner executes operations
type Runner interface {
	// Run executes an operation with the given context
	Run(ctx context.Context, op Operation) error
}
