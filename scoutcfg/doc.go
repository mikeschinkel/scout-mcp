// Package scoutcfg provides robust configuration file management for Scout MCP
// and related applications. It implements a secure, efficient abstraction over
// filesystem operations with a focus on JSON configuration files stored in
// user home directories following XDG Base Directory Specification conventions.
//
// # Design Philosophy
//
// The scoutcfg package follows several key design principles:
//
//   - Security First: All file paths are validated to prevent directory traversal
//   - Atomic Operations: File writes use temporary files where possible for consistency
//   - Lazy Initialization: Resources are allocated only when needed
//   - Clear Error Handling: Comprehensive error messages for common failure scenarios
//   - Testability: Filesystem abstraction allows for easy mocking and testing
//
// # Architecture Overview
//
// The package centers around the FileStore type, which provides a high-level
// interface for configuration file operations. The FileStore manages:
//
//   - Configuration directory resolution and caching
//   - Filesystem abstraction for testability
//   - JSON serialization with human-readable formatting
//   - Automatic directory creation and path validation
//   - Consistent error handling and logging
//
// # Security Model
//
// The package implements several security measures:
//
//   - Path Validation: All filenames are validated using fs.ValidPath to prevent
//     directory traversal attacks (../../../etc/passwd)
//   - Sandboxed Access: Operations are restricted to the application's configuration
//     directory under ~/.config/<appname>
//   - Controlled Permissions: New directories are created with 0755 permissions,
//     new files with 0644 permissions
//   - Input Sanitization: JSON data is validated during both serialization and
//     deserialization to prevent injection attacks
//
// # Configuration Directory Structure
//
// The package follows XDG Base Directory conventions:
//
//	$HOME/.config/
//	├── scout-mcp/          # Application-specific directory
//	│   ├── config.json     # Main configuration file
//	│   ├── tokens/         # Subdirectories are supported
//	│   │   └── user@domain.json
//	│   └── logs/           # Log files and append operations
//	│       └── activity.log
//
// # Usage Patterns
//
// ## Basic Configuration Management
//
//	store := scoutcfg.NewFileStore("my-app")
//	
//	// Save configuration
//	config := AppConfig{Setting: "value", Debug: true}
//	err := store.Save("config.json", &config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	
//	// Load configuration
//	var loadedConfig AppConfig
//	err = store.Load("config.json", &loadedConfig)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// ## Logging and Append Operations
//
//	logEntry := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), message)
//	err := store.Append("logs/activity.log", []byte(logEntry))
//	if err != nil {
//		log.Printf("Failed to log: %v", err)
//	}
//
// ## Conditional Configuration Loading
//
//	if store.Exists("config.json") {
//		err := store.Load("config.json", &config)
//		// handle error
//	} else {
//		// Use default configuration
//		config = DefaultConfig()
//		err := store.Save("config.json", &config)
//		// handle error
//	}
//
// ## Testing with Custom Directories
//
//	func TestMyConfig(t *testing.T) {
//		store := scoutcfg.NewFileStore("test-app")
//		store.SetBaseDir(t.TempDir()) // Use temporary directory for tests
//		
//		// Test configuration operations without affecting user files
//		testConfig := MyConfig{Value: "test"}
//		err := store.Save("test.json", &testConfig)
//		require.NoError(t, err)
//	}
//
// # Error Handling Patterns
//
// The package provides specific error types for common scenarios:
//
//   - fs.PathError: File system access errors (permissions, disk full)
//   - json.SyntaxError: Invalid JSON in configuration files
//   - json.UnmarshalTypeError: JSON structure mismatch with Go types
//
// Applications should handle these errors appropriately:
//
//	err := store.Load("config.json", &config)
//	if err != nil {
//		if os.IsNotExist(err) {
//			// First run - create default config
//			config = DefaultConfig()
//			store.Save("config.json", &config)
//		} else if jsonErr, ok := err.(*json.SyntaxError); ok {
//			// Corrupted config file
//			log.Printf("Config file corrupted at byte %d: %v", jsonErr.Offset, err)
//		} else {
//			// Other errors (permissions, disk full, etc.)
//			log.Fatal(err)
//		}
//	}
//
// # Performance Considerations
//
// The package is optimized for typical configuration file usage patterns:
//
//   - Lazy Initialization: Configuration directory and filesystem interface
//     are only resolved when first needed
//   - Path Caching: Configuration directory path is cached after first resolution
//   - Efficient JSON: Uses json.Marshal/Unmarshal for optimal performance
//   - Minimal Allocations: Reuses buffers where possible
//
// For high-frequency operations, consider:
//   - Batching multiple configuration changes
//   - Caching loaded configuration in memory
//   - Using separate FileStore instances for different configuration categories
//
// # Thread Safety
//
// FileStore instances are safe for concurrent use across multiple goroutines.
// The underlying filesystem operations are atomic where supported by the OS,
// and the package does not maintain mutable state that could cause race conditions.
//
// However, applications should implement their own synchronization for complex
// scenarios like:
//   - Read-modify-write operations across multiple files
//   - Coordinated updates to related configuration files
//   - Cache invalidation in multi-instance applications
//
// # Integration with Scout MCP
//
// This package is specifically designed to support Scout MCP's configuration
// requirements including:
//
//   - Session token storage and retrieval
//   - User approval workflow state persistence  
//   - Tool configuration and permission management
//   - Activity logging and audit trails
//   - Multi-user configuration isolation
//
// The package integrates with Scout MCP's security model by providing
// controlled filesystem access that respects the MCP server's permission
// boundaries and logging requirements.
package scoutcfg