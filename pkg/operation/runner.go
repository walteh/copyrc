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
	"sync"

	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"
)

// 🏃 OperationRunner executes operations
type OperationRunner struct {
	logger *zerolog.Logger
	async  bool
}

// 🏗️ NewRunner creates a new runner
func NewRunner(logger *zerolog.Logger, async bool) *OperationRunner {
	return &OperationRunner{
		logger: logger,
		async:  async,
	}
}

// 🏃 Run executes an operation
func (r *OperationRunner) Run(ctx context.Context, op Operation) error {
	if r.async {
		return r.runAsync(ctx, op)
	}
	return r.runSync(ctx, op)
}

// 🔄 runSync runs an operation synchronously
func (r *OperationRunner) runSync(ctx context.Context, op Operation) error {
	return op.Execute(ctx)
}

// ⚡ runAsync runs an operation asynchronously
func (r *OperationRunner) runAsync(ctx context.Context, op Operation) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := op.Execute(ctx); err != nil {
			errCh <- errors.Errorf("executing operation: %w", err)
		}
	}()

	// Wait for completion or context cancellation
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return errors.Errorf("operation cancelled: %w", ctx.Err())
	case err := <-errCh:
		return err
	case <-done:
		return nil
	}
}
