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

/*
Package status manages the state and tracking of file operations in copyrc.

🎯 Purpose:
The status package is responsible for tracking, persisting, and validating the state
of file operations. It provides a centralized way to manage the status of copied
files and ensure consistency between operations.

🔄 Status Flow:

	                        ┌─────────────┐
	                        │  Operation  │
	                        │   Package   │
	                        └─────────────┘
	                              │
	        ┌───────────────┬─────┴─────┬───────────────┐
	        ▼               ▼           ▼               ▼
	┌─────────────┐  ┌─────────────┐   ┌─────────────┐
	│    Load     │  │   Update    │   │    Save     │
	│   Status    │  │   Status    │   │   Status    │
	└─────────────┘  └─────────────┘   └─────────────┘
	        │               │                   │
	        └───────────────┴───────────────────┘
	                       │
	                ┌─────────────┐
	                │  .copyrc    │
	                │    .lock    │
	                └─────────────┘

📦 Package Structure:

 1. Core Types:
    ```go
    // Status represents the current state of copied files
    type Status struct {
    CommitHash     string                // Current commit hash
    LastUpdated    time.Time             // Last update time
    Config         *config.Config        // Current configuration
    CopiedFiles    map[string]FileInfo   // Tracked files
    GeneratedFiles map[string]FileInfo   // Generated files
    }

    // FileInfo represents a tracked file's status
    type FileInfo struct {
    Path       string    // File path
    Hash       string    // Content hash
    UpdatedAt  time.Time // Last update time
    Source     string    // Source location
    Permalink  string    // Source permalink
    }
    ```

 2. Core Interface:
    ```go
    // Manager handles status persistence and validation
    type Manager interface {
    Load(ctx context.Context) (*Status, error)
    Save(ctx context.Context, status *Status) error
    Update(ctx context.Context, file FileInfo) error
    Validate(ctx context.Context) error
    }
    ```

🔑 Key Responsibilities:

1. Status Management:
  - Load and save status files
  - Track file modifications
  - Validate file consistency
  - Handle status locking

2. File Tracking:
  - Track copied files
  - Track generated files
  - Maintain file metadata
  - Handle file hashing

3. Status Validation:
  - Check file existence
  - Verify file hashes
  - Compare timestamps
  - Detect modifications

4. Lock File Management:
  - Create lock files
  - Handle concurrent access
  - Clean up stale locks
  - Maintain lock integrity

🤝 Integration with Operations:

1. Operation Package Relationship:

  - Operations use status for state

  - Status validates operations

  - Status tracks operation results

  - Status enforces consistency

    2. Status Flow in Operations:
    ```
    Operation Start
    │
    ▼
    Load Status ────────┐
    │             │
    ▼             │
    Check Status       │
    │            │
    ▼            │
    Execute Operation  │
    │            │
    ▼            │
    Update Status ◄────┘
    │
    ▼
    Operation End
    ```

🔒 Lock File Format:
```json

	{
	    "commit_hash": "abc123",
	    "last_updated": "2024-01-22T15:04:05Z",
	    "config": {
	        // Configuration snapshot
	    },
	    "copied_files": {
	        "file.go": {
	            "path": "pkg/file.go",
	            "hash": "sha256:...",
	            "updated_at": "2024-01-22T15:04:05Z",
	            "source": "github.com/org/repo",
	            "permalink": "https://..."
	        }
	    }
	}

```

🎯 Implementation Guidelines:

1. Status Operations:
  - Always atomic file operations
  - Use temporary files for updates
  - Handle partial writes
  - Maintain backup copies

2. Error Handling:
  - Recover from corrupted files
  - Handle missing files gracefully
  - Provide detailed error context
  - Support status rollback

3. Concurrency:
  - Use file-based locking
  - Handle concurrent updates
  - Prevent race conditions
  - Support distributed ops

4. Performance:
  - Cache status in memory
  - Batch status updates
  - Optimize file I/O
  - Use efficient formats

🔍 Testing Strategy:

1. Unit Tests:
  - Test file operations
  - Test status validation
  - Test error handling
  - Test concurrency

2. Integration Tests:
  - Test with operations
  - Test file consistency
  - Test recovery scenarios
  - Test performance

3. Stress Tests:
  - Test concurrent access
  - Test large file counts
  - Test frequent updates
  - Test error recovery

🔜 Future Enhancements:
1. [ ] Distributed locking
2. [ ] Status compression
3. [ ] Status versioning
4. [ ] Remote status sync
5. [ ] Status backup/restore
*/
package status
