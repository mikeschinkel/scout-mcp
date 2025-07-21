package testutil

import (
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// MockConfig implements mcputil.Config for testing
type MockConfig struct {
	allowedPaths []string
}

func (m *MockConfig) IsAllowedPath(path string) (bool, error) {
	// For tests, allow any path that starts with one of our allowed paths
	for _, allowedPath := range m.allowedPaths {
		if filepath.HasPrefix(path, allowedPath) {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockConfig) Path() string {
	if len(m.allowedPaths) > 0 {
		return m.allowedPaths[0]
	}
	return ""
}

func (m *MockConfig) AllowedPaths() []string {
	return m.allowedPaths
}

func (m *MockConfig) ServerPort() string {
	return "8080"
}

func (m *MockConfig) ServerName() string {
	return "test-server"
}

func (m *MockConfig) AllowedOrigins() []string {
	return []string{"localhost"}
}

func (m *MockConfig) ToMap() (map[string]any, error) {
	return map[string]any{
		"allowedPaths":   m.allowedPaths,
		"serverPort":     m.ServerPort(),
		"serverName":     m.ServerName(),
		"allowedOrigins": m.AllowedOrigins(),
	}, nil
}

// NewMockConfig creates a mock config with specified allowed paths
func NewMockConfig(allowedPaths []string) mcputil.Config {
	return &MockConfig{allowedPaths: allowedPaths}
}
