package mcputil

import (
	"fmt"
)

// ToolRequest provides access to tool call parameters with additional array support
type ToolRequest interface {
	CallToolRequest() CallToolRequest
	RequireArray(key string) ([]any, error)
	GetArray(key string, defaultValue []any) []any
}

// toolRequest implements ToolRequest interface by wrapping CallToolRequest.
// This wrapper provides additional functionality for array handling and
// parameter extraction beyond the basic MCP protocol types.
type toolRequest struct {
	req CallToolRequest
}

// CallToolRequest returns the underlying CallToolRequest.
// This method provides access to the raw MCP protocol request data.
func (w *toolRequest) CallToolRequest() CallToolRequest {
	return w.req
}

// RequireArray extracts a required array parameter from the tool request.
// This method returns an error if the parameter is missing or not an array.
func (w *toolRequest) RequireArray(key string) (array []any, err error) {
	var ok bool

	raw, exists := w.req.GetArguments()[key]
	if !exists {
		err = fmt.Errorf("required parameter %s not found", key)
		goto end
	}

	array, ok = raw.([]any)
	if !ok {
		err = fmt.Errorf("parameter %s must be an array", key)
		goto end
	}
end:
	return array, err
}

// GetArray extracts an optional array parameter with a default value.
// This method returns the default value if the parameter is missing or invalid.
func (w *toolRequest) GetArray(key string, defaultValue []any) (array []any) {
	var err error
	array, err = w.RequireArray(key)
	if err != nil {
		array = defaultValue
	}
	return array
}
