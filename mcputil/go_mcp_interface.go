// Package mcputil provides a comprehensive framework for building Model Context Protocol (MCP) servers.
//
// This package includes:
//   - MCP server abstraction with session management and security controls
//   - Property system for defining tool parameters with validation
//   - Tool base classes with automatic session validation
//   - Request/response handling with JSON-RPC 2.0 compliance
//   - Risk assessment and approval workflows for secure file operations
//   - Testing utilities and mock implementations
//
// The package wraps the mark3labs/mcp-go library to provide a higher-level interface
// for building secure MCP tools with consistent session management, path validation,
// and error handling patterns.
//
// Key Components:
//   - Server: MCP server interface with stdio transport support
//   - Tool: Interface for MCP tools with session validation
//   - Property: Type-safe parameter definitions with validation
//   - ToolBase: Base implementation providing common tool functionality
//   - Session management: Token-based authentication with expiration
//   - Config: Server configuration with path whitelisting
//
// This file contains type aliases and variable assignments to provide a cleaner
// interface while maintaining compatibility with the underlying MCP library.
package mcputil

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// Type aliases for mark3labs/mcp-go types to provide cleaner imports.
type (
	mcpPropertyOption = mcp.PropertyOption
	mcpToolOption     = mcp.ToolOption
	mcpTool           = mcp.Tool
	CallToolRequest   = mcp.CallToolRequest
	CallToolResult    = mcp.CallToolResult
	CallToolParams    = mcp.CallToolParams
)

// Variable assignments for mark3labs/mcp-go functions to provide cleaner access.
var (
	mcpDescription        = mcp.Description
	mcpDefaultString      = mcp.DefaultString
	mcpRequired           = mcp.Required
	mcpPattern            = mcp.Pattern
	mcpEnum               = mcp.Enum
	mcpMin                = mcp.Min
	mcpMax                = mcp.Max
	mcpMinLength          = mcp.MinLength
	mcpMaxLength          = mcp.MaxLength
	mcpMinItems           = mcp.MinItems
	mcpMaxItems           = mcp.MaxItems
	mcpWithArray          = mcp.WithArray
	mcpWithString         = mcp.WithString
	mcpWithBoolean        = mcp.WithBoolean
	mcpWithNumber         = mcp.WithNumber
	mcpWithDescription    = mcp.WithDescription
	mcpNewTool            = mcp.NewTool
	mcpNewToolResultError = mcp.NewToolResultError
	mcpNewToolResultText  = mcp.NewToolResultText
	mcpDefaultNumber      = mcp.DefaultNumber
)

// mcpDefaultArray wraps mcp.DefaultArray to provide a generic default array function.
// This function enables type-safe default array values for array-type properties
// while maintaining compatibility with the underlying MCP library interface.
func mcpDefaultArray[T any](value []T) mcpPropertyOption {
	return mcp.DefaultArray(value)
}
