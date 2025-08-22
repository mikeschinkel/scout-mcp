package mcputil

// Standard required properties for MCP tools.
var (
	RequiredSessionTokenProperty = SessionTokenProperty.Required()
)

// Common property definitions used across multiple MCP tools.
var (
	ToolProperty         = String("tool", "Tool name for help documentation")
	SessionTokenProperty = String("session_token", "Session token from start_session")
)
