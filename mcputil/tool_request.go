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

// toolRequest implements ToolRequest interface by wrapping CallToolRequest
type toolRequest struct {
	req CallToolRequest
}

// CallToolRequest returns the underlying CallToolRequest
func (w *toolRequest) CallToolRequest() CallToolRequest {
	return w.req
}

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

func (w *toolRequest) GetArray(key string, defaultValue []any) (array []any) {
	var err error
	array, err = w.RequireArray(key)
	if err != nil {
		array = defaultValue
	}
	return array
}
