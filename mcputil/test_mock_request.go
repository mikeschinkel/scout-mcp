package mcputil

import (
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockRequest implements the ToolRequest interface for testing purposes.
// It provides a simplified request implementation that allows tests to
// specify parameters directly without JSON-RPC protocol overhead.
type MockRequest struct {
	params map[string]any // Parameter map for tool testing
}

// RequireString extracts a required string parameter from the mock request.
// This method returns an error if the parameter is missing or not a string.
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

// RequireInt extracts a required integer parameter from the mock request.
// This method handles type conversion from int, float64, or string to int.
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

// RequireBool extracts a required boolean parameter from the mock request.
// This method returns an error if the parameter is missing or not a boolean.
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

// RequireFloat extracts a required float64 parameter from the mock request.
// This method returns an error if the parameter is missing or not a float64.
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

// GetString extracts an optional string parameter with a default value.
// This method returns the default value if the parameter is missing or invalid.
func (m *MockRequest) GetString(key, defaultValue string) string {
	if val, exists := m.params[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetInt extracts an optional integer parameter with a default value.
// This method handles type conversion and returns the default if conversion fails.
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

// GetBool extracts an optional boolean parameter with a default value.
// This method returns the default value if the parameter is missing or invalid.
func (m *MockRequest) GetBool(key string, defaultValue bool) bool {
	if val, exists := m.params[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// GetFloat extracts an optional float64 parameter with a default value.
// This method returns the default value if the parameter is missing or invalid.
func (m *MockRequest) GetFloat(key string, defaultValue float64) float64 {
	if val, exists := m.params[key]; exists {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultValue
}

// GetArray extracts an optional array parameter with a default value.
// This method uses ConvertContainedSlice for type conversion and returns the default if invalid.
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

// GetArguments returns the complete parameter map for the mock request.
// This method provides access to all parameters set in the mock request.
func (m *MockRequest) GetArguments() map[string]any {
	return m.params
}

// CallToolRequest returns a CallToolRequest containing the mock parameters.
// This method converts the mock request to the standard MCP protocol format.
func (m *MockRequest) CallToolRequest() CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: m.params,
		},
	}
}

// RequireArray extracts a required array parameter from the mock request.
// This method returns an error if the parameter is missing or not an array.
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

// Params is a type alias for parameter maps used in testing.
// This provides a convenient shorthand for map[string]any when
// creating mock requests with test parameters.
type Params = map[string]any

// NewMockRequest creates a mock request with the specified parameters.
// This function is used in unit tests to create ToolRequest instances
// with predefined parameter values for testing tool behavior.
func NewMockRequest(params Params) ToolRequest {
	return &MockRequest{params: params}
}
