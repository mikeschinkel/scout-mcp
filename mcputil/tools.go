package mcputil

import (
	"context"
	"encoding/json"
)

// ToolOptions contains options for defining a tool
type ToolOptions struct {
	Name        string
	Description string
	Properties  []Property
}

// ToolRequest represents a tool call request
type ToolRequest interface {
	RequireString(key string) (string, error)
	RequireInt(key string) (int, error)
	RequireFloat(key string) (float64, error)
	RequireBool(key string) (bool, error)
	GetString(key string, defaultValue string) string
	GetInt(key string, defaultValue int) int
	GetFloat(key string, defaultValue float64) float64
	GetBool(key string, defaultValue bool) bool
	GetArguments() map[string]any
}

// ToolResult represents the result of a tool call
type ToolResult interface {
	ToolResult() // Marker method
}

// ToolHandler is the function signature for tool handlers
type ToolHandler func(ctx context.Context, req ToolRequest) (ToolResult, error)

// ToolResult implementations
type textResult struct {
	text string
}

func (textResult) ToolResult() {}

type errorResult struct {
	message string
}

func (errorResult) ToolResult() {}

// NewToolResultText creates a text result for a tool call
func NewToolResultText(text string) ToolResult {
	return &textResult{text: text}
}

// NewToolResultError creates an error result for a tool call
func NewToolResultError(message string) ToolResult {
	return &errorResult{message: message}
}

// NewToolResultJSON creates a JSON result for a tool call
func NewToolResultJSON(data any) ToolResult {
	jsonData, _ := json.Marshal(data)
	return &textResult{text: string(jsonData)}
}
