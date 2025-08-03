package mcputil

import (
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockRequest implements mcputil.ToolRequest for testing
type MockRequest struct {
	params map[string]any
}

func (m *MockRequest) RequireString(key string) (result string, err error) {
	var val any
	var exists bool
	var ok bool

	val, exists = m.params[key]
	if !exists {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

	result, ok = val.(string)
	if !ok {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

end:
	return result, err
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

func (m *MockRequest) RequireBool(key string) (result bool, err error) {
	var val any
	var exists bool
	var ok bool

	val, exists = m.params[key]
	if !exists {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

	result, ok = val.(bool)
	if !ok {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

end:
	return result, err
}

func (m *MockRequest) RequireFloat(key string) (result float64, err error) {
	var val any
	var exists bool
	var ok bool

	val, exists = m.params[key]
	if !exists {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

	result, ok = val.(float64)
	if !ok {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

end:
	return result, err
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
	arr = ConvertContainedSlice(val)
	if arr != nil {
		a = arr
	}
end:
	return a
}

func (m *MockRequest) GetArguments() map[string]any {
	return m.params
}

func (m *MockRequest) CallToolRequest() CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: m.params,
		},
	}
}

func (m *MockRequest) RequireArray(key string) (result []any, err error) {
	var val any
	var exists bool
	var ok bool

	val, exists = m.params[key]
	if !exists {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

	result, ok = val.([]any)
	if !ok {
		err = fmt.Errorf("required argument %q not found", key)
		goto end
	}

end:
	return result, err
}

type Params = map[string]any

// NewMockRequest creates a mock request with the specified parameters
func NewMockRequest(params Params) ToolRequest {
	return &MockRequest{params: params}
}
