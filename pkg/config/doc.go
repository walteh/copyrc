/*
Package config manages configuration parsing and validation for copyrc.

	            +-------------+
	            |   Config    |
	            | (Settings)  |
	            +------+------+
	                   |
	      +-----------+-----------+
	      |                       |
	+-----+-----+           +----+----+
	|   YAML    |           |   HCL   |
	| Parser    |           | Parser  |
	+-----------+           +---------+

🎯 Purpose:
- Manages configuration loading and parsing
- Validates configuration values
- Provides type-safe config access
- Supports multiple config formats

🔄 Flow:
1. Reads configuration from file
2. Parses format-specific syntax
3. Validates configuration values
4. Provides validated config to other packages

⚡ Key Responsibilities:
- Configuration parsing
- Schema validation
- Default value management
- Type safety
- Format abstraction

🤝 Interfaces:
- Parser: Format-specific parsing
- Validator: Configuration validation
- Config: Type-safe config access

📝 Design Philosophy:
The config package is the source of truth for all configuration. It:
- Provides a clean interface for config access
- Ensures type safety and validation
- Abstracts away format-specific details
- Makes configuration errors clear and actionable

🚧 Current Issues & TODOs:
1. Validation:
  - Enhanced validation rules
  - Custom validators
  - Path normalization
  - Environment variable support

2. Schema Management:
  - Version control for schemas
  - Migration support
  - Backward compatibility
  - Schema documentation

3. Error Handling:
  - Better validation errors
  - Configuration suggestions
  - Default value hints
  - Format detection

4. Testing:
  - More edge cases
  - Format compatibility
  - Migration testing
  - Error message clarity

💡 Ideal Validation:

	type Validator interface {
		Validate(cfg *Config) error
		SetDefaults(cfg *Config)
		Normalize(cfg *Config) error
	}

	// Example custom validator
	type PathValidator struct {
		RequiredPaths []string
		AllowedExts  []string
	}

	func (v *PathValidator) Validate(cfg *Config) error {
		// Path existence
		// Extension validation
		// Permission checks
		return nil
	}

🔍 Example:

	// Load with validation
	cfg, err := config.Load(ctx, "config.yaml")
	if err != nil {
		var verr *ValidationError
		if errors.As(err, &verr) {
			// Show helpful message
			fmt.Printf("Config error: %s\n", verr.Suggestion)
		}
		return err
	}

	// Access with type safety
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
*/
package config
