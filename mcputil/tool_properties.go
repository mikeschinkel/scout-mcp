package mcputil

var (
	RequiredSessionTokenProperty = SessionTokenProperty.Required()
)

var (
	ToolProperty         = String("tool", "Tool name for help documentation")
	SessionTokenProperty = String("session_token", "Session token from start_session")
)
