package mcputil

import (
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
