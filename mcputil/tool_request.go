package mcputil

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// toolRequest implements ToolRequest interface by wrapping mcp.CallToolRequest
type toolRequest struct {
	req mcp.CallToolRequest
}

func (w *toolRequest) RequireString(key string) (string, error) {
	return w.req.RequireString(key)
}

func (w *toolRequest) RequireInt(key string) (int, error) {
	return w.req.RequireInt(key)
}

func (w *toolRequest) RequireFloat(key string) (float64, error) {
	return w.req.RequireFloat(key)
}

func (w *toolRequest) RequireBool(key string) (bool, error) {
	return w.req.RequireBool(key)
}

func (w *toolRequest) GetString(key string, defaultValue string) string {
	return w.req.GetString(key, defaultValue)
}

func (w *toolRequest) GetInt(key string, defaultValue int) int {
	return w.req.GetInt(key, defaultValue)
}

func (w *toolRequest) GetFloat(key string, defaultValue float64) float64 {
	return w.req.GetFloat(key, defaultValue)
}

func (w *toolRequest) GetBool(key string, defaultValue bool) bool {
	return w.req.GetBool(key, defaultValue)
}

func (w *toolRequest) GetArguments() map[string]any {
	return w.req.GetArguments()
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
