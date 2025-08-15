package scout

// ToolDefinition describes an MCP tool's metadata including its name,
// description, and input schema for parameter validation.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ClientInfo contains information about the MCP client (typically Claude Desktop)
// that is connecting to the server.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeParams contains the parameters sent during MCP connection initialization,
// including protocol version, client capabilities, and client information.
type InitializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      ClientInfo     `json:"clientInfo"`
}

// ServerCapabilities describes what functionality this MCP server supports,
// including available tools, resources, prompts, and logging capabilities.
type ServerCapabilities struct {
	Tools     []ToolDefinition `json:"tools"`
	Resources map[string]any   `json:"resources,omitempty"`
	Prompts   map[string]any   `json:"prompts,omitempty"`
	Logging   map[string]any   `json:"logging,omitempty"`
}

// ToolsCapability describes the server's tool-related capabilities,
// particularly whether the tool list can change dynamically.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Implementation contains server implementation details including
// name and version information.
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult is the response sent to the client during MCP initialization,
// containing server capabilities, protocol version, and optional instructions.
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

// MCPServerOpts contains options for configuring MCP server behavior.
type MCPServerOpts struct {
	OnlyMode bool
}
