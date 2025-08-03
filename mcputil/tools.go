package mcputil

import (
	"context"
	"encoding/json"
)

// ToolHandler is the function signature for tool handlers
type ToolHandler func(context.Context, ToolRequest) (ToolResult, error)

type Config interface {
	IsAllowedPath(string) bool
	Path() string
	AllowedPaths() []string
	ServerPort() string
	ServerName() string
	AllowedOrigins() []string
	ToMap() (map[string]any, error)
}

type Tool interface {
	Name() string
	Options() ToolOptions
	Handle(context.Context, ToolRequest) (ToolResult, error)
	EnsurePreconditions(ToolRequest) error
	SetConfig(c Config)
	HasRequiredParams() bool
}

// ToolOptions contains options for defining a tool
type ToolOptions struct {
	Name        string
	Description string
	Properties  []Property
	Requires    []Requirement // Complex parameter requirements
	QuickHelp   string        // Short description for quick help list (empty = not included)
}

// Requirement interface for declarative parameter requirements
type Requirement interface {
	RequirementOption() // Marker method
	IsSatisfied(ToolRequest) bool
	Description() string
}

// ToolResult represents the result of a tool call
type ToolResult interface {
	ToolResult()   // Marker method
	Value() string // Get the actual result value
}

// ToolResult implementations
type jsonResult struct {
	json string
}

func (*jsonResult) ToolResult() {}

func (t *jsonResult) Value() string {
	return t.json
}

type errorResult struct {
	message string
}

func (*errorResult) ToolResult() {}

func (e *errorResult) Value() string {
	return e.message
}

// NewToolResultError creates an error result for a tool call
func NewToolResultError(err error) ToolResult {
	return &errorResult{message: err.Error()}
}

// NewToolResultJSON creates a JSON result for a tool call
func NewToolResultJSON(data any) ToolResult {
	jsonData, _ := json.Marshal(data)
	return &jsonResult{json: string(jsonData)}
}
