package mcputil

import (
	"github.com/mark3labs/mcp-go/mcp"
)

type (
	mcpPropertyOption = mcp.PropertyOption
	mcpToolOption     = mcp.ToolOption
	mcpTool           = mcp.Tool
	CallToolRequest   = mcp.CallToolRequest
	CallToolResult    = mcp.CallToolResult
	CallToolParams    = mcp.CallToolParams
)

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

func mcpDefaultArray[T any](value []T) mcpPropertyOption {
	return mcp.DefaultArray(value)
}
