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

	"github.com/walteh/copyrc/pkg/status"
	"gitlab.com/tozd/go/errors"
)

// 🧹 NewCleanOperation creates a new clean operation
func NewCleanOperation(opts Options) Operation {
	return &cleanOperation{
		BaseOperation: NewBaseOperation(opts),
	}
}

// 🧹 cleanOperation implements the clean operation
type cleanOperation struct {
	BaseOperation
}

// 🏃 Execute runs the clean operation
func (op *cleanOperation) Execute(ctx context.Context) error {
	// Get list of tracked files
	files, err := op.StatusMgr.ListFiles(ctx)
	if err != nil {
		return errors.Errorf("listing files: %w", err)
	}

	// Start tracking progress
	op.StatusMgr.StartOperation(ctx, len(files))
	defer op.StatusMgr.FinishOperation(ctx)

	// Clean each file
	for i, file := range files {
		if err := op.cleanFile(ctx, file); err != nil {
			// Track error status
			op.StatusMgr.UpdateStatus(ctx, file.Path, status.StatusUnknown, &status.FileEntry{
				Status: status.StatusUnknown,
				Metadata: map[string]string{
					"error": err.Error(),
				},
			})
			return errors.Errorf("cleaning file %s: %w", file.Path, err)
		}
		op.StatusMgr.UpdateProgress(ctx, i+1)
	}

	// Update lock file with empty state
	if err := op.StatusMgr.UpdateLockFile(ctx, "", op.Config); err != nil {
		return errors.Errorf("updating lock file: %w", err)
	}

	return nil
}

// 🗑️ cleanFile removes a file and updates its status
func (op *cleanOperation) cleanFile(ctx context.Context, file status.FileInfo) error {
	// Delete file using status manager
	if err := op.StatusMgr.DeleteFile(ctx, file.Path); err != nil {
		return errors.Errorf("deleting file: %w", err)
	}

	// Update status to deleted
	op.StatusMgr.UpdateStatus(ctx, file.Path, status.StatusDeleted, &status.FileEntry{
		Status: status.StatusDeleted,
		Metadata: map[string]string{
			"reason": "cleaned by operation",
		},
	})

	return nil
}
