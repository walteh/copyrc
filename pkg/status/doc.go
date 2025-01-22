/*
Package status manages file storage and status tracking for copyrc.

	            +-------------+
	            |   Status    |
	            |  (Storage)  |
	            +------+------+
	                   |
	      +-----------+-----------+
	      |                       |
	+-----+-----+           +----+----+
	|   Files   |           |  Logs   |
	| (Storage) |           | (UI/UX) |
	+-----------+           +---------+

🎯 Purpose:
- Manages file storage operations
- Tracks file status (new, modified, removed)
- Provides user-friendly status reporting
- Handles file system operations safely

🔄 Flow:
1. Receives transformed content from operation
2. Manages file system operations (create, update, delete)
3. Tracks file status changes
4. Reports changes in a user-friendly format

⚡ Key Responsibilities:
- File system operations
- Status tracking
- Progress reporting
- Error handling for I/O
- Backup management (if needed)

🤝 Interfaces:
- FileManager: Handles file operations
- StatusReporter: Reports status changes
- Logger: Provides user feedback
- Formatter: Formats status messages (TODO)

📝 Design Philosophy:
The status package is responsible for all file system operations and status
tracking. It provides a clean abstraction over the file system and ensures:
- Safe file operations
- Consistent status tracking
- Beautiful progress reporting
- Clear error messages

🚧 Current Issues & TODOs:
1. File Management:
  - Create FileManager interface ✅
  - Implement safe atomic writes ✅
  - Add backup/restore capability ✅
  - Handle directory creation/cleanup ✅

2. Status Tracking:
  - Define clear file states ✅
  - Track file metadata ✅
  - Implement diff detection
  - Support for dry-run mode

3. Progress Reporting:
  - Add progress bar for large operations
  - Implement live updates ✅
  - Better error formatting
  - Support for different output formats

4. Testing:
  - Mock filesystem for testing
  - Test error conditions
  - Verify atomic operations
  - Test concurrent access

5. Missing Abstractions:
  - FormatFileOperation should be moved from operation package
  - Add FileFormatter interface for customizable output
  - Support for different status display formats (text, json, etc)
  - Better separation between storage and presentation logic

🔍 Deeper Reflection:
The current implementation has a few architectural issues:

 1. Presentation Logic Leak:
    Operation package is formatting status messages, which should be
    handled here. We need a proper FileFormatter interface:

    type FileFormatter interface {
    FormatFileOperation(path, fileType, status string, isNew, isModified, isRemoved bool) string
    FormatProgress(current, total int) string
    FormatError(err error) string
    }

 2. Status Management:
    Currently mixing storage and status tracking. Should split into:
    - StorageManager: Pure file operations
    - StatusTracker: Status and metadata
    - StatusFormatter: Presentation logic

 3. Event System:
    Should implement an event system for status changes:
    - FileCreated
    - FileModified
    - FileDeleted
    - OperationStarted
    - OperationProgress
    - OperationCompleted

 4. Async Operations:
    Need better support for:
    - Progress streaming
    - Cancellation
    - Rate limiting
    - Batch operations

Next Steps:
1. Create FileFormatter interface
2. Move formatting logic from operation package
3. Split Manager into smaller focused types
4. Add event system
5. Implement comprehensive tests

🔍 Example:

	status := status.New(cfg, logger)

	// File operations
	err := status.WriteFile(ctx, path, content)

	// Status tracking
	info := status.GetFileInfo(path)

	// Progress reporting
	status.ReportProgress(ctx, total, processed)

	// Status formatting (TODO)
	formatted := status.FormatFileOperation(path, fileType, status, isNew, isModified, isRemoved)
*/
package status
