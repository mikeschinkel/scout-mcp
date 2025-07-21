package testutil

import (
	"fmt"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// MockRequest implements mcputil.ToolRequest for testing
type MockRequest struct {
	params map[string]interface{}
}

func (m *MockRequest) RequireString(key string) (string, error) {
	if val, exists := m.params[key]; exists {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}
	return "", fmt.Errorf("required argument %q not found", key)
}

func (m *MockRequest) RequireInt(key string) (int, error) {
	if val, exists := m.params[key]; exists {
		if i, ok := val.(int); ok {
			return i, nil
		}
		if str, ok := val.(string); ok {
			// Try to parse string as int for convenience
			if str == "1" {
				return 1, nil
			}
			if str == "2" {
				return 2, nil
			}
			if str == "3" {
				return 3, nil
			}
			if str == "4" {
				return 4, nil
			}
		}
	}
	return 0, fmt.Errorf("required argument %q not found", key)
}

func (m *MockRequest) RequireBool(key string) (bool, error) {
	if val, exists := m.params[key]; exists {
		if b, ok := val.(bool); ok {
			return b, nil
		}
	}
	return false, fmt.Errorf("required argument %q not found", key)
}

func (m *MockRequest) RequireFloat(key string) (float64, error) {
	if val, exists := m.params[key]; exists {
		if f, ok := val.(float64); ok {
			return f, nil
		}
	}
	return 0, fmt.Errorf("required argument %q not found", key)
}

func (m *MockRequest) GetString(key, defaultValue string) string {
	if val, exists := m.params[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (m *MockRequest) GetInt(key string, defaultValue int) int {
	if val, exists := m.params[key]; exists {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return defaultValue
}

func (m *MockRequest) GetBool(key string, defaultValue bool) bool {
	if val, exists := m.params[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func (m *MockRequest) GetFloat(key string, defaultValue float64) float64 {
	if val, exists := m.params[key]; exists {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultValue
}

func (m *MockRequest) GetArray(key string, defaultValue []any) []any {
	if val, exists := m.params[key]; exists {
		if arr, ok := val.([]any); ok {
			return arr
		}
	}
	return defaultValue
}

func (m *MockRequest) GetArguments() map[string]any {
	return m.params
}

func (m *MockRequest) RequireArray(key string) ([]any, error) {
	if val, exists := m.params[key]; exists {
		if arr, ok := val.([]any); ok {
			return arr, nil
		}
	}
	return nil, fmt.Errorf("required argument %q not found", key)
}

// NewMockRequest creates a mock request with the specified parameters
func NewMockRequest(params map[string]interface{}) mcputil.ToolRequest {
	return &MockRequest{params: params}
}
