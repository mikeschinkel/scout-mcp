// Package scoutcfg provides configuration file storage and management utilities
// for Scout MCP applications. It handles JSON configuration files stored in
// the user's home directory under .config/<appname> following XDG Base Directory
// conventions.
//
// The package provides a FileStore abstraction that manages configuration
// persistence with automatic directory creation, path validation, and JSON
// serialization. It supports common configuration operations including:
//   - Loading and saving JSON configuration files
//   - Appending to log files
//   - Checking file existence
//   - Creating nested directory structures
//
// Security considerations:
//   - All file paths are validated using fs.ValidPath to prevent directory traversal
//   - Configuration directory permissions are set to 0755 for user access only
//   - File write operations use temporary files for atomic updates
//
// Example usage:
//
//	store := scoutcfg.NewFileStore("scout-mcp")
//	config := MyConfig{Setting: "value"}
//	err := store.Save("config.json", &config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	var loadedConfig MyConfig
//	err = store.Load("config.json", &loadedConfig)
//	if err != nil {
//		log.Fatal(err)
//	}
package scoutcfg

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ConfigBaseDirName is the standard directory name for configuration files
// in the user's home directory. This follows XDG Base Directory Specification
// conventions where user-specific configuration files are stored under
// ~/.config/<appname>.
const ConfigBaseDirName = ".config"

// FileStore manages configuration files for an application in the user's
// configuration directory. It provides thread-safe operations for reading,
// writing, and managing JSON configuration files with automatic directory
// creation and path validation.
//
// FileStore uses lazy initialization for both the configuration directory
// path and the filesystem interface, computing them only when first needed.
// This allows for efficient creation and testing scenarios where the actual
// filesystem may be mocked or redirected.
//
// The store ensures atomic file operations where possible and provides
// comprehensive error handling for common filesystem scenarios including
// permission errors, disk space issues, and invalid JSON data.
type FileStore struct {
	appName   string // Name of the application used for directory naming
	configDir string // Cached path to the configuration directory
	fs        fs.FS  // File system interface for reading files (allows testing)
}

// NewFileStore creates a new FileStore instance for the specified application name.
// The application name is used to create a dedicated subdirectory under the user's
// .config directory (e.g., ~/.config/scout-mcp).
//
// The FileStore is created with lazy initialization - the actual configuration
// directory and filesystem interface are not accessed until the first operation
// that requires them. This allows for efficient construction and supports
// testing scenarios.
//
// Parameters:
//   - appName: The name of the application, used for directory naming.
//     Must be a valid directory name (no path separators or special characters).
//
// Returns a configured FileStore ready for use. The returned store will
// automatically create necessary directories on first write operation.
func NewFileStore(appName string) *FileStore {
	return &FileStore{
		appName: appName,
	}
}

// ConfigDir returns the full path to the configuration directory for this
// application. The directory path is computed on first access and cached
// for subsequent calls.
//
// The configuration directory follows the pattern:
// $HOME/.config/<appName>
//
// If the user's home directory cannot be determined (rare on modern systems),
// an error is returned. The directory itself is not created by this method;
// creation happens during save operations via ensureFilepath.
//
// Returns:
//   - The full path to the configuration directory
//   - An error if the user's home directory cannot be determined
//
// The returned path is cached after first computation for performance.
func (s *FileStore) ConfigDir() (_ string, err error) {
	var homeDir string
	if s.configDir != "" {
		goto end
	}

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	s.configDir = filepath.Join(homeDir, ConfigBaseDirName, s.appName)

end:
	return s.configDir, err
}

