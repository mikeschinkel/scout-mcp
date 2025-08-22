package mcputil

import (
	"path/filepath"
)

// MockConfig implements the Config interface for testing purposes.
// It provides a simplified configuration that allows specified paths
// and returns default values for other configuration settings.
type MockConfig struct {
	allowedPaths []string // Paths that tools are allowed to access
}

// MockConfigArgs contains the arguments for creating a MockConfig instance.
// This struct allows tests to specify which paths should be allowed
// for file operations during testing.
type MockConfigArgs struct {
	AllowedPaths []string // List of paths that should be allowed for testing
}

// NewMockConfig creates a mock config with specified allowed paths.
// This constructor is used in unit tests to provide controlled configuration settings.
func NewMockConfig(args MockConfigArgs) Config {
	return &MockConfig{
		allowedPaths: args.AllowedPaths,
	}
}

// IsAllowedPath checks if a path is allowed based on the mock configuration.
// This method implements the Config interface for testing purposes.
func (m *MockConfig) IsAllowedPath(path string) (allowed bool) {
	// For tests, allow any path that starts with one of our allowed paths
	for _, allowedPath := range m.allowedPaths {
		_, err := filepath.Rel(allowedPath, path)
		if err != nil {
			goto end
		}
	}
	allowed = true
end:
	return allowed
}

// Path returns the first allowed path or empty string if none are configured.
// This method implements the Config interface for testing purposes.
func (m *MockConfig) Path() string {
	if len(m.allowedPaths) > 0 {
		return m.allowedPaths[0]
	}
	return ""
}

// AllowedPaths returns the list of allowed paths configured for this mock.
// This method implements the Config interface for testing purposes.
func (m *MockConfig) AllowedPaths() []string {
	return m.allowedPaths
}

// ServerPort returns a mock server port for testing.
// This method implements the Config interface for testing purposes.
func (m *MockConfig) ServerPort() string {
	return "8080"
}

// ServerName returns a mock server name for testing.
// This method implements the Config interface for testing purposes.
func (m *MockConfig) ServerName() string {
	return "test-server"
}

// AllowedOrigins returns mock allowed origins for testing.
// This method implements the Config interface for testing purposes.
func (m *MockConfig) AllowedOrigins() []string {
	return []string{"localhost"}
}

// ToMap converts the mock configuration to a map representation.
// This method implements the Config interface for testing purposes.
func (m *MockConfig) ToMap() (map[string]any, error) {
	return map[string]any{
		"allowedPaths":   m.allowedPaths,
		"serverPort":     m.ServerPort(),
		"serverName":     m.ServerName(),
		"allowedOrigins": m.AllowedOrigins(),
	}, nil
}
