package mcputil

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolHandler is the function signature for tool handlers
type ToolHandler func(context.Context, ToolRequest) (ToolResult, error)

// Config provides access to server configuration including allowed paths,
// port settings, and security restrictions. Tools receive a Config instance
// to validate operations against the server's security policies.
type Config interface {
	IsAllowedPath(string) bool
	Path() string
	AllowedPaths() []string
	ServerPort() string
	ServerName() string
	AllowedOrigins() []string
	ToMap() (map[string]any, error)
}

// Tool represents an MCP tool that can be invoked by clients.
// Tools must implement session validation, parameter checking,
// and operation handling with appropriate security controls.
type Tool interface {
	Name() string
	Options() ToolOptions
	Handle(context.Context, ToolRequest) (ToolResult, error)
	EnsurePreconditions(context.Context, ToolRequest) error
	SetConfig(Config)
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

// jsonResult implements ToolResult for JSON responses.
// This type wraps JSON data for successful tool execution results
// that need to be returned to the MCP client.
type jsonResult struct {
	json string
}

// NewToolResultJSON creates a JSON result for a tool call.
// This function serializes the provided data to JSON and wraps it in a ToolResult.
func NewToolResultJSON(data any) ToolResult {
	jsonData, _ := json.Marshal(data)
	return &jsonResult{json: string(jsonData)}
}

// ToolResult implements the ToolResult interface marker method.
func (*jsonResult) ToolResult() {}

// Value returns the JSON string representation of the result.
func (t *jsonResult) Value() string {
	return t.json
}

// errorResult implements ToolResult for error responses.
// This type wraps error messages for failed tool execution results
// that should be returned as errors to the MCP client.
type errorResult struct {
	message string
}

// NewToolResultError creates an error result for a tool call.
// This function wraps an error message in a ToolResult for failed operations.
func NewToolResultError(err error) ToolResult {
	return &errorResult{message: err.Error()}
}

// ToolResult implements the ToolResult interface marker method.
func (*errorResult) ToolResult() {}

// Value returns the error message string.
func (e *errorResult) Value() string {
	return e.message
}

// InternalError represents system-level errors that should be returned as errors
type InternalError struct {
	message string
	cause   error
}

// Error returns the formatted error message.
// This method implements the error interface for InternalError.
func (e *InternalError) Error() string {
	return e.message
}

// Unwrap returns the underlying cause error for error wrapping.
// This method supports Go's error unwrapping functionality.
func (e *InternalError) Unwrap() error {
	return e.cause
}

// NewInternalError creates a system-level error with a formatted message.
// This function wraps a cause error with additional context for debugging.
func NewInternalError(cause error, format string, args ...any) *InternalError {
	return &InternalError{
		cause:   cause,
		message: fmt.Sprintf(format, args...),
	}
}
