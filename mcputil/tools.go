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
}

// ToolOptions contains options for defining a tool
type ToolOptions struct {
	Name        string
	Description string
	Properties  []Property
	QuickHelp   string // Short description for quick help list (empty = not included)
}

// ToolRequest represents a tool call request
// TODO: Why do we need both Request*() and Get*() methods?
//
//	Seems Get has a default and ignores errors (bad software engineering) and
//	Request has errors but no default? Why not just have Get*() with defaults that
//	returns errors? Oh. Wait! It seems we have this because of the shitty API
//	created bv mcp-go? Well, let's clean it up, not replicate his bad design
//	decisions.
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
	RequireArray(key string) ([]any, error)
	GetArray(key string, defaultValue []any) []any
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
