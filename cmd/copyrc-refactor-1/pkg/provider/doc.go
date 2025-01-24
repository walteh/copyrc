/*
Package provider defines the interface for source code providers in copyrc.

	            +-------------+
	            |  Provider   |
	            |  (Source)   |
	            +------+------+
	                   |
	      +-----------+-----------+
	      |                       |
	+-----+-----+           +----+----+
	|  GitHub   |           |  Local  |
	| Provider  |           | Files   |
	+-----------+           +---------+

🎯 Purpose:
- Abstracts source code retrieval
- Provides unified interface for different sources
- Handles remote/local file access
- Manages source metadata

🔄 Flow:
1. Receives configuration from user
2. Connects to source (GitHub, local, etc)
3. Lists available files
4. Provides file content on demand
5. Tracks source metadata (commit hash, etc)

⚡ Key Responsibilities:
- Source connection management
- File listing and retrieval
- Error handling for source access
- Metadata management
- Caching (if implemented)

🤝 Interfaces:
- Provider: Core interface for source access
- SourceInfo: Metadata about the source
- FileReader: File content access

📝 Design Philosophy:
The provider package is the source of truth for all file content. It:
- Abstracts away source-specific details
- Provides a clean interface for file access
- Handles all source-specific error cases
- Manages source metadata consistently

🚧 Current Issues & TODOs:
1. Error Handling:
  - Better context in errors
  - Retry mechanisms for network issues
  - Rate limiting support
  - Connection pooling

2. Caching:
  - Implement caching layer
  - Cache invalidation strategy
  - Memory vs disk caching
  - Cache size limits

3. Metadata:
  - Enhanced source tracking
  - Better commit info
  - File history support
  - Source statistics

4. Testing:
  - Mock HTTP responses
  - Test rate limiting
  - Test connection issues
  - Verify caching behavior

💡 Ideal Error Handling:

	// Rich error types
	type ProviderError struct {
		Op      string // Operation being performed
		Source  string // Source information
		Path    string // File path if applicable
		Err     error  // Underlying error
		Retried int    // Number of retries attempted
	}

	// Retry mechanism
	func (p *Provider) withRetry(ctx context.Context, op func() error) error {
		var lastErr error
		for attempt := 0; attempt < p.maxRetries; attempt++ {
			if err := op(); err == nil {
				return nil
			} else {
				lastErr = err
				time.Sleep(p.backoff(attempt))
			}
		}
		return lastErr
	}

🔍 Example:

	provider, err := github.New(ctx)

	// With retries and rate limiting
	files, err := provider.ListFiles(ctx, args)
	if err != nil {
		// Rich error information
		var perr *ProviderError
		if errors.As(err, &perr) {
			log.Printf("Failed to list files: %s", perr)
		}
	}

	// With caching
	content, err := provider.GetFile(ctx, args, path)
*/
package provider
