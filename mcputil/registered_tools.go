package mcputil

import (
	"strings"
)

var registeredTools []Tool

func RegisteredTools() []Tool {
	return registeredTools
}
func RegisteredToolsMap() (m map[string]Tool) {
	m = make(map[string]Tool, len(registeredTools))
	for _, tool := range registeredTools {
		m[tool.Name()] = tool
	}
	return m
}

func RegisterTool(tool Tool) {
	registeredTools = append(registeredTools, tool)
}

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
