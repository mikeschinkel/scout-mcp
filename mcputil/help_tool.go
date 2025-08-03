package mcputil

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"
)

//go:embed HELP_TOOL_CONTENT.md
var helpToolContent string

var _ Tool = (*HelpTool)(nil)

// NewHelpTool creates a new help tool with optional payload for additional content
func NewHelpTool(payload Payload) *HelpTool {
	return &HelpTool{
		ToolBase: NewToolBase(ToolOptions{
			Name:        "help",
			Description: "Get detailed documentation for MCP tools and best practices",
			Properties: []Property{
				String("session_token", "Session token from start_session").Required(),
				ToolProperty,
			},
		}),
		Payload: payload,
	}
}

type HelpTool struct {
	*ToolBase
	Payload Payload
}

func (t *HelpTool) Handle(_ context.Context, req ToolRequest) (result ToolResult, err error) {
	var toolName string
	var helpContent string
	var ptn string

	logger.Info("Tool called", "tool", "help")

	toolName, err = ToolProperty.String(req)
	if err != nil {
		goto end
	}

	// Initialize payload if available
	if t.Payload != nil {
		ptn = reflect.TypeOf(t.Payload).Elem().String()
		err = t.Payload.Initialize(t, req)
		if err != nil {
			result = NewToolResultError(err)
			goto end
		}
	}

	if toolName == "" {
		// Return full documentation with generated tools section
		helpContent, err = t.generateFullDocumentation()
		if err != nil {
			result = NewToolResultError(fmt.Errorf("failed to generate full documentation: %v", err))
			goto end
		}
		logger.Info("Tool completed", "tool", "help", "type", "full_documentation")
	} else {
		// Return tool-specific documentation
		helpContent = t.getToolSpecificHelp(toolName)
		logger.Info("Tool completed", "tool", "help", "type", "tool_specific", "requested_tool", toolName)
	}

	result = NewToolResultJSON(map[string]any{
		"tool":    toolName,
		"content": helpContent,
		"type": func() string {
			if toolName == "" {
				return "full_documentation"
			}
			return "tool_specific"
		}(),
		"payload_type_name": ptn,
		"payload":           t.Payload,
	})

end:
	return result, err
}

func (t *HelpTool) getToolSpecificHelp(toolName string) (helpText string) {
	var tools []Tool
	var tool Tool
	var options ToolOptions
	var prop Property
	var propOptions []PropertyOption
	var found bool

	tools = RegisteredTools()
	for _, tool = range tools {
		options = tool.Options()
		if options.Name == toolName {
			// Found the tool, generate its documentation
			helpText = fmt.Sprintf("## %s\n\n%s\n\n", options.Name, options.Description)

			if len(options.Properties) > 0 {
				helpText += "### Parameters\n\n"
				for _, prop = range options.Properties {
					propOptions = prop.PropertyOptions()
					helpText += fmt.Sprintf("- **%s** (%s)", prop.GetName(), prop.GetType())

					// Check if required
					for _, opt := range propOptions {
						p, ok := opt.(RequiredProperty)
						if ok && p.Required {
							helpText += " *[required]*"
							break
						}
					}

					helpText += fmt.Sprintf(": %s\n", t.getPropertyDescription(prop))
				}
				helpText += "\n"
			}
			found = true
			break
		}
	}

	if !found {
		helpText = t.getToolNotFoundHelp(toolName)
	}

	return helpText
}

func (t *HelpTool) getToolNotFoundHelp(toolName string) (helpText string) {
	var tools []Tool
	var tool Tool
	var options ToolOptions

	helpText = "Tool '" + toolName + "' not found.\n\n"
	helpText += "Available tools:\n"

	tools = RegisteredTools()
	for _, tool = range tools {
		options = tool.Options()
		helpText += "- " + options.Name + "\n"
	}

	helpText += "\nCall help without parameters to see full documentation, or specify a tool name:\n"
	helpText += `{"tool": "help", "parameters": {"tool": "start_session"}}`

	return helpText
}

// generateFullDocumentation combines embedded content with generated tool documentation
func (t *HelpTool) generateFullDocumentation() (helpText string, err error) {
	var generatedTools string

	// Start with embedded content
	helpText = helpToolContent

	// Generate tools section
	generatedTools, err = t.generateToolHelp()
	if err != nil {
		goto end
	}

	// Append generated tools section
	helpText += "\n\n" + generatedTools

end:
	return helpText, err
}

// generateToolHelp creates tool documentation from registered tools
// with start_session first and help last
func (t *HelpTool) generateToolHelp() (helpText string, err error) {
	var tools []Tool
	var tool Tool
	var options ToolOptions
	var prop Property
	var propOptions []PropertyOption

	tools = RegisteredTools()
	helpText = ""

	// Order tools: start_session first, help last, others in between
	var startSessionTool Tool
	var helpTool Tool
	var otherTools []Tool

	for _, tool = range tools {
		options = tool.Options()
		switch options.Name {
		case "start_session":
			startSessionTool = tool
		case "help":
			helpTool = tool
		default:
			otherTools = append(otherTools, tool)
		}
	}

	// Generate documentation in order: start_session, others, help
	orderedTools := make([]Tool, 0, len(tools))
	if startSessionTool != nil {
		orderedTools = append(orderedTools, startSessionTool)
	}
	orderedTools = append(orderedTools, otherTools...)
	if helpTool != nil {
		orderedTools = append(orderedTools, helpTool)
	}

	for _, tool = range orderedTools {
		options = tool.Options()
		helpText += fmt.Sprintf("## %s\n\n%s\n\n", options.Name, options.Description)

		if len(options.Properties) > 0 {
			helpText += "### Parameters\n\n"
			for _, prop = range options.Properties {
				propOptions = prop.PropertyOptions()
				helpText += fmt.Sprintf("- **%s** (%s)", prop.GetName(), prop.GetType())

				// Check if required
				for _, opt := range propOptions {
					p, ok := opt.(RequiredProperty)
					if ok && p.Required {
						helpText += " *[required]*"
						break
					}
				}

				helpText += fmt.Sprintf(": %s\n", t.getPropertyDescription(prop))
			}
			helpText += "\n"
		}
	}

	return helpText, err
}

func (t *HelpTool) getPropertyDescription(prop Property) string {
	// Extract description from property options
	var options []PropertyOption
	var opt PropertyOption

	options = prop.PropertyOptions()
	for _, opt = range options {
		p, ok := opt.(DescriptionProperty)
		if ok {
			return p.Description
		}
	}
	return "No description available"
}
