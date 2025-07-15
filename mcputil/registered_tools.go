package mcputil

var registeredTools []Tool

func RegisteredTools() []Tool {
	return registeredTools
}

func RegisterTool(tool Tool) {
	registeredTools = append(registeredTools, tool)
}
