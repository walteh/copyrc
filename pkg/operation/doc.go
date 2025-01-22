/*
Package operation implements the core business logic for copying and transforming files.

	+-------------+
	|  Operation  |
	| (Core Logic)|
	+------+------+
	       |
	+------+------+
	|   Process   |
	| (Transform) |
	+------+------+

🎯 Purpose:
- Orchestrates the copying and transformation of files
- Manages file content replacements and transformations
- Coordinates between provider (source) and status (destination)

🔄 Flow:
1. Receives file paths from provider
2. Applies transformations (replacements, etc)
3. Delegates file storage to status package
4. Reports progress via logging

⚡ Key Responsibilities:
- File content transformation
- Pattern matching for ignores
- Coordinating async operations
- Error handling and recovery

🤝 Interfaces:
- Provider: Source of truth for files
- Status: Handles file storage and status tracking
- Config: Provides operation parameters

📝 Design Philosophy:
The operation package is the heart of copyrc, but it should remain focused on
transformation logic. It delegates file I/O to the status package and source
retrieval to providers. This separation allows for:
- Clear responsibility boundaries
- Easier testing
- Flexible storage implementations
- Independent provider implementations

🚧 Current Issues & TODOs:
1. File I/O Responsibility:
  - Move all file I/O operations to status package
  - Remove direct os.* calls
  - Use status.FileManager interface instead

2. Status Management:
  - Remove status tracking logic
  - Delegate to status package
  - Pass file metadata through instead of calculating

3. Error Handling:
  - Improve error context
  - Add recovery mechanisms for async operations
  - Better error aggregation for batch operations

4. Testing:
  - Add more edge cases
  - Better async testing
  - Mock status package properly

🔍 Example:

	mgr := operation.New(cfg, provider, status, logger)
	err := mgr.ProcessFiles(ctx, files)

💡 Ideal Flow:
1. Get files from provider
2. Transform content (replacements)
3. Pass to status for storage
4. Handle any errors

The operation package should be like a pure function:
Input (provider) -> Transform -> Output (status)
*/
package operation
