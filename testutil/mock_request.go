package testutil

import (
	"fmt"
	"strconv"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// MockRequest implements mcputil.ToolRequest for testing
type MockRequest struct {
	params map[string]any
}

func (m *MockRequest) RequireString(key string) (string, error) {
	if val, exists := m.params[key]; exists {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}
	return "", fmt.Errorf("required argument %q not found", key)
}

func (m *MockRequest) RequireInt(key string) (result int, err error) {
	var val any
	var exists bool
	var i int
	var f float64
	var str string
	var ok bool
	var parseErr error

	val, exists = m.params[key]
	if !exists {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

	i, ok = val.(int)
	if ok {
		result = i
		goto end
	}

	f, ok = val.(float64)
	if ok {
		result = int(f)
		goto end
	}

	str, ok = val.(string)
	if !ok {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

	result, parseErr = strconv.Atoi(str)
	if parseErr != nil {
		err = fmt.Errorf("'%s' must be a valid number: %w", key, parseErr)
		goto end
	}

end:
	return result, err
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
	if val, ok := m.params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
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

func (m *MockRequest) GetArray(key string, defaultValue []any) (a []any) {
	var ok bool
	var arr []any
	var val any

	a = defaultValue
	val, ok = m.params[key]
	if !ok {
		goto end
	}
	arr = convertSlice(val)
	if arr != nil {
		a = arr
	}
end:
	return a
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