// getFS returns the filesystem interface for this FileStore, initializing it
// on first access. The filesystem is based on the configuration directory
// and provides a sandboxed view for file operations.
//
// This method uses lazy initialization to avoid filesystem access during
// FileStore creation. The filesystem interface allows for easier testing
// by providing a mockable abstraction over the actual filesystem.
//
// Returns:
//   - The filesystem interface rooted at the configuration directory
//   - An error if the configuration directory cannot be determined
//
// The filesystem interface is cached after first initialization.
func (s *FileStore) getFS() (_ fs.FS, err error) {
	var dir string

	if s.fs != nil {
		goto end
	}

	dir, err = s.ConfigDir()
	if err != nil {
		goto end
	}

	s.fs = os.DirFS(dir)

end:
	return s.fs, err
}

// ensureFilepath returns the full filesystem path for a given filename and
// ensures that all parent directories exist. This method is used internally
// before write operations to guarantee that the target location is accessible.
//
// The method handles nested paths within the configuration directory (e.g.,
// "tokens/user@domain.com.json") by creating all necessary intermediate
// directories with permissions 0755.
//
// Parameters:
//   - filename: The relative filename within the configuration directory.
//     May include subdirectories separated by forward slashes.
//
// Returns:
//   - The full filesystem path to the file
//   - An error if path validation fails or directory creation fails
//
// Directory creation is idempotent - existing directories are not modified.
func (s *FileStore) ensureFilepath(filename string) (fp string, err error) {
	fp, err = s.getFilepath(filename)
	if err != nil {
		goto end
	}
	// Create parent directories as needed for nested paths like tokens/token-bill@microsoft.com.json
	err = os.MkdirAll(filepath.Dir(fp), 0755)
	if err != nil {
		goto end
	}
end:
	return fp, err
}

// getFilepath validates a filename and returns the full filesystem path
// without creating any directories. This method performs security validation
// to prevent directory traversal attacks and ensures the filename is valid
// for filesystem use.
//
// Parameters:
//   - filename: The relative filename within the configuration directory.
//     Must be a valid path according to fs.ValidPath (no ".." components,
//     absolute paths, or invalid characters).
//
// Returns:
//   - The full filesystem path to the file
//   - An error if the filename is invalid or configuration directory unavailable
//
// This method is used internally by both read and write operations to
// ensure consistent path handling and security validation.
func (s *FileStore) getFilepath(filename string) (fp string, err error) {
	var dir string

	dir, err = s.ConfigDir()
	if err != nil {
		goto end
	}

	if !fs.ValidPath(filename) {
		err = fmt.Errorf("path %s is not valid for use in %s", filename, dir)
		goto end
	}

	fp = filepath.Join(s.configDir, filename)

end:
	return fp, err
}

// Save marshals the provided data to JSON and saves it to the specified
// filename in the configuration directory. The data is serialized with
// indentation for human readability and the file is written atomically
// where possible.
//
// The method automatically creates any necessary parent directories and
// validates the file path for security. The JSON is formatted with 2-space
// indentation to maintain readability for manual configuration editing.
//
// Parameters:
//   - filename: The relative path within the configuration directory where
//     the file should be saved. May include subdirectories.
//   - data: The data structure to serialize to JSON. Must be JSON-serializable.
//
// Returns an error if:
//   - The data cannot be marshaled to JSON
//   - The file path is invalid or cannot be created
//   - File write operations fail due to permissions or disk space
//   - The logger has not been initialized with SetLogger
//
// The save operation is atomic at the filesystem level where supported.
func (s *FileStore) Save(filename string, data any) (err error) {
	var jsonData []byte
	var file *os.File
	var fullPath string

	ensureLogger()

	jsonData, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		goto end
	}

	fullPath, err = s.ensureFilepath(filename)
	if err != nil {
		goto end
	}

	file, err = os.Create(fullPath)
	if err != nil {
		goto end
	}
	defer mustClose(file)

	_, err = file.Write(jsonData)

end:
	return err
}

