package mcputil

import (
	"strings"
)

// registeredTools holds all tools that have been registered with the MCP server.
// Tools are registered during package initialization via init() functions.
var registeredTools []Tool

// RegisteredTools returns a slice of all registered tools.
// This is used by the MCP server to get the complete list of available tools.
func RegisteredTools() []Tool {
	return registeredTools
}

// RegisteredToolsMap returns a map of tool names to Tool instances
// for efficient tool lookup by name.
func RegisteredToolsMap() (m map[string]Tool) {
	m = make(map[string]Tool, len(registeredTools))
	for _, tool := range registeredTools {
		m[tool.Name()] = tool
	}
	return m
}

// RegisterTool adds a tool to the global registry.
// This is typically called during package initialization.
func RegisterTool(tool Tool) {
	registeredTools = append(registeredTools, tool)
}

// GetRegisteredTool finds a registered tool by name (case-insensitive).
// Returns nil if no tool with the given name is found.
func GetRegisteredTool(name string) (t Tool) {
	name = strings.ToLower(name)
	for _, tool := range registeredTools {
		if tool.Name() != name {
			continue
		}
		t = tool
		goto end
	}
end:
	return t
}

// GetRegisteredToolNames returns the names of all registered tools
func GetRegisteredToolNames() []string {
	if registeredTools == nil {
		panic("No Scout MCP server tools have been registered.\nDid you forget to import github.com/mikeschinkel/scout-mcp/mcptools for side effects (by prefixing it with '_')?")
	}
	names := make([]string, 0, len(registeredTools))
	for _, tool := range registeredTools {
		names = append(names, tool.Name())
	}
	return names
}
