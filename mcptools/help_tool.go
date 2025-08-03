package mcptools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

//go:embed README.md
var readmeContent string

var _ mcputil.Tool = (*HelpTool)(nil)

func init() {
	mcputil.RegisterTool(&HelpTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "help",
			Description: "Get detailed documentation for MCP tools and best practices",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				ToolProperty,
			},
		}),
	})
}

type HelpTool struct {
	*mcputil.ToolBase
}

func (t *HelpTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var toolName string
	var helpContent string

	logger.Info("Tool called", "tool", "help")

	toolName, err = ToolProperty.String(req)
	if err != nil {
		goto end
	}

	if toolName == "" {
		// Return full documentation
		helpContent = readmeContent
		logger.Info("Tool completed", "tool", "help", "type", "full_documentation")
	} else {
		// Return tool-specific documentation
		helpContent = t.getToolSpecificHelp(toolName)
		logger.Info("Tool completed", "tool", "help", "type", "tool_specific", "requested_tool", toolName)
	}

	result = mcputil.NewToolResultJSON(map[string]any{
		"tool":    toolName,
		"content": helpContent,
		"type": func() string {
			if toolName == "" {
				return "full_documentation"
			}
			return "tool_specific"
		}(),
	})

end:
	return result, err
}

func (t *HelpTool) getToolSpecificHelp(toolName string) (helpText string) {
	var sections map[string]string
	var found bool

	// Parse the README into tool-specific sections
	sections = t.parseToolSections()

	helpText, found = sections[toolName]
	if !found {
		helpText = t.getToolNotFoundHelp(toolName)
	}

	return helpText
}

func (t *HelpTool) parseToolSections() (sections map[string]string) {
	var lines []string
	var currentSection string
	var currentContent strings.Builder
	var inToolSection bool

	sections = make(map[string]string)
	lines = strings.Split(readmeContent, "\n")

	for _, line := range lines {
		// Check if this is a tool header (### `tool_name`)
		if strings.HasPrefix(line, "### `") && strings.HasSuffix(line, "`") {
			// Save previous section if we were in one
			if inToolSection && currentSection != "" {
				sections[currentSection] = strings.TrimSpace(currentContent.String())
			}

			// Start new section
			currentSection = t.extractToolName(line)
			currentContent.Reset()
			currentContent.WriteString(line + "\n")
			inToolSection = true
		} else if strings.HasPrefix(line, "## ") {
			// Save current section when we hit a major heading
			if inToolSection && currentSection != "" {
				sections[currentSection] = strings.TrimSpace(currentContent.String())
				inToolSection = false
			}
		} else if inToolSection {
			// Add line to current section
			currentContent.WriteString(line + "\n")
		}
	}

	// Save final section
	if inToolSection && currentSection != "" {
		sections[currentSection] = strings.TrimSpace(currentContent.String())
	}

	return sections
}

func (t *HelpTool) extractToolName(line string) (toolName string) {
	// Extract tool name from "### `tool_name`"
	start := strings.Index(line, "`")
	end := strings.LastIndex(line, "`")

	if start != -1 && end != -1 && start < end {
		toolName = line[start+1 : end]
	}

	return toolName
}

func (t *HelpTool) getToolNotFoundHelp(toolName string) (helpText string) {

	helpText = "Tool '" + toolName + "' not found.\n\n"
	helpText += "Available tools:\n"

	for _, tool := range AvailableTools {
		helpText += "- " + tool + "\n"
	}

	helpText += "\nCall help without parameters to see full documentation, or specify a tool name:\n"
	helpText += `{"tool": "help", "parameters": {"tool": "read_files"}}`

	return helpText
}

// generateToolHelp creates the same output as the help tool
// THIS IS CURRENTLY NOT USED, BUT WE MIGHT WANT TO USE LATER
func (t *HelpTool) generateToolHelp() (helpText string, err error) {
	var tools []mcputil.Tool
	var tool mcputil.Tool
	var options mcputil.ToolOptions
	var prop mcputil.Property
	var propOptions []mcputil.PropertyOption

	tools = mcputil.RegisteredTools()
	helpText = "# Scout-MCP Tools Documentation\n\n"

	for _, tool = range tools {
		options = tool.Options()
		helpText += fmt.Sprintf("## %s\n\n%s\n\n", options.Name, options.Description)

		if len(options.Properties) > 0 {
			helpText += "### Parameters\n\n"
			for _, prop = range options.Properties {
				propOptions = prop.PropertyOptions()
				helpText += fmt.Sprintf("- **%s** (%s)", prop.GetName(), prop.GetType())

				// Check if required
				for _, opt := range propOptions {
					p, ok := opt.(mcputil.RequiredProperty)
					if ok && p.Required {
						helpText += " *[required]*"
						break
					}
				}

				helpText += fmt.Sprintf(": %s\n", getPropertyDescription(prop))
			}
			helpText += "\n"
		}
	}

	return helpText, err
}