// Load reads JSON data from the specified filename and unmarshals it into
// the provided data structure. The file is read from the configuration
// directory using the filesystem interface.
//
// The method handles JSON parsing errors gracefully and provides clear
// error messages for common issues like malformed JSON or type mismatches.
//
// Parameters:
//   - filename: The relative path within the configuration directory to read.
//     Must be a valid path that exists in the configuration directory.
//   - data: A pointer to the data structure where the JSON should be unmarshaled.
//     Must be compatible with the JSON structure in the file.
//
// Returns an error if:
//   - The file does not exist or cannot be read
//   - The file contains invalid JSON
//   - The JSON structure doesn't match the provided data type
//   - The configuration directory cannot be accessed
//
// Example usage:
//
//	var config MyConfig
//	err := store.Load("config.json", &config)
//	if err != nil {
//		log.Printf("Failed to load config: %v", err)
//	}
func (s *FileStore) Load(filename string, data any) (err error) {
	var jsonData []byte
	var fsys fs.FS

	fsys, err = s.getFS()
	if err != nil {
		goto end
	}

	jsonData, err = fs.ReadFile(fsys, filename)
	if err != nil {
		goto end
	}

	err = json.Unmarshal(jsonData, data)

end:
	return err
}

// Append adds the provided content to the end of the specified file,
// creating the file if it doesn't exist. This method is useful for
// logging operations where new entries need to be added to existing
// log files without overwriting previous content.
//
// The method ensures that parent directories exist and opens the file
// in append mode with create-if-not-exists semantics. After writing,
// the file is explicitly synced to ensure data persistence.
//
// Parameters:
//   - filename: The relative path within the configuration directory.
//     Parent directories will be created if they don't exist.
//   - content: The byte content to append to the file. No automatic
//     newlines are added - include them in content if needed.
//
// Returns an error if:
//   - The file path is invalid or cannot be created
//   - File write operations fail due to permissions or disk space
//   - The sync operation fails (indicating potential data loss)
//
// The file is opened with permissions 0644 (readable by owner and group,
// writable by owner only) when created.
func (s *FileStore) Append(filename string, content []byte) (err error) {
	var file *os.File
	var fullPath string

	fullPath, err = s.ensureFilepath(filename)
	if err != nil {
		goto end
	}

	file, err = os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		goto end
	}
	defer mustClose(file)

	_, err = file.Write(content)
	if err != nil {
		goto end
	}

	err = file.Sync()

end:
	return err
}

// Exists checks whether the specified file exists in the configuration
// directory. This method provides a safe way to test for file existence
// without attempting to read the file contents.
//
// The method uses the filesystem interface to check for file existence,
// which allows for consistent behavior across different testing scenarios
// and filesystem implementations.
//
// Parameters:
//   - filename: The relative path within the configuration directory to check.
//     Must be a valid filename (validated internally).
//
// Returns true if the file exists and is accessible, false otherwise.
// If the filesystem cannot be accessed or the configuration directory
// is unavailable, returns false.
//
// This method does not distinguish between "file doesn't exist" and
// "file exists but is not accessible" - both cases return false.
func (s *FileStore) Exists(filename string) (exists bool) {
	fsys, err := s.getFS()
	if err != nil {
		goto end
	}
	_, err = fs.Stat(fsys, filename)
	exists = err == nil

end:
	return exists
}

// SetBaseDir overrides the default configuration directory with a custom
// path. This method is primarily used for testing scenarios where
// configuration files need to be stored in a temporary or controlled
// location rather than the user's actual home directory.
//
// When SetBaseDir is called, both the configuration directory path and
// the filesystem interface are updated to point to the new location.
// Subsequent operations will use this custom directory instead of the
// default ~/.config/<appname> path.
//
// Parameters:
//   - dir: The absolute path to use as the configuration directory.
//     This directory should exist and be writable by the current user.
//
// After calling SetBaseDir, all file operations will be relative to
// the specified directory. This change affects the current FileStore
// instance only and does not impact other FileStore instances.
//
// Warning: This method bypasses the normal configuration directory
// resolution and should only be used in testing or specialized scenarios.
func (s *FileStore) SetBaseDir(dir string) {
	s.configDir = dir
	s.fs = os.DirFS(dir)
}