/*
Package operation provides the core operation system for copyrc.

🎯 Purpose:
The operation package implements the core functionality for executing file operations
like copying, cleaning, and status checking. It provides a unified interface for
all operations and handles their lifecycle.

🔄 Operation Flow:

	┌─────────────┐
	│   Config    │
	└─────────────┘
	      │
	      ▼

┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Operation  │    │  Operation  │    │  Operation  │    │   Status    │
│  Registry   │◄───│   Factory  │◄───│   Runner   │───►│   Manager   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘

	                              │
	        ┌───────────────┬─────┴─────┬───────────────┐
	        ▼               ▼           ▼               ▼
	┌─────────────┐  ┌─────────────┐   ┌─────────────┐  ┌─────────────┐
	│    Copy     │  │   Clean     │   │   Status    │  │   Remote    │
	│  Operation  │  │  Operation  │   │  Operation  │  │   Status    │
	└─────────────┘  └─────────────┘   └─────────────┘  └─────────────┘

📦 Package Structure:

 1. Core Interfaces (operation.go):
    ```go
    type Operation interface {
    Execute(ctx context.Context) error
    Status(ctx context.Context) error
    Clean(ctx context.Context) error
    }

    type Factory interface {
    CreateOperation(cfg *config.Config) (Operation, error)
    }

    type Runner interface {
    Run(ctx context.Context, op Operation) error
    }
    ```

2. Registry System (registry.go):
  - RegisterOperation(name string, factory Factory)
  - GetOperation(name string) Factory
  - ListOperations() []string

3. Operation Implementations:

	a. Copy Operation (copy.go):
	   - copyOperation struct
	   - Execute: Process and copy files
	   - Status: Check file status
	   - Clean: Remove copied files
	   - Helpers:
	     * processFile(ctx, file string) error
	     * applyReplacements(content []byte) []byte
	     * addFileHeader(file string, content []byte) []byte

	b. Clean Operation (clean.go):
	   - cleanOperation struct
	   - Execute: Clean destination directory
	   - Status: Check what would be cleaned
	   - Helpers:
	     * removeFile(ctx, path string) error
	     * cleanDirectory(ctx, path string) error

	c. Status Operation (status.go):
	   - statusOperation struct
	   - Execute: Check and report status
	   - Helpers:
	     * checkLocalStatus() error
	     * generateReport() string

	d. Remote Status Operation (remote.go):
	   - remoteStatusOperation struct
	   - Execute: Check remote source status
	   - Helpers:
	     * checkRemoteStatus() error
	     * compareWithRemote() error

4. Support Services:

	a. Status Manager (status/manager.go):
	   - NewStatusManager(path string) *StatusManager
	   - Load() error
	   - Save() error
	   - UpdateFile(file FileInfo) error
	   - CheckStatus() error

	b. File Processor (processor/processor.go):
	   - ProcessFile(ctx context.Context, opts ProcessOptions) error
	   - GenerateHeader(file string) string
	   - HandleReplacements(content []byte) []byte

5. Special Handlers:

	a. Go Embed Handler (embed/handler.go):
	   - GenerateEmbedFile(opts EmbedOptions) error
	   - CreateEmbedDirective(file string) string

	b. Pattern Matcher (pattern/matcher.go):
	   - MatchPattern(pattern, path string) bool
	   - IsIgnored(path string, patterns []string) bool

🎯 Implementation Plan:

1. Core Framework (Phase 1):
  - [ ] operation.go: Core interfaces
  - [ ] registry.go: Operation registry
  - [ ] runner.go: Operation runner
  - [ ] factory.go: Operation factory

2. Basic Operations (Phase 2):
  - [ ] copy.go: Copy operation
  - [ ] clean.go: Clean operation
  - [ ] status.go: Status operation

3. Support Services (Phase 3):
  - [ ] status/manager.go: Status management
  - [ ] processor/processor.go: File processing
  - [ ] pattern/matcher.go: Pattern matching

4. Advanced Features (Phase 4):
  - [ ] remote.go: Remote status
  - [ ] embed/handler.go: Go embed support
  - [ ] async/runner.go: Async operation support

5. Integration (Phase 5):
  - [ ] Integration with config package
  - [ ] Integration with provider package
  - [ ] CLI command integration

🔍 Testing Strategy:

1. Unit Tests:
  - Test each operation in isolation
  - Mock file system operations
  - Mock remote provider calls

2. Integration Tests:
  - Test operation chaining
  - Test with real file system
  - Test with mock remote provider

3. Performance Tests:
  - Benchmark large file operations
  - Test concurrent operations
  - Memory usage analysis

4. Error Handling Tests:
  - Test various error conditions
  - Test cleanup on failure
  - Test partial success scenarios

🔜 Future Enhancements:
1. [ ] Operation dependency system
2. [ ] Operation rollback support
3. [ ] Progress reporting
4. [ ] Operation hooks
5. [ ] Custom operation plugins
*/
package operation
